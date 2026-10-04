'use client';

import { Box, Button, Group, Text } from '@mantine/core';
import { IconArrowRight, IconCheck, IconX } from '@tabler/icons-react';
import type { Color } from '@/types/generated';

interface InlineFeedbackBarProps {
  isCorrect: boolean;
  correctLeftColor?: Color;
  correctRightColor?: Color;
  onNext: () => void;
}

export function InlineFeedbackBar({
  isCorrect,
  correctLeftColor,
  correctRightColor,
  onNext,
}: InlineFeedbackBarProps) {
  return (
    <Box
      py={8}
      px={10}
      style={{
        width: '100%',
        maxWidth: 440,
        borderRadius: 14,
        backgroundColor: isCorrect
          ? 'rgba(43, 138, 62, 0.95)'
          : 'rgba(201, 42, 42, 0.95)',
        color: '#ffffff',
        backdropFilter: 'blur(12px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        gap: 6,
        boxShadow: '0 4px 16px rgba(0, 0, 0, 0.3)',
        animation: 'popIn 0.2s cubic-bezier(0.34, 1.56, 0.64, 1)',
        flexShrink: 0,
      }}
    >
      {/* 左側: 正誤アイコン ＋ 正解色 */}
      <Group
        gap={6}
        wrap="nowrap"
        style={{ overflow: 'hidden', flexShrink: 1 }}
      >
        <Box
          style={{
            width: 28,
            height: 28,
            borderRadius: '50%',
            backgroundColor: 'rgba(255, 255, 255, 0.25)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
          }}
        >
          {isCorrect ? (
            <IconCheck size={18} stroke={3} />
          ) : (
            <IconX size={18} stroke={3} />
          )}
        </Box>

        {isCorrect ? (
          <Text size="sm" fw={800} style={{ whiteSpace: 'nowrap' }}>
            大正解！
          </Text>
        ) : (
          <Group gap={4} wrap="nowrap" style={{ overflow: 'hidden' }}>
            <Text
              size="xs"
              fw={700}
              style={{ whiteSpace: 'nowrap', flexShrink: 0 }}
            >
              正解:
            </Text>
            {correctLeftColor && (
              <Box
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 5,
                  backgroundColor: 'rgba(0, 0, 0, 0.65)',
                  border: `1.5px solid ${correctLeftColor.hex_code}`,
                  boxShadow: `0 0 10px ${correctLeftColor.hex_code}88`,
                  padding: '3px 8px',
                  borderRadius: 16,
                  whiteSpace: 'nowrap',
                  flexShrink: 0,
                }}
              >
                <Box
                  style={{
                    width: 10,
                    height: 10,
                    borderRadius: '50%',
                    backgroundColor: correctLeftColor.hex_code,
                    boxShadow: `0 0 6px ${correctLeftColor.hex_code}`,
                    flexShrink: 0,
                  }}
                />
                <Text
                  size="11px"
                  fw={800}
                  c="#ffffff"
                  style={{ whiteSpace: 'nowrap', lineHeight: 1 }}
                >
                  {correctLeftColor.name}
                </Text>
              </Box>
            )}
            <Text
              size="xs"
              fw={900}
              c="rgba(255,255,255,0.7)"
              style={{ flexShrink: 0 }}
            >
              +
            </Text>
            {correctRightColor && (
              <Box
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 5,
                  backgroundColor: 'rgba(0, 0, 0, 0.65)',
                  border: `1.5px solid ${correctRightColor.hex_code}`,
                  boxShadow: `0 0 10px ${correctRightColor.hex_code}88`,
                  padding: '3px 8px',
                  borderRadius: 16,
                  whiteSpace: 'nowrap',
                  flexShrink: 0,
                }}
              >
                <Box
                  style={{
                    width: 10,
                    height: 10,
                    borderRadius: '50%',
                    backgroundColor: correctRightColor.hex_code,
                    boxShadow: `0 0 6px ${correctRightColor.hex_code}`,
                    flexShrink: 0,
                  }}
                />
                <Text
                  size="11px"
                  fw={800}
                  c="#ffffff"
                  style={{ whiteSpace: 'nowrap', lineHeight: 1 }}
                >
                  {correctRightColor.name}
                </Text>
              </Box>
            )}
          </Group>
        )}
      </Group>

      {/* 右側: 次へ進むボタン */}
      <Button
        size="xs"
        variant="white"
        color="dark"
        radius="xl"
        rightSection={<IconArrowRight size={14} />}
        onClick={onNext}
        style={{
          flexShrink: 0,
          fontWeight: 800,
          paddingLeft: 10,
          paddingRight: 10,
          height: 32,
        }}
      >
        次へ
      </Button>
    </Box>
  );
}
