'use client';

import {
  Badge,
  Box,
  Button,
  Card,
  Container,
  Group,
  Paper,
  SegmentedControl,
  SimpleGrid,
  Stack,
  Text,
  ThemeIcon,
  UnstyledButton,
} from '@mantine/core';
import {
  IconArrowRight,
  IconColorSwatch,
  IconFlame,
  IconMusic,
  IconUsers,
} from '@tabler/icons-react';
import { useEffect, useState } from 'react';
import { GroupLogo } from '@/features/portal/components/GroupLogo';
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
  selectedGroupId: string;
  songMode: boolean;
  onGroupChange: (groupId: string) => void;
  onSongModeChange: (songMode: boolean) => void;
  onStartQuiz: () => void;
}

export function PortalView({
  series,
  groups,
  members,
  songs,
  selectedGroupId,
  songMode,
  onGroupChange,
  onSongModeChange,
  onStartQuiz,
}: PortalViewProps) {
  const activeGroup = groups.find((g) => g.id === selectedGroupId) || groups[0];

  const [selectedSeriesId, setSelectedSeriesId] = useState(
    activeGroup?.series_id || series[0]?.id || '',
  );

  useEffect(() => {
    if (activeGroup?.series_id && activeGroup.series_id !== selectedSeriesId) {
      setSelectedSeriesId(activeGroup.series_id);
    }
  }, [activeGroup, selectedSeriesId]);

  const filteredGroups = groups.filter((g) => g.series_id === selectedSeriesId);

  const handleSeriesChange = (newSeriesId: string) => {
    setSelectedSeriesId(newSeriesId);
    const seriesGroups = groups.filter((g) => g.series_id === newSeriesId);
    if (seriesGroups[0]) {
      onGroupChange(seriesGroups[0].id);
    }
  };

  // Candidates: all active members with penlight colors or all songs for this group
  const groupMembers = members.filter(
    (m) =>
      m.group_id === activeGroup?.id &&
      m.status === 'active' &&
      m.penlight?.left_color_id &&
      m.penlight?.right_color_id,
  );
  const groupSongs = songs.filter((s) => s.group_id === activeGroup?.id);

  const candidateCount = songMode ? groupSongs.length : groupMembers.length;
  const isInsufficient = songMode
    ? groupSongs.length === 0
    : groupMembers.length === 0;

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
          <Group justify="flex-start" align="center">
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
                推しメンカラー & 楽曲ペンライトクイズ
              </Text>
            </Box>
          </Group>
        </Paper>

        {/* シリーズ選択 (坂道 / イコノイジョイ) */}
        {series.length > 1 && (
          <Stack gap="xs">
            <Text size="xs" fw={700} c="dimmed">
              シリーズ
            </Text>
            <SegmentedControl
              fullWidth
              size="sm"
              radius="md"
              value={selectedSeriesId}
              onChange={handleSeriesChange}
              data={series.map((s) => ({
                label: s.name,
                value: s.id,
              }))}
            />
          </Stack>
        )}

        {/* 配下グループ選択 (横並び配置) */}
        <Stack gap="xs">
          <Text size="xs" fw={700} c="dimmed">
            グループ
          </Text>

          <SimpleGrid
            cols={filteredGroups.length <= 3 ? filteredGroups.length : 3}
            spacing="xs"
          >
            {filteredGroups.map((g) => {
              const isSelected = g.id === activeGroup?.id;
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
                    transition: 'all 0.15s ease',
                    textAlign: 'center',
                    cursor: 'pointer',
                  }}
                >
                  <Stack align="center" justify="center" gap={0} h={48}>
                    <GroupLogo
                      slug={g.slug}
                      name={g.name}
                      height={44}
                      dimmed={!isSelected}
                    />
                  </Stack>
                </UnstyledButton>
              );
            })}
          </SimpleGrid>
        </Stack>

        {/* クイズ種別切替 (メンバー推しメンカラー vs 楽曲カラー) */}
        <Stack gap="xs">
          <Text size="xs" fw={700} c="dimmed">
            クイズ形式
          </Text>
          <SegmentedControl
            fullWidth
            size="md"
            radius="md"
            value={songMode ? 'song' : 'member'}
            onChange={(val) => onSongModeChange(val === 'song')}
            data={[
              {
                value: 'member',
                label: (
                  <Group gap={6} justify="center">
                    <IconUsers size={16} />
                    <span>メンバー</span>
                  </Group>
                ),
              },
              {
                value: 'song',
                label: (
                  <Group gap={6} justify="center">
                    <IconMusic size={16} />
                    <span>楽曲</span>
                  </Group>
                ),
              },
            ]}
          />
        </Stack>

        {/* 出題情報カード */}
        <Card withBorder radius="md" p="md">
          <Group justify="space-between" align="center">
            <Box>
              <Text size="sm" fw={700}>
                {activeGroup?.name}
              </Text>
            </Box>
            <Badge
              variant="light"
              color={isInsufficient ? 'red' : 'blue'}
              size="lg"
            >
              全 {candidateCount} {songMode ? '曲' : '名'}
            </Badge>
          </Group>
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
            {songMode
              ? `${activeGroup?.name} 楽曲クイズを開始`
              : `${activeGroup?.name} クイズを開始`}
          </Button>

          {isInsufficient && (
            <Text size="xs" c="red.6" ta="center" mt="xs">
              出題対象データがありません（別のグループを選択してください）
            </Text>
          )}
        </Box>
      </Stack>
    </Container>
  );
}
