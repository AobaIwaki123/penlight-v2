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
import type {
  Color,
  Group as IdolGroup,
  Member,
  Song,
} from '@/types/generated';

interface PortalViewProps {
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
                推しメンカラー & 楽曲ペンライト当て
              </Text>
            </Box>
          </Group>
        </Paper>

        {/* グループ選択 (系列表示なし・横並びグリッドで直接選択) */}
        <Stack gap="xs">
          <Text size="xs" fw={700} c="dimmed">
            グループを選択
          </Text>

          <SimpleGrid cols={3} spacing="xs">
            {groups.map((g) => {
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
                  <Stack align="center" gap={6}>
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
                    <span>推しメンカラー</span>
                  </Group>
                ),
              },
              {
                value: 'song',
                label: (
                  <Group gap={6} justify="center">
                    <IconMusic size={16} />
                    <span>楽曲カラー</span>
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
              <Text size="xs" c="dimmed">
                {songMode
                  ? '代表曲の指定カラークイズ'
                  : '現役メンバーの推しメンカラークイズ'}
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
