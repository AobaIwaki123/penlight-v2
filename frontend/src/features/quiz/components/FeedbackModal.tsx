'use client';

import {
  Badge,
  Box,
  Button,
  Group,
  Modal,
  Paper,
  Stack,
  Text,
} from '@mantine/core';
import { IconArrowRight, IconCheck, IconX } from '@tabler/icons-react';
import type { Color } from '@/types/generated';

interface FeedbackModalProps {
  opened: boolean;
  isCorrect: boolean;
  userLeftColor?: Color;
  userRightColor?: Color;
  correctLeftColor?: Color;
  correctRightColor?: Color;
  onNext: () => void;
}

export function FeedbackModal({
  opened,
  isCorrect,
  userLeftColor,
  userRightColor,
  correctLeftColor,
  correctRightColor,
  onNext,
}: FeedbackModalProps) {
  if (!opened) return null;

  return (
    <Modal
      opened={opened}
      onClose={onNext}
      centered
      withCloseButton={false}
      size="sm"
      radius="lg"
      padding="lg"
      overlayProps={{
        backgroundOpacity: 0.65,
        blur: 6,
      }}
      styles={{
        content: {
          backgroundColor: 'var(--mantine-color-body)',
          textAlign: 'center',
          border: isCorrect
            ? '2px solid rgba(43, 138, 62, 0.4)'
            : '2px solid rgba(201, 42, 42, 0.4)',
        },
      }}
    >
      <Stack gap="md" align="center">
        {/* アイコン & タイトル */}
        <Box
          style={{
            width: 56,
            height: 56,
            borderRadius: '50%',
            backgroundColor: isCorrect ? '#ebfbee' : '#fff5f5',
            color: isCorrect ? '#2b8a3e' : '#c92a2a',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          {isCorrect ? (
            <IconCheck size={36} stroke={3} />
          ) : (
            <IconX size={36} stroke={3} />
          )}
        </Box>

        <Box>
          <Text size="xl" fw={800} c={isCorrect ? 'green.7' : 'red.7'}>
            {isCorrect ? '大正解！' : '不正解...'}
          </Text>
          <Text size="xs" c="dimmed" mt={2}>
            {isCorrect
              ? '左右の順番は問わず正解です'
              : '惜しい！正解のカラーを確認しよう'}
          </Text>
        </Box>

        {/* 比較カード */}
        <Paper
          withBorder
          p="sm"
          radius="md"
          style={{
            width: '100%',
            backgroundColor: 'var(--mantine-color-default-hover)',
          }}
        >
          <Stack gap="xs">
            {/* あなたの回答 */}
            <Box
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
            >
              <Text size="xs" fw={600} c="dimmed">
                あなたの回答:
              </Text>
              <Group gap={6}>
                {userLeftColor && (
                  <Badge
                    size="sm"
                    variant="outline"
                    leftSection={
                      <Box
                        style={{
                          width: 10,
                          height: 10,
                          borderRadius: '50%',
                          backgroundColor: userLeftColor.hex_code,
                        }}
                      />
                    }
                  >
                    {userLeftColor.name}
                  </Badge>
                )}
                <Text size="xs" c="dimmed">
                  ×
                </Text>
                {userRightColor && (
                  <Badge
                    size="sm"
                    variant="outline"
                    leftSection={
                      <Box
                        style={{
                          width: 10,
                          height: 10,
                          borderRadius: '50%',
                          backgroundColor: userRightColor.hex_code,
                        }}
                      />
                    }
                  >
                    {userRightColor.name}
                  </Badge>
                )}
              </Group>
            </Box>

            {/* 正解のカラー */}
            <Box
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                paddingTop: 4,
                borderTop: '1px dashed var(--mantine-color-default-border)',
              }}
            >
              <Text size="xs" fw={700} c={isCorrect ? 'green.7' : 'orange.7'}>
                正解のカラー:
              </Text>
              <Group gap={6}>
                {correctLeftColor && (
                  <Badge
                    size="sm"
                    color={isCorrect ? 'green' : 'orange'}
                    variant="filled"
                    leftSection={
                      <Box
                        style={{
                          width: 10,
                          height: 10,
                          borderRadius: '50%',
                          backgroundColor: correctLeftColor.hex_code,
                          border: '1px solid rgba(255,255,255,0.7)',
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
                    color={isCorrect ? 'green' : 'orange'}
                    variant="filled"
                    leftSection={
                      <Box
                        style={{
                          width: 10,
                          height: 10,
                          borderRadius: '50%',
                          backgroundColor: correctRightColor.hex_code,
                          border: '1px solid rgba(255,255,255,0.7)',
                        }}
                      />
                    }
                  >
                    {correctRightColor.name}
                  </Badge>
                )}
              </Group>
            </Box>
          </Stack>
        </Paper>

        {/* 次へ進むボタン */}
        <Button
          fullWidth
          size="md"
          radius="md"
          color={isCorrect ? 'green' : 'blue'}
          rightSection={<IconArrowRight size={18} />}
          onClick={onNext}
          mt="xs"
        >
          次へ進む
        </Button>
      </Stack>
    </Modal>
  );
}
