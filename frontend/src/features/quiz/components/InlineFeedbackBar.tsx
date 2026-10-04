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
        gap: 8,
        boxShadow: '0 4px 16px rgba(0, 0, 0, 0.3)',
        animation: 'popIn 0.2s cubic-bezier(0.34, 1.56, 0.64, 1)',
        flexShrink: 0,
      }}
    >
      {/* 左側: 正誤アイコン ＋ 正解色 (幅不足時は折り返して見切れを防ぐ) */}
      <Group
        gap={6}
        wrap="nowrap"
        align="center"
        style={{ flex: 1, minWidth: 0 }}
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
          <Group gap={4} wrap="wrap" style={{ flex: 1, minWidth: 0 }}>
            <Text
              size="xs"
              fw={700}
              style={{ whiteSpace: 'nowrap', flexShrink: 0 }}
            >
              不正解:
            </Text>
            {correctLeftColor && (
              <Box
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 4,
                  backgroundColor: 'rgba(0,0,0,0.5)',
                  padding: '3px 7px',
                  borderRadius: 8,
                  whiteSpace: 'nowrap',
                  flexShrink: 0,
                }}
              >
                <Box
                  style={{
                    width: 8,
                    height: 8,
                    borderRadius: '50%',
                    backgroundColor: correctLeftColor.hex_code,
                    border: '1px solid rgba(255,255,255,0.9)',
                    flexShrink: 0,
                  }}
                />
                <Text
                  size="11px"
                  fw={700}
                  style={{ whiteSpace: 'nowrap', lineHeight: 1 }}
                >
                  {correctLeftColor.name}
                </Text>
              </Box>
            )}
            <Text size="xs" fw={700} style={{ flexShrink: 0 }}>
              ×
            </Text>
            {correctRightColor && (
              <Box
                style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 4,
                  backgroundColor: 'rgba(0,0,0,0.5)',
                  padding: '3px 7px',
                  borderRadius: 8,
                  whiteSpace: 'nowrap',
                  flexShrink: 0,
                }}
              >
                <Box
                  style={{
                    width: 8,
                    height: 8,
                    borderRadius: '50%',
                    backgroundColor: correctRightColor.hex_code,
                    border: '1px solid rgba(255,255,255,0.9)',
                    flexShrink: 0,
                  }}
                />
                <Text
                  size="11px"
                  fw={700}
                  style={{ whiteSpace: 'nowrap', lineHeight: 1 }}
                >
                  {correctRightColor.name}
                </Text>
              </Box>
            )}
          </Group>
        )}
      </Group>

      {/* 右側: 次へ進むボタン (大きく・白抜き・脈動で気付きやすく) */}
      <style>{`
        @keyframes nextPulse {
          0%, 100% { box-shadow: 0 0 0 0 rgba(255,255,255,0.75); }
          50% { box-shadow: 0 0 0 9px rgba(255,255,255,0); }
        }
        @keyframes nextNudge {
          0%, 100% { transform: translateX(0); }
          50% { transform: translateX(4px); }
        }
        @keyframes popIn {
          from { transform: translateY(8px); opacity: 0; }
          to { transform: translateY(0); opacity: 1; }
        }
      `}</style>
      <Button
        size="md"
        variant="white"
        color={isCorrect ? 'green.9' : 'red.9'}
        radius="xl"
        rightSection={
          <IconArrowRight
            size={20}
            stroke={3}
            style={{ animation: 'nextNudge 0.9s ease-in-out infinite' }}
          />
        }
        onClick={onNext}
        style={{
          flexShrink: 0,
          fontWeight: 900,
          fontSize: 16,
          paddingLeft: 18,
          paddingRight: 14,
          height: 46,
          animation: 'nextPulse 1.4s ease-out infinite',
        }}
      >
        次へ
      </Button>
    </Box>
  );
}
