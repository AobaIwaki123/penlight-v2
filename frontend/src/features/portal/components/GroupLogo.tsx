'use client';

import { Text, useComputedColorScheme } from '@mantine/core';
import { useState } from 'react';

// グループ slug → 静的ロゴ (frontend/public/logos/)。単一バイナリに内包される (ADR-0002, ADR-0007)。
// invertOnDark: 黒単色ロゴはダークモードで反転して視認性を確保する。
const LOGOS: Record<string, { file: string; invertOnDark: boolean }> = {
  hinatazaka46: { file: 'hinatazaka46.png', invertOnDark: false },
  sakurazaka46: { file: 'sakurazaka46.svg', invertOnDark: false },
  nogizaka46: { file: 'nogizaka46.png', invertOnDark: false },
  equal_love: { file: 'equal_love.png', invertOnDark: true },
  not_equal_me: { file: 'not_equal_me.png', invertOnDark: true },
  nearly_equal_joy: { file: 'nearly_equal_joy.png', invertOnDark: true },
};

interface GroupLogoProps {
  slug: string;
  name: string;
  height?: number;
  dimmed?: boolean;
}

export function GroupLogo({
  slug,
  name,
  height = 36,
  dimmed = false,
}: GroupLogoProps) {
  const scheme = useComputedColorScheme('light');
  const [failed, setFailed] = useState(false);
  const logo = LOGOS[slug];

  // ロゴ未登録・読み込み失敗時はグループ名テキストにフォールバック
  if (!logo || failed) {
    return (
      <Text size="sm" fw={700} ta="center">
        {name}
      </Text>
    );
  }

  return (
    // biome-ignore lint/performance/noImgElement: 静的エクスポート (output: 'export') では next/image 最適化を使わない
    <img
      src={`/logos/${logo.file}`}
      alt={name}
      onError={() => setFailed(true)}
      style={{
        height,
        maxWidth: '100%',
        objectFit: 'contain',
        opacity: dimmed ? 0.55 : 1,
        filter: logo.invertOnDark && scheme === 'dark' ? 'invert(1)' : 'none',
        transition: 'opacity 0.15s ease',
      }}
    />
  );
}
