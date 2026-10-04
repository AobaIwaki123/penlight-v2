'use client';

import { Box, Group, Modal, Paper, Text } from '@mantine/core';
import { useEffect, useState } from 'react';
import { PenlightStick } from '@/features/quiz/components/PenlightStick';
import type { Color } from '@/types/generated';

interface DonutRingModalProps {
  opened: boolean;
  onClose: () => void;
  colors: Color[];
  selectedLeftColor?: Color;
  selectedRightColor?: Color;
  onAnswer: (input: { leftColorId: string; rightColorId: string }) => void;
  onColorSelect: (hand: 'left' | 'right', color: Color) => void;
  disabled: boolean;
  initialHand?: 'left' | 'right';
}

export function DonutRingModal({
  opened,
  onClose,
  colors,
  selectedLeftColor,
  selectedRightColor,
  onAnswer,
  onColorSelect,
  disabled,
  initialHand = 'left',
}: DonutRingModalProps) {
  const [currentHand, setCurrentHand] = useState<'left' | 'right'>(initialHand);
  const [hoveredColor, setHoveredColor] = useState<Color | null>(null);

  // 開くたびに指定された手へ切り替える
  useEffect(() => {
    if (opened) setCurrentHand(initialHand);
  }, [opened, initialHand]);

  const radius = 120; // 円環の半径 (px)
  const buttonSize = 38; // 各カラージュエルの直径 (px)

  const handleColorClick = (color: Color) => {
    if (disabled) return;

    if (currentHand === 'left') {
      onColorSelect('left', color);
      if (selectedRightColor) {
        // Both colors selected -> submit and close
        onAnswer({
          leftColorId: color.id,
          rightColorId: selectedRightColor.id,
        });
        onClose();
      } else {
        // Auto-switch focus visually to right hand
        setCurrentHand('right');
      }
    } else {
      onColorSelect('right', color);
      if (selectedLeftColor) {
        // Both colors selected -> submit and close
        onAnswer({ leftColorId: selectedLeftColor.id, rightColorId: color.id });
        onClose();
      } else {
        // Auto-switch focus visually to left hand
        setCurrentHand('left');
      }
    }
  };

  const activeColor =
    currentHand === 'left' ? selectedLeftColor : selectedRightColor;
  const displayColor = hoveredColor || activeColor;

  return (
    <Modal
      opened={opened}
      onClose={onClose}
      centered
      withCloseButton={false}
      size="auto"
      padding={0}
      overlayProps={{
        backgroundOpacity: 0.35,
        blur: 0, // 推しの写真をぼかさず、くっきり透かす
      }}
      styles={{
        content: {
          backgroundColor: 'transparent',
          boxShadow: 'none',
          border: 'none',
          overflow: 'visible',
          // 画面中央から下方（胸元・お腹位置）へしっかりシフトして、顔（目・鼻・口）を100%完全にクリアにする
          transform: 'translateY(135px)',
        },
        body: {
          padding: 0,
          backgroundColor: 'transparent',
          overflow: 'visible',
        },
      }}
    >
      <Box
        style={{
          position: 'relative',
          width: radius * 2 + buttonSize + 32,
          height: radius * 2 + buttonSize + 32,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        {/* 背景の光彩ガイドリング */}
        <Box
          style={{
            position: 'absolute',
            width: radius * 2,
            height: radius * 2,
            borderRadius: '50%',
            border: '1.5px dashed rgba(255,255,255,0.4)',
            boxShadow: '0 0 20px rgba(0,0,0,0.3)',
            pointerEvents: 'none',
          }}
        />

        {/* ドーナツ中央の空洞: 2本のミニペンライトで左右を視覚化！ */}
        <Paper
          radius="50%"
          shadow="lg"
          style={{
            width: 154,
            height: 154,
            borderRadius: '50%',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            textAlign: 'center',
            padding: '6px 8px',
            zIndex: 2,
            backgroundColor: 'rgba(0, 0, 0, 0.72)',
            backdropFilter: 'blur(8px)',
            border: '2px solid rgba(255,255,255,0.25)',
            transition: 'all 0.2s ease',
          }}
        >
          {/* 視覚的ペンライト2本並び (横並びを厳守 wrap="nowrap") */}
          <Group gap={12} justify="center" align="center" wrap="nowrap">
            {/* 左ペンライト */}
            <Box
              role="button"
              tabIndex={0}
              onClick={() => setCurrentHand('left')}
              style={{
                cursor: 'pointer',
                padding: '4px 6px',
                borderRadius: 8,
                backgroundColor:
                  currentHand === 'left'
                    ? 'rgba(34, 139, 230, 0.45)'
                    : 'transparent',
                border:
                  currentHand === 'left'
                    ? '2px solid #339af0'
                    : '1px solid transparent',
                transform:
                  currentHand === 'left' ? 'scale(1.1)' : 'scale(0.92)',
                transition: 'all 0.15s ease',
              }}
            >
              <PenlightStick
                color={selectedLeftColor}
                label="左"
                height={46}
                width={20}
                textColor="#ffffff"
              />
            </Box>

            {/* 右ペンライト */}
            <Box
              role="button"
              tabIndex={0}
              onClick={() => setCurrentHand('right')}
              style={{
                cursor: 'pointer',
                padding: '4px 6px',
                borderRadius: 8,
                backgroundColor:
                  currentHand === 'right'
                    ? 'rgba(253, 126, 20, 0.45)'
                    : 'transparent',
                border:
                  currentHand === 'right'
                    ? '2px solid #ff922b'
                    : '1px solid transparent',
                transform:
                  currentHand === 'right' ? 'scale(1.1)' : 'scale(0.92)',
                transition: 'all 0.15s ease',
              }}
            >
              <PenlightStick
                color={selectedRightColor}
                label="右"
                height={46}
                width={20}
                textColor="#ffffff"
              />
            </Box>
          </Group>

          {/* 選択中/ホバー中の色名 (1行固定・文字数に合わせた動的サイズ) */}
          <Text
            fw={700}
            c={displayColor ? 'white' : 'gray.4'}
            mt={4}
            style={{
              fontSize: (displayColor?.name.length || 0) >= 7 ? '10px' : '11px',
              textShadow: '0 1px 3px rgba(0,0,0,0.8)',
              maxWidth: 124,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
              letterSpacing:
                (displayColor?.name.length || 0) >= 7 ? '-0.03em' : 'normal',
            }}
          >
            {displayColor ? displayColor.name : '色をタップ'}
          </Text>
        </Paper>

        {/* 円周上に浮かぶ 15 個のカラージュエル */}
        {colors.map((color, index) => {
          const total = colors.length;
          const angle = (2 * Math.PI * index) / total - Math.PI / 2;
          const x = radius * Math.cos(angle);
          const y = radius * Math.sin(angle);

          const isSelected = activeColor?.id === color.id;
          const isHovered = hoveredColor?.id === color.id;

          return (
            <Box
              key={color.id}
              role="button"
              tabIndex={0}
              aria-label={color.name}
              onClick={() => handleColorClick(color)}
              onMouseEnter={() => setHoveredColor(color)}
              onMouseLeave={() => setHoveredColor(null)}
              style={{
                position: 'absolute',
                left: `calc(50% + ${x}px - ${buttonSize / 2}px)`,
                top: `calc(50% + ${y}px - ${buttonSize / 2}px)`,
                width: buttonSize,
                height: buttonSize,
                borderRadius: '50%',
                backgroundColor: color.hex_code,
                border: isSelected
                  ? '3px solid #ffffff'
                  : '2px solid rgba(255,255,255,0.7)',
                boxShadow: isSelected
                  ? '0 0 16px #ffffff, 0 4px 10px rgba(0,0,0,0.5)'
                  : isHovered
                    ? `0 0 12px ${color.hex_code}, 0 2px 6px rgba(0,0,0,0.4)`
                    : '0 2px 6px rgba(0,0,0,0.3)',
                transform: isSelected || isHovered ? 'scale(1.22)' : 'scale(1)',
                transition: 'all 0.15s cubic-bezier(0.34, 1.56, 0.64, 1)',
                cursor: disabled ? 'not-allowed' : 'pointer',
                zIndex: isSelected || isHovered ? 5 : 1,
              }}
            />
          );
        })}
      </Box>
    </Modal>
  );
}
