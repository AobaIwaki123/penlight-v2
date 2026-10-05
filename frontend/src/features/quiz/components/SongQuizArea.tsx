'use client';

import {
  Badge,
  Box,
  Card,
  Group,
  Stack,
  Text,
  ThemeIcon,
  UnstyledButton,
} from '@mantine/core';
import { IconMusic, IconSparkles } from '@tabler/icons-react';
import type { ReactNode } from 'react';
import { PenlightStick } from '@/features/quiz/components/PenlightStick';
import type { Color, Group as IdolGroup, Song } from '@/types/generated';

interface SongQuizAreaProps {
  song: Song;
  group?: IdolGroup;
  selectedColor1?: Color;
  selectedColor2?: Color;
  isCorrect?: boolean;
  onOpenInput?: (slot: 1 | 2) => void;
  footer?: ReactNode;
}

export function SongQuizArea({
  song,
  group,
  selectedColor1,
  selectedColor2,
  onOpenInput,
  footer,
}: SongQuizAreaProps) {
  const isTwoColors = Boolean(song.color2_id);

  return (
    <Card
      withBorder
      radius="lg"
      p="lg"
      style={{
        width: '100%',
        maxWidth: 420,
        backgroundColor: 'var(--mantine-color-body)',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'space-between',
        minHeight: 280,
        position: 'relative',
        boxShadow: '0 4px 20px rgba(0,0,0,0.06)',
      }}
    >
      {/* 上部: 楽曲メタ情報 */}
      <Stack align="center" gap={6} style={{ width: '100%' }}>
        <Group gap="xs" justify="center">
          <Badge
            variant="filled"
            size="sm"
            style={{ backgroundColor: group?.theme_color_hex || '#7CC7E8' }}
          >
            {group?.name || '公式楽曲'}
          </Badge>
          <Badge
            variant="light"
            color={isTwoColors ? 'violet' : 'cyan'}
            size="sm"
          >
            {isTwoColors ? '2色指定' : '1色指定'}
          </Badge>
        </Group>

        {/* 読み仮名 (Kana) */}
        {song.kana && (
          <Text size="xs" c="dimmed" fw={500} ta="center">
            {song.kana}
          </Text>
        )}

        {/* 楽曲タイトル */}
        <Group gap={6} justify="center" align="center">
          <ThemeIcon size={24} radius="xl" color="violet" variant="light">
            <IconMusic size={14} />
          </ThemeIcon>
          <Text
            fw={800}
            size="xl"
            ta="center"
            style={{
              lineHeight: 1.25,
              wordBreak: 'break-word',
            }}
          >
            {song.title}
          </Text>
        </Group>
      </Stack>

      {/* 中央: ペンライト表示 (1本または2本) */}
      <Box
        py="md"
        style={{ width: '100%', display: 'flex', justifyContent: 'center' }}
      >
        {isTwoColors ? (
          <Group gap="xl" justify="center" align="flex-end">
            <UnstyledButton
              onClick={() => onOpenInput?.(1)}
              style={{ cursor: onOpenInput ? 'pointer' : 'default' }}
            >
              <PenlightStick
                color={selectedColor1}
                label="1色目"
                height={100}
                width={42}
              />
            </UnstyledButton>
            <UnstyledButton
              onClick={() => onOpenInput?.(2)}
              style={{ cursor: onOpenInput ? 'pointer' : 'default' }}
            >
              <PenlightStick
                color={selectedColor2}
                label="2色目"
                height={100}
                width={42}
              />
            </UnstyledButton>
          </Group>
        ) : (
          <UnstyledButton
            onClick={() => onOpenInput?.(1)}
            style={{ cursor: onOpenInput ? 'pointer' : 'default' }}
          >
            <PenlightStick
              color={selectedColor1}
              label="指定カラー"
              height={110}
              width={46}
            />
          </UnstyledButton>
        )}
      </Box>

      {/* ガイドテキスト */}
      {!selectedColor1 && (
        <Group gap={4} justify="center">
          <IconSparkles size={14} color="#fab005" />
          <Text size="xs" c="dimmed" fw={600}>
            {isTwoColors
              ? 'ペンライト色を2色選んでください'
              : 'ペンライト色を1色選んでください'}
          </Text>
        </Group>
      )}

      {/* フッター (インラインフィードバック等) */}
      {footer && <Box style={{ width: '100%' }}>{footer}</Box>}
    </Card>
  );
}
