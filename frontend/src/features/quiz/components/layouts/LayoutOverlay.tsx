'use client';

import { Badge, Box, Group, Image, Paper, Text } from '@mantine/core';
import type { ReactNode } from 'react';
import { getImageUrl } from '@/features/quiz/api/client';
import { PenlightStick } from '@/features/quiz/components/PenlightStick';
import type { TargetLayoutProps } from '@/features/quiz/types';

interface LayoutOverlayProps extends TargetLayoutProps {
  isFullscreen?: boolean;
  /** 写真の最下部に重ねて表示する領域 (解答フィードバックバーなど) */
  footer?: ReactNode;
  /** 編集画面で同じ写真レイアウトに重ねるメンバー情報と写真操作 */
  memberDetails?: ReactNode;
  photoControl?: ReactNode;
}

export function LayoutOverlay({
  target,
  costumeTitle,
  selectedLeftColor,
  selectedRightColor,
  onOpenInput,
  isFullscreen,
  footer,
  memberDetails,
  photoControl,
}: LayoutOverlayProps) {
  const primaryImg = target.images?.[0];
  const imageFallback = `data:image/svg+xml,${encodeURIComponent(
    '<svg xmlns="http://www.w3.org/2000/svg" width="400" height="600"><rect width="100%" height="100%" fill="#343a40"/><text x="50%" y="42%" text-anchor="middle" fill="white" font-size="18">画像を読み込めません</text></svg>',
  )}`;
  const imageSrc = getImageUrl(primaryImg?.image_key) || imageFallback;

  return (
    <Box
      style={{
        width: '100%',
        maxWidth: 440,
        alignSelf: 'stretch',
        minHeight: isFullscreen ? 0 : undefined,
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* 写真カード (自然な下部グラデーション + 角丸カード) */}
      <Paper
        radius="lg"
        shadow="md"
        style={{
          position: 'relative',
          width: '100%',
          height: isFullscreen ? undefined : 330,
          flexGrow: isFullscreen ? 1 : undefined,
          minHeight: isFullscreen ? 0 : undefined,
          overflow: 'hidden',
          backgroundColor: '#000',
        }}
      >
        <Image
          src={imageSrc}
          fallbackSrc={imageFallback}
          alt={`${target.family_name} ${target.given_name}`}
          fit="cover"
          style={{
            position: 'absolute',
            inset: 0,
            width: '100%',
            height: '100%',
          }}
        />

        {photoControl && (
          <Box style={{ position: 'absolute', top: 12, right: 12 }}>
            {photoControl}
          </Box>
        )}

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
            // フッター (フィードバックバー) 表示中は、その分だけ情報を上へ持ち上げる
            bottom: footer ? 112 : 12,
            transition: 'bottom 0.25s ease',
            left: 16,
            right: 16,
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'flex-end',
          }}
        >
          {/* 左側: メンバー名と衣装 */}
          <Box
            style={{
              color: '#fff',
              textShadow: '0 2px 4px rgba(0,0,0,0.6)',
              minWidth: 0,
              maxWidth: memberDetails ? 'calc(100% - 116px)' : undefined,
            }}
          >
            {memberDetails || (
              <>
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
              </>
            )}
          </Box>

          {/* 右側: 写真に重なる2本の光るペンライト (タップでカラー選択モーダル展開) */}
          <Group
            gap={16}
            align="flex-end"
            wrap="nowrap"
            style={{ flexShrink: 0 }}
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

        {/* 写真最下部に重ねるフッター (解答フィードバックバーなど) */}
        {footer && (
          <Box
            style={{
              position: 'absolute',
              left: 10,
              right: 10,
              bottom: 10,
              zIndex: 20,
              display: 'flex',
              justifyContent: 'center',
            }}
          >
            {footer}
          </Box>
        )}
      </Paper>
    </Box>
  );
}
