'use client';

import { Badge, Box, Group, Image, Paper, Stack, Text } from '@mantine/core';
import { getImageUrl } from '@/features/quiz/api/client';
import { PenlightStick } from '@/features/quiz/components/PenlightStick';
import type { TargetLayoutProps } from '@/features/quiz/types';

export function LayoutClassic({
  target,
  costumeTitle,
  selectedLeftColor,
  selectedRightColor,
}: TargetLayoutProps) {
  const primaryImg = target.images?.[0];
  const imageSrc =
    getImageUrl(primaryImg?.image_key) ||
    'https://placehold.co/320x320/7cc7e8/ffffff?text=Penlight+Quiz';

  return (
    <Stack gap="sm" align="center" style={{ width: '100%' }}>
      {/* メンバー名と期生 */}
      <Group gap="xs" justify="center">
        <Text size="xl" fw={700} c="orange.7">
          {target.family_name} {target.given_name}
        </Text>
        <Badge size="sm" variant="light" color="blue">
          {target.generation}期生
        </Badge>
      </Group>

      {/* メンバー写真 (旧版サイズ感: 角丸スクエア) */}
      <Paper
        shadow="xs"
        radius="md"
        withBorder
        style={{
          width: 240,
          height: 240,
          overflow: 'hidden',
          backgroundColor: '#fafafa',
        }}
      >
        <Image
          src={imageSrc}
          alt={`${target.family_name} ${target.given_name}`}
          w="100%"
          h="100%"
          fit="cover"
        />
      </Paper>

      {/* 衣装名 */}
      {costumeTitle && (
        <Text size="xs" c="dimmed">
          衣装: {costumeTitle}
        </Text>
      )}

      {/* 左右ペンライト (旧版と同じ2本並び) */}
      <Box pt={4}>
        <Group gap={40} justify="center">
          <PenlightStick
            color={selectedLeftColor}
            label="左手"
            height={90}
            width={40}
          />
          <PenlightStick
            color={selectedRightColor}
            label="右手"
            height={90}
            width={40}
          />
        </Group>
      </Box>
    </Stack>
  );
}
