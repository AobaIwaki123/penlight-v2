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
  activeHand = 'left',
}: LayoutOverlayProps) {
  const primaryImg = target.images?.[0];
  const imageSrc =
    getImageUrl(primaryImg?.image_key) ||
    'https://placehold.co/400x500/7cc7e8/ffffff?text=Penlight+Quiz';

  const isLeftActive = activeHand === 'left';
  const isRightActive = activeHand === 'right';

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

        {/* 左上隅: 左ペンライト (選択中なら青く発光オーラ！) */}
        <Box
          role="button"
          tabIndex={0}
          onClick={() => onOpenInput?.('left')}
          title="タップして左手の色を選択"
          style={{
            position: 'absolute',
            top: 14,
            left: 14,
            zIndex: 10,
            cursor: 'pointer',
            backgroundColor: isLeftActive
              ? 'rgba(0, 0, 0, 0.45)'
              : 'rgba(0, 0, 0, 0.15)',
            backdropFilter: 'blur(10px)',
            borderRadius: 999,
            padding: '6px 6px',
            border: isLeftActive
              ? '2px solid #339af0'
              : '1.5px solid rgba(255, 255, 255, 0.2)',
            boxShadow: isLeftActive
              ? '0 0 20px rgba(51, 154, 240, 0.8), inset 0 0 10px rgba(51, 154, 240, 0.3)'
              : '0 4px 12px rgba(0,0,0,0.15)',
            transform: isLeftActive ? 'scale(1.08)' : 'scale(1)',
            transition: 'all 0.25s cubic-bezier(0.34, 1.56, 0.64, 1)',
          }}
        >
          <PenlightStick
            color={selectedLeftColor}
            height={56}
            width={22}
            textColor="#ffffff"
          />
        </Box>

        {/* 右上隅: 右ペンライト (選択中ならオレンジに発光オーラ！) */}
        <Box
          role="button"
          tabIndex={0}
          onClick={() => onOpenInput?.('right')}
          title="タップして右手の色を選択"
          style={{
            position: 'absolute',
            top: 14,
            right: 14,
            zIndex: 10,
            cursor: 'pointer',
            backgroundColor: isRightActive
              ? 'rgba(0, 0, 0, 0.45)'
              : 'rgba(0, 0, 0, 0.15)',
            backdropFilter: 'blur(10px)',
            borderRadius: 999,
            padding: '6px 6px',
            border: isRightActive
              ? '2px solid #ff922b'
              : '1.5px solid rgba(255, 255, 255, 0.2)',
            boxShadow: isRightActive
              ? '0 0 20px rgba(255, 146, 43, 0.8), inset 0 0 10px rgba(255, 146, 43, 0.3)'
              : '0 4px 12px rgba(0,0,0,0.15)',
            transform: isRightActive ? 'scale(1.08)' : 'scale(1)',
            transition: 'all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1)',
          }}
        >
          <PenlightStick
            color={selectedRightColor}
            height={58}
            width={24}
            textColor="#ffffff"
          />
        </Box>

        {/* 自然な下部フェードグラデーション */}
        <Box
          style={{
            position: 'absolute',
            bottom: 0,
            left: 0,
            right: 0,
            height: '45%',
            background:
              'linear-gradient(to top, rgba(0,0,0,0.85) 0%, rgba(0,0,0,0.4) 60%, transparent 100%)',
            pointerEvents: 'none',
          }}
        />

        {/* 下部オーバーレイ情報 (メンバー名・期生・衣装) */}
        <Box
          style={{
            position: 'absolute',
            bottom: 16,
            left: 20,
            right: 20,
            color: '#fff',
            textShadow: '0 2px 4px rgba(0,0,0,0.6)',
          }}
        >
          <Group gap={8} align="center" mb={4}>
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
      </Paper>
    </Box>
  );
}
