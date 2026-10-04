'use client';

import { Badge, Box, Group, Image, Paper, Text } from '@mantine/core';
import { getImageUrl } from '@/features/quiz/api/client';
import { PenlightStick } from '@/features/quiz/components/PenlightStick';
import type { TargetLayoutProps } from '@/features/quiz/types';

export function LayoutOverlay({
  target,
  costumeTitle,
  selectedLeftColor,
  selectedRightColor,
}: TargetLayoutProps) {
  const primaryImg = target.images?.[0];
  const imageSrc =
    getImageUrl(primaryImg?.image_key) ||
    'https://placehold.co/400x500/7cc7e8/ffffff?text=Penlight+Quiz';

  return (
    <Box style={{ width: '100%', maxWidth: 420 }}>
      {/* 写真カード (大画面 + 自然な下部グラデーション) */}
      <Paper
        radius="lg"
        shadow="md"
        style={{
          position: 'relative',
          width: '100%',
          height: 330,
          overflow: 'hidden',
          backgroundColor: '#000',
        }}
      >
        <Image
          src={imageSrc}
          alt={`${target.family_name} ${target.given_name}`}
          w="100%"
          h="100%"
          fit="cover"
        />

        {/* 自然な下部フェードグラデーション */}
        <Box
          style={{
            position: 'absolute',
            bottom: 0,
            left: 0,
            right: 0,
            height: '55%',
            background:
              'linear-gradient(to top, rgba(0,0,0,0.9) 0%, rgba(0,0,0,0.5) 60%, transparent 100%)',
            pointerEvents: 'none',
          }}
        />

        {/* オーバーレイ情報 (メンバー名・衣装・光るペンライト) */}
        <Box
          style={{
            position: 'absolute',
            bottom: 12,
            left: 16,
            right: 16,
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'flex-end',
          }}
        >
          {/* 左側: メンバー名と衣装 */}
          <Box
            style={{ color: '#fff', textShadow: '0 2px 4px rgba(0,0,0,0.6)' }}
          >
            <Group gap={8} align="center" mb={2}>
              <Text size="xl" fw={800} c="white">
                {target.family_name} {target.given_name}
              </Text>
              <Badge size="sm" color="orange" variant="filled">
                {target.generation}期生
              </Badge>
            </Group>
            {costumeTitle && (
              <Text size="xs" c="gray.3">
                {costumeTitle}
              </Text>
            )}
          </Box>

          {/* 右側: 写真に重なる2本の光るペンライト */}
          <Group gap={16} align="flex-end">
            <PenlightStick
              color={selectedLeftColor}
              label="左"
              height={70}
              width={28}
              textColor="#ffffff"
            />
            <PenlightStick
              color={selectedRightColor}
              label="右"
              height={70}
              width={28}
              textColor="#ffffff"
            />
          </Group>
        </Box>
      </Paper>
    </Box>
  );
}
