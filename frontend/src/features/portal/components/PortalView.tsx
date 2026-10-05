'use client';

import {
  Badge,
  Box,
  Button,
  Card,
  Container,
  Group,
  Paper,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
  UnstyledButton,
} from '@mantine/core';
import {
  IconArrowRight,
  IconColorSwatch,
  IconFilter,
  IconFlame,
  IconMusic,
  IconSparkles,
  IconUsers,
} from '@tabler/icons-react';
import type { QuizFilterCriteria } from '@/features/quiz/components/FilterModal';
import type {
  Color,
  Group as IdolGroup,
  Member,
  Series,
  Song,
} from '@/types/generated';

interface PortalViewProps {
  series: Series[];
  groups: IdolGroup[];
  members: Member[];
  songs: Song[];
  colors: Color[];
  criteria: QuizFilterCriteria;
  onGroupChange: (groupId: string) => void;
  onOpenFilter: () => void;
  onStartQuiz: () => void;
}

export function PortalView({
  series,
  groups,
  members,
  songs,
  criteria,
  onGroupChange,
  onOpenFilter,
  onStartQuiz,
}: PortalViewProps) {
  // Current active group and series (derived directly from group)
  const activeGroup =
    groups.find((g) => g.id === criteria.groupId) || groups[0];

  // Candidates count
  const matchingMembers = members.filter((m) => {
    if (m.group_id !== activeGroup?.id) return false;
    if (!criteria.includeGraduated && m.status !== 'active') return false;
    if (!m.penlight?.left_color_id || !m.penlight?.right_color_id) return false;
    return criteria.generations.includes(m.generation);
  });

  const matchingSongs = songs.filter((s) => s.group_id === activeGroup?.id);

  const candidateCount = criteria.songMode
    ? matchingSongs.length
    : matchingMembers.length;

  const isInsufficient = criteria.songMode
    ? matchingSongs.length === 0
    : matchingMembers.length < 4;

  return (
    <Container
      size="xs"
      p="md"
      style={{ minHeight: '100dvh', display: 'flex', flexDirection: 'column' }}
    >
      <Stack gap="lg" style={{ flexGrow: 1 }}>
        {/* ヘッダーブランド */}
        <Paper
          p="md"
          radius="lg"
          withBorder
          style={{
            background: `linear-gradient(135deg, ${activeGroup?.theme_color_hex || '#7CC7E8'}15, transparent)`,
            borderColor: `${activeGroup?.theme_color_hex || '#7CC7E8'}40`,
          }}
        >
          <Group justify="space-between" align="center">
            <Group gap="xs">
              <ThemeIcon
                size={38}
                radius="md"
                style={{
                  backgroundColor: activeGroup?.theme_color_hex || '#7CC7E8',
                  color: '#ffffff',
                }}
              >
                <IconColorSwatch size={22} />
              </ThemeIcon>
              <Box>
                <Text fw={800} size="lg" style={{ letterSpacing: '0.02em' }}>
                  ペンライトクイズ
                </Text>
                <Text size="xs" c="dimmed">
                  アイドル推しメンカラー & 楽曲ペンライト当て
                </Text>
              </Box>
            </Group>
            <Badge
              variant="light"
              color="indigo"
              size="sm"
              leftSection={<IconSparkles size={12} />}
            >
              v2.0
            </Badge>
          </Group>
        </Paper>

        {/* グループ選択カード (1タップで直接選ぶミニマル設計) */}
        <Stack gap="xs">
          <Group justify="space-between" align="center">
            <Text size="xs" fw={700} c="dimmed">
              グループを選択
            </Text>
            <Text size="xs" c="dimmed">
              {groups.length} グループ
            </Text>
          </Group>

          <SimpleGrid cols={3} spacing="xs">
            {groups.map((g) => {
              const isSelected = g.id === activeGroup?.id;
              const seriesObj = series.find((s) => s.id === g.series_id);
              return (
                <UnstyledButton
                  key={g.id}
                  onClick={() => onGroupChange(g.id)}
                  style={{
                    borderRadius: 12,
                    padding: '12px 6px',
                    border: isSelected
                      ? `2px solid ${g.theme_color_hex}`
                      : '1px solid var(--mantine-color-default-border)',
                    backgroundColor: isSelected
                      ? `${g.theme_color_hex}15`
                      : 'var(--mantine-color-body)',
                    transition: 'all 0.2s ease',
                    textAlign: 'center',
                    cursor: 'pointer',
                  }}
                >
                  <Stack align="center" gap={4}>
                    <Box
                      style={{
                        width: 14,
                        height: 14,
                        borderRadius: '50%',
                        backgroundColor: g.theme_color_hex,
                        boxShadow: isSelected
                          ? `0 0 8px ${g.theme_color_hex}`
                          : 'none',
                      }}
                    />
                    <Text
                      size="sm"
                      fw={isSelected ? 700 : 500}
                      c={isSelected ? 'var(--mantine-color-text)' : 'dimmed'}
                      style={{
                        whiteSpace: 'nowrap',
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                        maxWidth: '100%',
                      }}
                    >
                      {g.name}
                    </Text>
                    {seriesObj && (
                      <Text size="9px" c="dimmed" lh={1}>
                        {seriesObj.name}
                      </Text>
                    )}
                  </Stack>
                </UnstyledButton>
              );
            })}
          </SimpleGrid>
        </Stack>

        {/* 現在のクイズ出題設定サマリー & 変更ボタン (ADR-0029) */}
        <Card withBorder radius="md" p="md">
          <Stack gap="sm">
            <Group justify="space-between" align="center">
              <Group gap="xs">
                {criteria.songMode ? (
                  <ThemeIcon color="violet" variant="light" size="sm">
                    <IconMusic size={14} />
                  </ThemeIcon>
                ) : (
                  <ThemeIcon color="blue" variant="light" size="sm">
                    <IconUsers size={14} />
                  </ThemeIcon>
                )}
                <Text size="sm" fw={700}>
                  {criteria.songMode
                    ? '楽曲カラークイズ'
                    : 'メンバー推しメンカラー'}
                </Text>
              </Group>
              <Button
                variant="subtle"
                size="compact-xs"
                color="blue"
                leftSection={<IconFilter size={14} />}
                onClick={onOpenFilter}
              >
                条件を変更
              </Button>
            </Group>

            <Group gap="xs">
              <Badge variant="outline" color="gray" size="sm">
                {activeGroup?.name}
              </Badge>
              {!criteria.songMode ? (
                <>
                  <Badge variant="outline" color="blue" size="sm">
                    {criteria.generations.length === 0
                      ? '全期生'
                      : `${criteria.generations.join(', ')} 期生`}
                  </Badge>
                  {criteria.includeGraduated && (
                    <Badge variant="outline" color="orange" size="sm">
                      卒業生含む
                    </Badge>
                  )}
                </>
              ) : (
                <Badge variant="outline" color="violet" size="sm">
                  代表曲
                </Badge>
              )}
              <Badge
                variant="light"
                color={isInsufficient ? 'red' : 'green'}
                size="sm"
              >
                出題対象: {candidateCount} {criteria.songMode ? '曲' : '名'}
              </Badge>
            </Group>
          </Stack>
        </Card>

        {/* クイズ開始 CTA ボタン */}
        <Box mt="auto" pt="md">
          <Button
            fullWidth
            size="lg"
            radius="md"
            style={{
              backgroundColor: isInsufficient
                ? undefined
                : activeGroup?.theme_color_hex || '#228be6',
            }}
            rightSection={<IconArrowRight size={20} />}
            leftSection={<IconFlame size={20} />}
            onClick={onStartQuiz}
            disabled={isInsufficient}
          >
            {criteria.songMode
              ? `${activeGroup?.name} 楽曲クイズを開始`
              : `${activeGroup?.name} 推しメンクイズを開始`}
          </Button>

          {isInsufficient && (
            <Text size="xs" c="red.6" ta="center" mt="xs">
              出題対象が不足しています（「条件を変更」から期生またはグループを選択してください）
            </Text>
          )}
        </Box>
      </Stack>
    </Container>
  );
}
