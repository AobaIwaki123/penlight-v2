'use client';

import { Box, Button, SimpleGrid, Text } from '@mantine/core';
import { useState } from 'react';
import type { AnswerInputProps } from '@/features/quiz/types';
import type { Color } from '@/types/generated';

interface PaletteGridInputProps extends AnswerInputProps {
  onColorSelect?: (step: 'left' | 'right', color: Color) => void;
  onResetSelection?: () => void;
}

export function PaletteGridInput({
  colors,
  onAnswer,
  disabled,
  onColorSelect,
  onResetSelection,
}: PaletteGridInputProps) {
  // Currently selected first color (left hand)
  const [firstColor, setFirstColor] = useState<Color | null>(null);

  const handleColorClick = (color: Color) => {
    if (disabled) return;

    if (!firstColor) {
      // Step 1: Set left hand color
      setFirstColor(color);
      onColorSelect?.('left', color);
    } else {
      // Step 2: Set right hand color and trigger immediate judgment
      const leftId = firstColor.id;
      const rightId = color.id;
      onColorSelect?.('right', color);

      // Trigger answer
      onAnswer({ leftColorId: leftId, rightColorId: rightId });

      // Reset internal selection for next question
      setFirstColor(null);
    }
  };

  const handleCancelSelection = () => {
    setFirstColor(null);
    onResetSelection?.();
  };

  return (
    <Box style={{ width: '100%', maxWidth: 420 }}>
      {/* 選択ガイドと取り消しボタン */}
      <Box
        mb={8}
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <Text size="xs" fw={600} c={firstColor ? 'blue.6' : 'dimmed'}>
          {firstColor
            ? '👉 2本目（右手）をタップしてください'
            : '👉 1本目（左手）をタップしてください'}
        </Text>
        {firstColor && (
          <Button
            size="compact-xs"
            variant="subtle"
            color="gray"
            onClick={handleCancelSelection}
            disabled={disabled}
          >
            選び直す
          </Button>
        )}
      </Box>

      {/* カラーパレット (3列グリッド: タップしやすい高さ44px) */}
      <SimpleGrid cols={3} spacing={6} verticalSpacing={6}>
        {colors.map((c) => {
          const isFirstSelected = firstColor?.id === c.id;
          return (
            <Button
              key={c.id}
              onClick={() => handleColorClick(c)}
              disabled={disabled}
              variant="default"
              h={44}
              p={6}
              radius="md"
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'flex-start',
                gap: 8,
                border: isFirstSelected
                  ? '2px solid #228be6'
                  : '1px solid #e9ecef',
                backgroundColor: isFirstSelected
                  ? 'rgba(34, 139, 230, 0.08)'
                  : undefined,
                transition: 'all 0.15s ease',
              }}
            >
              {/* カラーサークル */}
              <Box
                style={{
                  width: 20,
                  height: 20,
                  borderRadius: '50%',
                  backgroundColor: c.hex_code,
                  border: '1px solid rgba(0,0,0,0.15)',
                  flexShrink: 0,
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                }}
              />
              {/* 色名 */}
              <Text
                size="11px"
                fw={600}
                truncate
                style={{ flexGrow: 1, textAlign: 'left' }}
              >
                {c.name}
              </Text>
            </Button>
          );
        })}
      </SimpleGrid>
    </Box>
  );
}
