'use client';

import { Badge, Box, Group, Image, Paper, Stack, Text } from '@mantine/core';
import { getImageUrl } from '@/features/quiz/api/client';
import { PenlightStick } from '@/features/quiz/components/PenlightStick';
import type { TargetLayoutProps } from '@/features/quiz/types';

export function LayoutCompact({
  target,
  costumeTitle,
  selectedLeftColor,
  selectedRightColor,
}: TargetLayoutProps) {
  const primaryImg = target.images?.[0];
  const imageSrc =
    getImageUrl(primaryImg?.image_key) ||
    'https://placehold.co/240x240/7cc7e8/ffffff?text=Penlight';

  return (
    <Paper
      withBorder
      p="xs"
      radius="md"
      style={{ width: '100%', maxWidth: 400 }}
    >
      <Group gap="md" align="center" wrap="nowrap">
        {/* コンパクト写真 */}
        <Paper
          radius="sm"
          style={{
            width: 110,
            height: 110,
            flexShrink: 0,
            overflow: 'hidden',
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

        {/* メンバー情報 ＋ ペンライト */}
        <Stack gap={6} style={{ flexGrow: 1 }}>
          <Group gap={6}>
            <Text size="md" fw={700} c="orange.7">
              {target.family_name} {target.given_name}
            </Text>
            <Badge size="xs" variant="light">
              {target.generation}期
            </Badge>
          </Group>
          {costumeTitle && (
            <Text size="11px" c="dimmed" lineClamp={1}>
              {costumeTitle}
            </Text>
          )}
          <Box pt={4}>
            <Group gap={20}>
              <PenlightStick
                color={selectedLeftColor}
                label="左"
                height={55}
                width={24}
              />
              <PenlightStick
                color={selectedRightColor}
                label="右"
                height={55}
                width={24}
              />
            </Group>
          </Box>
        </Stack>
      </Group>
    </Paper>
  );
}
