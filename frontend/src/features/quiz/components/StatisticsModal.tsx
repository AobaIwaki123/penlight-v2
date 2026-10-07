'use client';

import {
  Badge,
  Box,
  Button,
  Card,
  Divider,
  Group,
  Loader,
  Modal,
  RingProgress,
  ScrollArea,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
} from '@mantine/core';
import {
  IconAlertCircle,
  IconClock,
  IconFlame,
  IconMusic,
  IconSparkles,
  IconTarget,
  IconTrophy,
  IconUser,
} from '@tabler/icons-react';
import { useEffect, useState } from 'react';
import { getApiBaseUrl } from '@/features/quiz/api/client';
import { getAnswerHistory } from '@/features/quiz/utils/idb';
import { calculateLocalStatistics } from '@/features/quiz/utils/statistics';
import type {
  Group as IdolGroup,
  Member,
  QuizStatisticsResponse,
  Song,
  TargetStat,
} from '@/types/generated';

interface StatisticsModalProps {
  opened: boolean;
  onClose: () => void;
  groups: IdolGroup[];
  members: Member[];
  songs: Song[];
  onStartWeakTargetQuiz?: (targets: TargetStat[]) => void;
}

export function StatisticsModal({
  opened,
  onClose,
  groups,
  members,
  songs,
  onStartWeakTargetQuiz,
}: StatisticsModalProps) {
  const [stats, setStats] = useState<QuizStatisticsResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isOfflineSource, setIsOfflineSource] = useState(false);

  useEffect(() => {
    if (!opened) return;

    let isMounted = true;
    setIsLoading(true);

    const loadStats = async () => {
      // 1. Try fetching from backend API if online
      if (typeof navigator !== 'undefined' && navigator.onLine) {
        try {
          const baseUrl = getApiBaseUrl();
          const res = await fetch(`${baseUrl}/api/v1/quiz/statistics?limit=5`);
          if (res.ok) {
            const data: QuizStatisticsResponse = await res.json();
            if (isMounted) {
              setStats(data);
              setIsOfflineSource(false);
              setIsLoading(false);
              return;
            }
          }
        } catch {
          // Fallback to local IndexedDB
        }
      }

      // 2. Fallback to local IndexedDB history (ADR-0007 Local-First)
      try {
        const historyItems = await getAnswerHistory();
        const localStats = calculateLocalStatistics(
          historyItems,
          groups,
          members,
          songs,
          5,
        );
        if (isMounted) {
          setStats(localStats);
          setIsOfflineSource(true);
          setIsLoading(false);
        }
      } catch (err) {
        console.warn('Failed to calculate local statistics:', err);
        if (isMounted) {
          setIsLoading(false);
        }
      }
    };

    loadStats();

    return () => {
      isMounted = false;
    };
  }, [opened, groups, members, songs]);

  const accuracyPercent = Math.round((stats?.accuracy_rate || 0) * 100);
  const weakTargets = stats?.weak_targets || [];

  return (
    <Modal
      opened={opened}
      onClose={onClose}
      title={
        <Group gap="xs">
          <IconTrophy size={20} color="#fcc419" />
          <Text fw={700}>回答履歴・統計分析</Text>
          {isOfflineSource && (
            <Badge size="xs" color="gray" variant="light">
              オフライン集計
            </Badge>
          )}
        </Group>
      }
      centered
      radius="md"
      size="md"
    >
      {isLoading ? (
        <Stack align="center" justify="center" py="xl" gap="sm">
          <Loader size="md" color="blue" />
          <Text size="sm" c="dimmed">
            統計データを集計中...
          </Text>
        </Stack>
      ) : !stats || stats.total_answers === 0 ? (
        <Stack align="center" justify="center" py="xl" gap="sm">
          <IconAlertCircle size={40} color="gray" />
          <Text size="sm" fw={600} c="dimmed">
            回答ログがまだありません
          </Text>
          <Text size="xs" c="dimmed" ta="center">
            クイズに回答すると、自動的に正答率や苦手カラーがここに集計されます。
          </Text>
        </Stack>
      ) : (
        <ScrollArea.Autosize maxHeight="70vh" offsetScrollbars>
          <Stack gap="md" py="xs">
            {/* サマリーカード: 全体正答率・回答数・平均時間 */}
            <Card withBorder radius="md" p="md">
              <Group justify="space-between" align="center">
                <Box>
                  <Text size="xs" c="dimmed" fw={600}>
                    累計回答数
                  </Text>
                  <Text size="xl" fw={800} c="blue.6">
                    {stats.total_answers}{' '}
                    <Text span size="sm" fw={500} c="dimmed">
                      問
                    </Text>
                  </Text>

                  <Group gap={6} mt={6} align="center">
                    <IconClock size={15} color="gray" />
                    <Text size="xs" c="dimmed">
                      平均速度:{' '}
                      <Text span fw={600}>
                        {(stats.average_response_time_ms / 1000).toFixed(1)}
                      </Text>{' '}
                      秒
                    </Text>
                  </Group>
                </Box>

                <RingProgress
                  size={90}
                  thickness={9}
                  roundCaps
                  sections={[
                    {
                      value: accuracyPercent,
                      color:
                        accuracyPercent >= 80
                          ? 'teal'
                          : accuracyPercent >= 50
                            ? 'blue'
                            : 'orange',
                    },
                  ]}
                  label={
                    <Text ta="center" size="sm" fw={800}>
                      {accuracyPercent}%
                    </Text>
                  }
                />
              </Group>
            </Card>

            {/* グループ別正答率 */}
            {stats.groups && stats.groups.length > 0 && (
              <Box>
                <Text size="xs" fw={700} c="dimmed" mb={6}>
                  グループ別成績
                </Text>
                <SimpleGrid cols={2} spacing="xs">
                  {stats.groups.map((g) => {
                    const groupAcc = Math.round(g.accuracy_rate * 100);
                    return (
                      <Card
                        key={g.group_id}
                        withBorder
                        radius="md"
                        p="xs"
                        bg="var(--mantine-color-default-hover)"
                      >
                        <Text size="xs" fw={700} truncate>
                          {g.group_name}
                        </Text>
                        <Group justify="space-between" mt={4}>
                          <Text size="xs" c="dimmed">
                            {g.total_answers}問
                          </Text>
                          <Badge
                            size="sm"
                            variant="light"
                            color={
                              groupAcc >= 80
                                ? 'teal'
                                : groupAcc >= 50
                                  ? 'blue'
                                  : 'orange'
                            }
                          >
                            {groupAcc}%
                          </Badge>
                        </Group>
                      </Card>
                    );
                  })}
                </SimpleGrid>
              </Box>
            )}

            <Divider />

            {/* 苦手克服対象ランキング */}
            <Box>
              <Group justify="space-between" align="center" mb={6}>
                <Group gap={4}>
                  <IconFlame size={16} color="#fa5252" />
                  <Text size="xs" fw={700} c="dimmed">
                    要復習・苦手対象 (ワースト)
                  </Text>
                </Group>
                {onStartWeakTargetQuiz && weakTargets.length > 0 && (
                  <Button
                    size="compact-xs"
                    variant="light"
                    color="red"
                    leftSection={<IconTarget size={14} />}
                    onClick={() => {
                      onStartWeakTargetQuiz(weakTargets);
                      onClose();
                    }}
                  >
                    苦手克服クイズ
                  </Button>
                )}
              </Group>

              {weakTargets.length === 0 ? (
                <Text size="xs" c="dimmed">
                  誤答データがありません（素晴らしい正答率です！）
                </Text>
              ) : (
                <Stack gap={6}>
                  {weakTargets.map((item, index) => {
                    const acc = Math.round(item.accuracy_rate * 100);
                    const isMember = item.target_type === 'member';

                    return (
                      <Card
                        key={`${item.target_type}-${item.target_id}`}
                        withBorder
                        radius="md"
                        p="xs"
                      >
                        <Group justify="space-between" wrap="nowrap">
                          <Group gap={8} wrap="nowrap">
                            <ThemeIcon
                              size="sm"
                              radius="xl"
                              variant="light"
                              color={index === 0 ? 'red' : 'gray'}
                            >
                              <Text size="xs" fw={700}>
                                {index + 1}
                              </Text>
                            </ThemeIcon>

                            <ThemeIcon
                              size="sm"
                              radius="md"
                              variant="subtle"
                              color={isMember ? 'blue' : 'violet'}
                            >
                              {isMember ? (
                                <IconUser size={14} />
                              ) : (
                                <IconMusic size={14} />
                              )}
                            </ThemeIcon>

                            <Box>
                              <Text size="xs" fw={700} truncate maxW={160}>
                                {item.name}
                              </Text>
                              <Text size="10px" c="dimmed">
                                {item.total_answers}問中 {item.correct_answers}
                                問正解
                              </Text>
                            </Box>
                          </Group>

                          <Badge
                            size="sm"
                            variant="filled"
                            color={
                              acc <= 30
                                ? 'red'
                                : acc <= 60
                                  ? 'orange'
                                  : 'yellow'
                            }
                          >
                            正答率 {acc}%
                          </Badge>
                        </Group>
                      </Card>
                    );
                  })}
                </Stack>
              )}
            </Box>

            {onStartWeakTargetQuiz && weakTargets.length > 0 && (
              <Button
                fullWidth
                mt="xs"
                color="red"
                variant="gradient"
                gradient={{ from: 'orange', to: 'red' }}
                leftSection={<IconSparkles size={16} />}
                onClick={() => {
                  onStartWeakTargetQuiz(weakTargets);
                  onClose();
                }}
              >
                苦手対象を重点復習する
              </Button>
            )}
          </Stack>
        </ScrollArea.Autosize>
      )}
    </Modal>
  );
}
