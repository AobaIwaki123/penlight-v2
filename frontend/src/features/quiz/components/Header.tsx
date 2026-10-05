'use client';

import {
  ActionIcon,
  Badge,
  Box,
  Group,
  Menu,
  Text,
  useMantineColorScheme,
} from '@mantine/core';
import {
  IconCheck,
  IconChevronLeft,
  IconCircleDot,
  IconColorSwatch,
  IconLayout,
  IconMoon,
  IconMusic,
  IconSun,
} from '@tabler/icons-react';
import { useEffect, useState } from 'react';
import type { InputMode, LayoutMode } from '@/features/quiz/types';

interface HeaderProps {
  layoutMode: LayoutMode;
  onLayoutModeChange: (mode: LayoutMode) => void;
  inputMode: InputMode;
  onInputModeChange: (mode: InputMode) => void;
  onGoHome?: () => void;
  groupThemeColor?: string;
  seriesName?: string;
  groupName?: string;
  isSongMode?: boolean;
}

export function Header({
  layoutMode,
  onLayoutModeChange,
  inputMode,
  onInputModeChange,
  onGoHome,
  groupThemeColor,
  seriesName,
  groupName,
  isSongMode,
}: HeaderProps) {
  const { colorScheme, toggleColorScheme } = useMantineColorScheme();
  const isDark = colorScheme === 'dark';

  // ローカル環境（localhost / 127.0.0.1）または ?dev=true の時のみ開発・切替メニューを表示
  const [isDev, setIsDev] = useState(false);

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const hostname = window.location.hostname;
      const isLocal =
        hostname === 'localhost' ||
        hostname === '127.0.0.1' ||
        hostname === '[::1]' ||
        hostname.endsWith('.local') ||
        window.location.search.includes('dev=true');
      setIsDev(isLocal);
    }
  }, []);

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
      {/* 左: ポータルへ戻るボタン または ロゴ */}
      <Group gap={8} align="center">
        {onGoHome ? (
          <ActionIcon
            variant="subtle"
            color="gray"
            size="md"
            radius="md"
            onClick={onGoHome}
            title="トップへ戻る"
          >
            <IconChevronLeft size={20} />
          </ActionIcon>
        ) : null}

        <Box
          style={{
            width: 28,
            height: 28,
            borderRadius: 6,
            backgroundColor: groupThemeColor || '#7CC7E8',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: '#fff',
          }}
        >
          {isSongMode ? <IconMusic size={18} /> : <IconColorSwatch size={18} />}
        </Box>

        <Box>
          <Group gap={4} align="center">
            <Text size="sm" fw={700} c="blue.7">
              {groupName || 'ペンライトクイズ'}
            </Text>
            {isSongMode && (
              <Badge size="xs" variant="dot" color="violet">
                楽曲
              </Badge>
            )}
          </Group>
          {seriesName && (
            <Text size="10px" c="dimmed" lh={1}>
              {seriesName}
            </Text>
          )}
        </Box>
      </Group>

      {/* 右: レイアウト切替、フィルター、ダークモード切替 */}
      <Group gap={6}>
        {/* ローカル環境（localhost）のみ表示する開発・切り替えメニュー */}
        {isDev && (
          <>
            {/* レイアウト切り替えメニュー (クラシック vs オーバーレイ vs コンパクト) */}
            <Menu shadow="md" width={180}>
              <Menu.Target>
                <ActionIcon
                  variant="light"
                  color="gray"
                  size="lg"
                  radius="md"
                  title="レイアウト変更 (開発用)"
                >
                  <IconLayout size={18} />
                </ActionIcon>
              </Menu.Target>
              <Menu.Dropdown>
                <Menu.Label>出題レイアウト (Dev)</Menu.Label>
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

            {/* パレット入力方式切り替えメニュー (ドーナツ vs 15色グリッド) */}
            <Menu shadow="md" width={180}>
              <Menu.Target>
                <ActionIcon
                  variant="light"
                  color={inputMode === 'donut' ? 'indigo' : 'gray'}
                  size="lg"
                  radius="md"
                  title="カラーパレット方式 (開発用)"
                >
                  <IconCircleDot size={18} />
                </ActionIcon>
              </Menu.Target>
              <Menu.Dropdown>
                <Menu.Label>パレットUI (Dev)</Menu.Label>
                <Menu.Item
                  leftSection={
                    inputMode === 'donut' ? (
                      <IconCheck size={14} />
                    ) : (
                      <Box w={14} />
                    )
                  }
                  onClick={() => onInputModeChange('donut')}
                >
                  ドーナツサークル型 ★
                </Menu.Item>
                <Menu.Item
                  leftSection={
                    inputMode === 'grid' ? (
                      <IconCheck size={14} />
                    ) : (
                      <Box w={14} />
                    )
                  }
                  onClick={() => onInputModeChange('grid')}
                >
                  15色グリッド型
                </Menu.Item>
              </Menu.Dropdown>
            </Menu>
          </>
        )}

        {/* ダークモードボタン */}
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
