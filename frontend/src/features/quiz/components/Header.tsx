'use client';

import {
  ActionIcon,
  Box,
  Group,
  Menu,
  Text,
  useMantineColorScheme,
} from '@mantine/core';
import {
  IconCheck,
  IconColorSwatch,
  IconFilter,
  IconLayout,
  IconMoon,
  IconSun,
} from '@tabler/icons-react';
import type { LayoutMode } from '@/features/quiz/types';

interface HeaderProps {
  layoutMode: LayoutMode;
  onLayoutModeChange: (mode: LayoutMode) => void;
  onOpenFilter?: () => void;
}

export function Header({
  layoutMode,
  onLayoutModeChange,
  onOpenFilter,
}: HeaderProps) {
  const { colorScheme, toggleColorScheme } = useMantineColorScheme();
  const isDark = colorScheme === 'dark';

  return (
    <Box
      py="xs"
      px="md"
      style={{
        width: '100%',
        maxWidth: 440,
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        borderBottom: '1px solid var(--mantine-color-default-border)',
      }}
    >
      {/* 左: ロゴ & タイトル (旧版デザイン準拠) */}
      <Group gap={8} align="center">
        <Box
          style={{
            width: 28,
            height: 28,
            borderRadius: 6,
            backgroundColor: '#7CC7E8',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
          }}
        >
          <IconColorSwatch size={18} />
        </Box>
        <Text size="md" fw={700} c="blue.7">
          ペンライトクイズ
        </Text>
      </Group>

      {/* 右: レイアウト切替、フィルター、ダークモード切替 */}
      <Group gap={6}>
        {/* レイアウト切り替えメニュー (クラシック vs オーバーレイ vs コンパクト) */}
        <Menu shadow="md" width={180}>
          <Menu.Target>
            <ActionIcon
              variant="light"
              color="gray"
              size="lg"
              radius="md"
              title="レイアウト変更"
            >
              <IconLayout size={18} />
            </ActionIcon>
          </Menu.Target>
          <Menu.Dropdown>
            <Menu.Label>出題レイアウト</Menu.Label>
            <Menu.Item
              leftSection={
                layoutMode === 'overlay' ? (
                  <IconCheck size={14} />
                ) : (
                  <Box w={14} />
                )
              }
              onClick={() => onLayoutModeChange('overlay')}
            >
              オーバーレイ没入型 ★
            </Menu.Item>
            <Menu.Item
              leftSection={
                layoutMode === 'classic' ? (
                  <IconCheck size={14} />
                ) : (
                  <Box w={14} />
                )
              }
              onClick={() => onLayoutModeChange('classic')}
            >
              旧版クラシック (縦並び)
            </Menu.Item>
            <Menu.Item
              leftSection={
                layoutMode === 'compact' ? (
                  <IconCheck size={14} />
                ) : (
                  <Box w={14} />
                )
              }
              onClick={() => onLayoutModeChange('compact')}
            >
              コンパクト操作重視
            </Menu.Item>
          </Menu.Dropdown>
        </Menu>

        {/* フィルターボタン (旧版準拠) */}
        <ActionIcon
          variant="light"
          color="blue"
          size="lg"
          radius="md"
          onClick={onOpenFilter}
          title="絞り込みフィルター"
        >
          <IconFilter size={18} />
        </ActionIcon>

        {/* ダークモードボタン (旧版準拠) */}
        <ActionIcon
          variant="light"
          color={isDark ? 'yellow' : 'gray'}
          size="lg"
          radius="md"
          onClick={() => toggleColorScheme()}
          title="テーマ切り替え"
        >
          {isDark ? <IconSun size={18} /> : <IconMoon size={18} />}
        </ActionIcon>
      </Group>
    </Box>
  );
}
