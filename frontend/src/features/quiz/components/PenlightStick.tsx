'use client';

import { Box, Stack, Text } from '@mantine/core';
import type { Color } from '@/types/generated';

interface PenlightStickProps {
  color?: Color;
  label?: string;
  height?: number;
  width?: number;
}

export function PenlightStick({
  color,
  label,
  height = 90,
  width = 38,
}: PenlightStickProps) {
  const hex = color?.hex_code || '#E0E0E0';
  const name = color?.name || '未選択';
  const isSelected = Boolean(color);

  return (
    <Stack gap={4} align="center">
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
      <Text
        size="xs"
        fw={600}
        c={isSelected ? 'inherit' : 'dimmed'}
        ta="center"
      >
        {name}
      </Text>
      {label && (
        <Text size="10px" c="dimmed">
          {label}
        </Text>
      )}
    </Stack>
  );
}
