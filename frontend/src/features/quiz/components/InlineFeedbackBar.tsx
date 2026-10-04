'use client';

import { Badge, Box, Button, Group, Text } from '@mantine/core';
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
      px="md"
      style={{
        width: '100%',
        maxWidth: 440,
        borderRadius: 12,
        backgroundColor: isCorrect
          ? 'rgba(43, 138, 62, 0.92)'
          : 'rgba(201, 42, 42, 0.92)',
        color: '#ffffff',
        backdropFilter: 'blur(10px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        gap: 8,
        boxShadow: '0 4px 14px rgba(0, 0, 0, 0.25)',
        animation: 'popIn 0.2s ease',
      }}
    >
      {/* 左側: 正誤アイコン ＋ 正解色 */}
      <Group gap={8} wrap="nowrap" style={{ overflow: 'hidden' }}>
        <Box
          style={{
            width: 26,
            height: 26,
            borderRadius: '50%',
            backgroundColor: 'rgba(255, 255, 255, 0.2)',
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
          <Group gap={6} wrap="nowrap" style={{ overflow: 'hidden' }}>
            <Text size="xs" fw={700} style={{ whiteSpace: 'nowrap' }}>
              正解:
            </Text>
            {correctLeftColor && (
              <Badge
                size="sm"
                variant="filled"
                color="dark"
                style={{
                  backgroundColor: 'rgba(0,0,0,0.4)',
                  paddingLeft: 4,
                  paddingRight: 8,
                }}
                leftSection={
                  <Box
                    style={{
                      width: 10,
                      height: 10,
                      borderRadius: '50%',
                      backgroundColor: correctLeftColor.hex_code,
                      border: '1px solid rgba(255,255,255,0.8)',
                    }}
                  />
                }
              >
                {correctLeftColor.name}
              </Badge>
            )}
            <Text size="xs" fw={700}>
              ×
            </Text>
            {correctRightColor && (
              <Badge
                size="sm"
                variant="filled"
                color="dark"
                style={{
                  backgroundColor: 'rgba(0,0,0,0.4)',
                  paddingLeft: 4,
                  paddingRight: 8,
                }}
                leftSection={
                  <Box
                    style={{
                      width: 10,
                      height: 10,
                      borderRadius: '50%',
                      backgroundColor: correctRightColor.hex_code,
                      border: '1px solid rgba(255,255,255,0.8)',
                    }}
                  />
                }
              >
                {correctRightColor.name}
              </Badge>
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
        style={{ flexShrink: 0, fontWeight: 700 }}
      >
        次へ
      </Button>
    </Box>
  );
}
