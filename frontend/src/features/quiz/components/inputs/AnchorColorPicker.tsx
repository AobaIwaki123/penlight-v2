'use client';

import { Box, Paper, Text } from '@mantine/core';
import { useState } from 'react';
import type { Color } from '@/types/generated';

interface AnchorColorPickerProps {
  opened: boolean;
  onClose: () => void;
  colors: Color[];
  selectedLeftColor?: Color;
  selectedRightColor?: Color;
  activeHand: 'left' | 'right';
  onColorSelect: (hand: 'left' | 'right', color: Color) => void;
  onAnswer: (input: { leftColorId: string; rightColorId: string }) => void;
  disabled: boolean;
}

export function AnchorColorPicker({
  opened,
  onClose,
  colors,
  selectedLeftColor,
  selectedRightColor,
  activeHand: initialHand,
  onColorSelect,
  onAnswer,
  disabled,
}: AnchorColorPickerProps) {
  const [currentHand, setCurrentHand] = useState<'left' | 'right'>(initialHand);
  const [hoveredColor, setHoveredColor] = useState<Color | null>(null);

  if (!opened) return null;

  const radius = 80; // コンパクトな半径
  const buttonSize = 30; // 小さめのジュエル

  const handleColorClick = (color: Color) => {
    if (disabled) return;

    if (currentHand === 'left') {
      onColorSelect('left', color);
      if (selectedRightColor) {
        onAnswer({
          leftColorId: color.id,
          rightColorId: selectedRightColor.id,
        });
        onClose();
      } else {
        setCurrentHand('right');
      }
    } else {
      onColorSelect('right', color);
      if (selectedLeftColor) {
        onAnswer({ leftColorId: selectedLeftColor.id, rightColorId: color.id });
        onClose();
      } else {
        setCurrentHand('left');
      }
    }
  };

  const activeColor =
    currentHand === 'left' ? selectedLeftColor : selectedRightColor;
  const displayColor = hoveredColor || activeColor;

  return (
    <>
      {/* 背景クリックで閉じる透明オーバーレイ */}
      <Box
        onClick={onClose}
        style={{
          position: 'fixed',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          zIndex: 40,
          backgroundColor: 'transparent',
        }}
      />

      {/* ペンライト直上に浮かぶコンパクトサークル */}
      <Box
        style={{
          position: 'absolute',
          bottom: 110,
          right: 16,
          zIndex: 50,
          width: radius * 2 + buttonSize + 16,
          height: radius * 2 + buttonSize + 16,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          animation: 'popIn 0.18s cubic-bezier(0.34, 1.56, 0.64, 1)',
        }}
      >
        {/* 中央バッジ */}
        <Paper
          radius="50%"
          style={{
            width: 80,
            height: 80,
            borderRadius: '50%',
            backgroundColor: 'rgba(0, 0, 0, 0.78)',
            backdropFilter: 'blur(8px)',
            border: `2px solid ${currentHand === 'left' ? '#339af0' : '#ff922b'}`,
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            padding: 4,
            boxShadow: '0 4px 16px rgba(0,0,0,0.5)',
          }}
        >
          <Text
            size="9px"
            fw={800}
            c={currentHand === 'left' ? 'blue.4' : 'orange.4'}
            style={{ textTransform: 'uppercase' }}
          >
            {currentHand === 'left' ? '左手' : '右手'}
          </Text>
          <Text
            size="10px"
            fw={700}
            c={displayColor ? 'white' : 'gray.4'}
            style={{
              whiteSpace: 'nowrap',
              maxWidth: 72,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              textAlign: 'center',
            }}
          >
            {displayColor ? displayColor.name : '色をタップ'}
          </Text>
        </Paper>

        {/* 15色のカラージュエル */}
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
              onClick={(e) => {
                e.stopPropagation();
                handleColorClick(color);
              }}
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
                  ? '2.5px solid #ffffff'
                  : '1.5px solid rgba(255,255,255,0.7)',
                boxShadow: isSelected
                  ? `0 0 12px ${color.hex_code}, 0 2px 8px rgba(0,0,0,0.6)`
                  : isHovered
                    ? `0 0 10px ${color.hex_code}`
                    : '0 2px 5px rgba(0,0,0,0.4)',
                transform: isSelected || isHovered ? 'scale(1.2)' : 'scale(1)',
                transition: 'all 0.12s ease',
                cursor: 'pointer',
                zIndex: isSelected || isHovered ? 10 : 2,
              }}
            />
          );
        })}
      </Box>
    </>
  );
}
