'use client';

import { Badge, Box, Group, Image, Paper, Text } from '@mantine/core';
import { getImageUrl } from '@/features/quiz/api/client';
import { PenlightStick } from '@/features/quiz/components/PenlightStick';
import type { TargetLayoutProps } from '@/features/quiz/types';

interface LayoutOverlayProps extends TargetLayoutProps {
  isFullscreen?: boolean;
}

export function LayoutOverlay({
  target,
  costumeTitle,
  selectedLeftColor,
  selectedRightColor,
  onOpenInput,
  isFullscreen,
  isInputActive,
}: LayoutOverlayProps) {
  const primaryImg = target.images?.[0];
  const imageSrc =
    getImageUrl(primaryImg?.image_key) ||
    'https://placehold.co/400x500/7cc7e8/ffffff?text=Penlight+Quiz';

  return (
    <Box
      style={{
        width: '100%',
        maxWidth: 440,
        height: isFullscreen ? 'calc(100dvh - 180px)' : undefined,
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* 写真カード (全画面または大画面 + 自然な下部グラデーション) */}
      <Paper
        radius="lg"
        shadow="md"
        style={{
          position: 'relative',
          width: '100%',
          height: isFullscreen ? '100%' : 330,
          flexGrow: isFullscreen ? 1 : undefined,
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

          {/* 右側: 写真に重なる2本の光るペンライト (タップでカラー選択モーダル展開) */}
          <Group
            gap={16}
            align="flex-end"
            style={{
              opacity: isInputActive ? 0.2 : 1,
              transition: 'opacity 0.2s ease',
            }}
          >
            <Box
              role="button"
              tabIndex={0}
              onClick={() => onOpenInput?.('left')}
              title="タップして左手の色を選択"
              style={{
                cursor: 'pointer',
                transition: 'transform 0.15s ease',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.transform = 'scale(1.08)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = 'scale(1)';
              }}
            >
              <PenlightStick
                color={selectedLeftColor}
                label="左 (タップ)"
                height={isFullscreen ? 84 : 70}
                width={isFullscreen ? 32 : 28}
                textColor="#ffffff"
              />
            </Box>

            <Box
              role="button"
              tabIndex={0}
              onClick={() => onOpenInput?.('right')}
              title="タップして右手の色を選択"
              style={{
                cursor: 'pointer',
                transition: 'transform 0.15s ease',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.transform = 'scale(1.08)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = 'scale(1)';
              }}
            >
              <PenlightStick
                color={selectedRightColor}
                label="右 (タップ)"
                height={isFullscreen ? 84 : 70}
                width={isFullscreen ? 32 : 28}
                textColor="#ffffff"
              />
            </Box>
          </Group>
        </Box>
      </Paper>
    </Box>
  );
}
