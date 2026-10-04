'use client';

import {
  Badge,
  Box,
  Modal,
  Paper,
  SegmentedControl,
  Text,
} from '@mantine/core';
import { useState } from 'react';
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

  const radius = 114; // 円環の半径 (px)
  const buttonSize = 38; // 各カラージュエルの直径 (px)

  const handleColorClick = (color: Color) => {
    if (disabled) return;

    if (currentHand === 'left') {
      onColorSelect('left', color);
      // If right hand is already selected, submit answer and close
      if (selectedRightColor) {
        onAnswer({
          leftColorId: color.id,
          rightColorId: selectedRightColor.id,
        });
        onClose();
      } else {
        // Automatically advance to right hand
        setCurrentHand('right');
      }
    } else {
      onColorSelect('right', color);
      // If left hand is already selected, submit answer and close
      if (selectedLeftColor) {
        onAnswer({ leftColorId: selectedLeftColor.id, rightColorId: color.id });
        onClose();
      } else {
        // Automatically switch to left hand
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
      withCloseButton={true}
      size="auto"
      radius="xl"
      padding="md"
      title={
        <Text size="sm" fw={700}>
          ペンライトカラー選択
        </Text>
      }
      overlayProps={{
        backgroundOpacity: 0.65,
        blur: 5,
      }}
      styles={{
        content: {
          backgroundColor: 'var(--mantine-color-body)',
        },
      }}
    >
      <Box
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: 12,
          padding: '4px 0',
        }}
      >
        {/* 左手 / 右手 の切り替えタブ */}
        <SegmentedControl
          value={currentHand}
          onChange={(val) => setCurrentHand(val as 'left' | 'right')}
          data={[
            {
              label: selectedLeftColor
                ? `左手: ${selectedLeftColor.name}`
                : '👈 1本目 (左手)',
              value: 'left',
            },
            {
              label: selectedRightColor
                ? `右手: ${selectedRightColor.name}`
                : '👉 2本目 (右手)',
              value: 'right',
            },
          ]}
          color="blue"
          size="xs"
          radius="xl"
        />

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
          {/* 背景の光彩リング */}
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
              width: 125,
              height: 125,
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
              color={currentHand === 'left' ? 'blue' : 'orange'}
              mb={4}
            >
              {currentHand === 'left' ? '左手の色' : '右手の色'}
            </Badge>

            <Text
              size="sm"
              fw={700}
              c={displayColor ? 'inherit' : 'dimmed'}
              style={{
                lineHeight: 1.2,
                minHeight: 20,
                maxWidth: 105,
                overflow: 'hidden',
                textOverflow: 'ellipsis',
                whiteSpace: 'nowrap',
              }}
            >
              {displayColor ? displayColor.name : '色をタップ'}
            </Text>

            <Text size="10px" c="dimmed" mt={4}>
              {currentHand === 'left'
                ? 'タップして左手を決定'
                : 'タップして右手を決定'}
            </Text>
          </Paper>

          {/* 円周上に配置された 15 個のカラージュエル */}
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
                    ? '3px solid #228be6'
                    : '2px solid rgba(255,255,255,0.85)',
                  boxShadow: isSelected
                    ? '0 0 12px #228be6, 0 2px 6px rgba(0,0,0,0.3)'
                    : isHovered
                      ? `0 0 10px ${color.hex_code}, 0 2px 5px rgba(0,0,0,0.2)`
                      : '0 2px 4px rgba(0,0,0,0.15)',
                  transform:
                    isSelected || isHovered ? 'scale(1.2)' : 'scale(1)',
                  transition: 'all 0.15s cubic-bezier(0.34, 1.56, 0.64, 1)',
                  cursor: disabled ? 'not-allowed' : 'pointer',
                  zIndex: isSelected || isHovered ? 5 : 1,
                }}
              />
            );
          })}
        </Box>
      </Box>
    </Modal>
  );
}
