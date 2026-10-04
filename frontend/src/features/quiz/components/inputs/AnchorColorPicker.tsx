'use client';

import { Box, Paper } from '@mantine/core';
import { IconArrowsLeftRight } from '@tabler/icons-react';
import { useState } from 'react';
import type { Color } from '@/types/generated';

interface AnchorColorPickerProps {
  colors: Color[];
  selectedLeftColor?: Color;
  selectedRightColor?: Color;
  activeHand: 'left' | 'right';
  onHandChange: (hand: 'left' | 'right') => void;
  onColorSelect: (hand: 'left' | 'right', color: Color) => void;
  onAnswer: (input: { leftColorId: string; rightColorId: string }) => void;
  disabled: boolean;
}

export function AnchorColorPicker({
  colors,
  selectedLeftColor,
  selectedRightColor,
  activeHand,
  onHandChange,
  onColorSelect,
  onAnswer,
  disabled,
}: AnchorColorPickerProps) {
  const [hoveredColor, setHoveredColor] = useState<Color | null>(null);

  const radius = 82; // 常時表示に適したコンパクトな半径
  const buttonSize = 30; // 押しやすいジュエル径

  const handleColorClick = (color: Color) => {
    if (disabled) return;

    if (activeHand === 'left') {
      onColorSelect('left', color);
      if (selectedRightColor) {
        // Both selected: short delay for visual satisfaction, then submit
        setTimeout(() => {
          onAnswer({
            leftColorId: color.id,
            rightColorId: selectedRightColor.id,
          });
        }, 280);
      } else {
        // Auto-switch focus to right hand
        onHandChange('right');
      }
    } else {
      onColorSelect('right', color);
      if (selectedLeftColor) {
        // Both selected: short delay for visual satisfaction, then submit
        setTimeout(() => {
          onAnswer({
            leftColorId: selectedLeftColor.id,
            rightColorId: color.id,
          });
        }, 280);
      } else {
        // Auto-switch focus to left hand
        onHandChange('left');
      }
    }
  };

  const activeColor =
    activeHand === 'left' ? selectedLeftColor : selectedRightColor;

  return (
    <Box
      style={{
        position: 'absolute',
        bottom: 12,
        right: 12,
        zIndex: 30,
        width: radius * 2 + buttonSize + 16,
        height: radius * 2 + buttonSize + 16,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        pointerEvents: disabled ? 'none' : 'auto',
        opacity: disabled ? 0.4 : 1,
        transition: 'opacity 0.2s ease',
      }}
    >
      {/* 背景の光彩リング */}
      <Box
        style={{
          position: 'absolute',
          width: radius * 2,
          height: radius * 2,
          borderRadius: '50%',
          border: '1px dashed rgba(255,255,255,0.25)',
          pointerEvents: 'none',
        }}
      />

      {/* 中央ミニマル切り替えボタン: タップで左↔右を瞬時にスイッチ */}
      <Paper
        role="button"
        tabIndex={0}
        onClick={() => onHandChange(activeHand === 'left' ? 'right' : 'left')}
        title={`タップして${activeHand === 'left' ? '右手' : '左手'}へ切り替え`}
        radius="50%"
        style={{
          width: 54,
          height: 54,
          borderRadius: '50%',
          backgroundColor: 'rgba(0, 0, 0, 0.75)',
          backdropFilter: 'blur(10px)',
          border:
            activeHand === 'left' ? '2px solid #339af0' : '2px solid #ff922b',
          boxShadow:
            activeHand === 'left'
              ? '0 0 16px rgba(51, 154, 240, 0.75), inset 0 0 8px rgba(51, 154, 240, 0.25)'
              : '0 0 16px rgba(255, 146, 43, 0.75), inset 0 0 8px rgba(255, 146, 43, 0.25)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          cursor: 'pointer',
          zIndex: 2,
          transition: 'all 0.2s cubic-bezier(0.34, 1.56, 0.64, 1)',
        }}
      >
        <IconArrowsLeftRight
          size={22}
          color={activeHand === 'left' ? '#339af0' : '#ff922b'}
          stroke={2.2}
        />
      </Paper>

      {/* 円周上に並ぶ 15 色のカラージュエル */}
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
                ? '2.5px solid #ffffff'
                : '1.5px solid rgba(255,255,255,0.75)',
              boxShadow: isSelected
                ? `0 0 12px ${color.hex_code}, 0 2px 8px rgba(0,0,0,0.6)`
                : isHovered
                  ? `0 0 10px ${color.hex_code}`
                  : '0 2px 5px rgba(0,0,0,0.4)',
              transform: isSelected || isHovered ? 'scale(1.2)' : 'scale(1)',
              transition: 'all 0.12s cubic-bezier(0.34, 1.56, 0.64, 1)',
              cursor: disabled ? 'not-allowed' : 'pointer',
              zIndex: isSelected || isHovered ? 10 : 2,
            }}
          />
        );
      })}
    </Box>
  );
}
