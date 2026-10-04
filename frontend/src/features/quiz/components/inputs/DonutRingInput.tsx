'use client';

import { Badge, Box, Button, Paper, Text } from '@mantine/core';
import { useState } from 'react';
import type { AnswerInputProps } from '@/features/quiz/types';
import type { Color } from '@/types/generated';

interface DonutRingInputProps extends AnswerInputProps {
  onColorSelect?: (step: 'left' | 'right', color: Color) => void;
  onResetSelection?: () => void;
}

export function DonutRingInput({
  colors,
  onAnswer,
  disabled,
  onColorSelect,
  onResetSelection,
}: DonutRingInputProps) {
  const [firstColor, setFirstColor] = useState<Color | null>(null);
  const [hoveredColor, setHoveredColor] = useState<Color | null>(null);

  const radius = 118; // 円環の半径 (px)
  const buttonSize = 38; // 各カラージュエルの直径 (px)

  const handleColorClick = (color: Color) => {
    if (disabled) return;

    if (!firstColor) {
      // Step 1: Left hand color chosen
      setFirstColor(color);
      onColorSelect?.('left', color);
    } else {
      // Step 2: Right hand color chosen -> trigger immediate judgment
      const leftId = firstColor.id;
      const rightId = color.id;
      onColorSelect?.('right', color);
      onAnswer({ leftColorId: leftId, rightColorId: rightId });
      setFirstColor(null);
      setHoveredColor(null);
    }
  };

  const handleReset = () => {
    setFirstColor(null);
    setHoveredColor(null);
    onResetSelection?.();
  };

  // Center display: current active/hovered color name or prompt
  const displayColor = hoveredColor || firstColor;

  return (
    <Box
      style={{
        width: '100%',
        maxWidth: 380,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        padding: '8px 0',
      }}
    >
      {/* ドーナツリング本体 */}
      <Box
        style={{
          position: 'relative',
          width: radius * 2 + buttonSize + 16,
          height: radius * 2 + buttonSize + 16,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        {/* 背景のほのかな光彩リング */}
        <Box
          style={{
            position: 'absolute',
            width: radius * 2,
            height: radius * 2,
            borderRadius: '50%',
            border: '1px dashed var(--mantine-color-default-border)',
            pointerEvents: 'none',
          }}
        />

        {/* ドーナツ中央の空洞: 選択状態 & 色名表示エリア */}
        <Paper
          radius="50%"
          shadow="sm"
          withBorder
          style={{
            width: 130,
            height: 130,
            borderRadius: '50%',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            textAlign: 'center',
            padding: 8,
            zIndex: 2,
            backgroundColor: 'var(--mantine-color-body)',
            transition: 'all 0.2s ease',
          }}
        >
          <Badge
            size="xs"
            variant="filled"
            color={firstColor ? 'blue' : 'gray'}
            mb={4}
          >
            {firstColor ? '右手を選択' : '左手を選択'}
          </Badge>

          <Text
            size="sm"
            fw={700}
            c={displayColor ? 'inherit' : 'dimmed'}
            style={{
              lineHeight: 1.2,
              minHeight: 20,
              maxWidth: 110,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteHeight: 'nowrap',
            }}
          >
            {displayColor ? displayColor.name : '色をタップ'}
          </Text>

          {firstColor && (
            <Button
              size="compact-xs"
              variant="subtle"
              color="gray"
              mt={4}
              onClick={handleReset}
              disabled={disabled}
              style={{ fontSize: 10, height: 18 }}
            >
              選び直す
            </Button>
          )}
        </Paper>

        {/* 円周上に配置された 15 個のカラージュエル */}
        {colors.map((color, index) => {
          const total = colors.length;
          // Calculate angle: start from top (-PI/2)
          const angle = (2 * Math.PI * index) / total - Math.PI / 2;
          const x = radius * Math.cos(angle);
          const y = radius * Math.sin(angle);

          const isSelectedAsFirst = firstColor?.id === color.id;
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
                border: isSelectedAsFirst
                  ? '3px solid #228be6'
                  : '2px solid rgba(255,255,255,0.85)',
                boxShadow: isSelectedAsFirst
                  ? '0 0 12px #228be6, 0 2px 6px rgba(0,0,0,0.3)'
                  : isHovered
                    ? `0 0 10px ${color.hex_code}, 0 2px 5px rgba(0,0,0,0.2)`
                    : '0 2px 4px rgba(0,0,0,0.15)',
                transform:
                  isSelectedAsFirst || isHovered ? 'scale(1.18)' : 'scale(1)',
                transition: 'all 0.15s cubic-bezier(0.34, 1.56, 0.64, 1)',
                cursor: disabled ? 'not-allowed' : 'pointer',
                zIndex: isSelectedAsFirst || isHovered ? 5 : 1,
              }}
            />
          );
        })}
      </Box>
    </Box>
  );
}
