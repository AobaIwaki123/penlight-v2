'use client';

import { Box, Stack, Text } from '@mantine/core';
import type { Color } from '@/types/generated';

interface PenlightStickProps {
  color?: Color;
  label?: string;
  height?: number;
  width?: number;
  textColor?: string;
}

export function PenlightStick({
  color,
  label,
  height = 90,
  width = 38,
  textColor,
}: PenlightStickProps) {
  const hex = color?.hex_code || '#E0E0E0';
  const name = color?.name || '未選択';
  const isSelected = Boolean(color);

  // 文字数に応じた動的フォントサイズ調整 (最大文字数「エメラルドグリーン」等でも1行に収める)
  const fontSize =
    name.length >= 7 ? '9px' : name.length >= 5 ? '10px' : '11px';

  return (
    <Stack gap={2} align="center" style={{ minWidth: 44, maxWidth: 84 }}>
      <Box
        style={{
          width,
          height,
          borderRadius: 12,
          backgroundColor: hex,
          border: isSelected
            ? '2px solid rgba(0,0,0,0.1)'
            : '2px dashed #BDBDBD',
          boxShadow: isSelected
            ? `0 0 16px ${hex}88, inset 0 0 8px rgba(255,255,255,0.4)`
            : 'none',
          transition: 'all 0.25s ease',
        }}
      />
      {isSelected && (
        <Text
          fw={700}
          c={textColor || 'inherit'}
          ta="center"
          style={{
            fontSize,
            letterSpacing: name.length >= 7 ? '-0.04em' : '-0.01em',
            textShadow: textColor ? '0 1px 3px rgba(0,0,0,0.8)' : undefined,
            lineHeight: 1.15,
            whiteSpace: 'nowrap',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            maxWidth: '100%',
            display: 'block',
          }}
        >
          {name}
        </Text>
      )}
      {label && (
        <Text
          size="10px"
          c={textColor || 'dimmed'}
          style={{
            textShadow: textColor ? '0 1px 3px rgba(0,0,0,0.8)' : undefined,
            lineHeight: 1.1,
            whiteSpace: 'nowrap',
          }}
        >
          {label}
        </Text>
      )}
    </Stack>
  );
}
