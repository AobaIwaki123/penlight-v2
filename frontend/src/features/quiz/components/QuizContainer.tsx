'use client';

import {
  Badge,
  Box,
  Button,
  Container,
  Loader,
  Modal,
  Progress,
  Stack,
  Text,
} from '@mantine/core';
import { IconRotateClockwise, IconTrophy } from '@tabler/icons-react';
import { useEffect, useState } from 'react';
import { fetchBootstrapData } from '@/features/quiz/api/client';
import { Header } from '@/features/quiz/components/Header';
import { PaletteGridInput } from '@/features/quiz/components/inputs/PaletteGridInput';
import { LayoutClassic } from '@/features/quiz/components/layouts/LayoutClassic';
import { LayoutCompact } from '@/features/quiz/components/layouts/LayoutCompact';
import { LayoutOverlay } from '@/features/quiz/components/layouts/LayoutOverlay';
import type { LayoutMode } from '@/features/quiz/types';
import type { Color, Member } from '@/types/generated';

export function QuizContainer() {
  // Master data from backend API
  const [members, setMembers] = useState<Member[]>([]);
  const [colors, setColors] = useState<Color[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  // Settings
  const [layoutMode, setLayoutMode] = useState<LayoutMode>('overlay');

  // Session state
  const [currentIndex, setCurrentIndex] = useState(0);
  const [score, setScore] = useState(0);
  const [isFinished, setIsFinished] = useState(false);

  // Current question selection state
  const [selectedLeft, setSelectedLeft] = useState<Color | undefined>();
  const [selectedRight, setSelectedRight] = useState<Color | undefined>();
  const [feedback, setFeedback] = useState<'idle' | 'correct' | 'wrong'>(
    'idle',
  );

  useEffect(() => {
    fetchBootstrapData()
      .then((data) => {
        // Filter active members with penlight configuration
        const activeMembers = (data.members || []).filter(
          (m: Member) => m.status === 'active' && m.penlight?.left_color_id,
        );
        // Default to Hinatazaka46 members if available, or all active
        const hinata = activeMembers.filter(
          (m: Member) => m.group_id === 'grp_e6722901acc15ce2af3dacee3a83840c',
        );
        setMembers(hinata.length > 0 ? hinata : activeMembers);

        // Filter colors for the chosen group
        const groupColors = (data.colors || []).filter(
          (c: Color) => c.group_id === 'grp_e6722901acc15ce2af3dacee3a83840c',
        );
        setColors(
          groupColors.length > 0 ? groupColors : data.colors.slice(0, 15),
        );
        setIsLoading(false);
      })
      .catch((err) => {
        setLoadError(err instanceof Error ? err.message : String(err));
        setIsLoading(false);
      });
  }, []);

  const currentMember = members[currentIndex] || members[0];
  const colorMap = new Map<string, Color>(colors.map((c) => [c.id, c]));

  // Handle color selection updates for live preview
  const handleColorSelect = (step: 'left' | 'right', color: Color) => {
    if (step === 'left') {
      setSelectedLeft(color);
      setSelectedRight(undefined);
    } else {
      setSelectedRight(color);
    }
  };

  const handleResetSelection = () => {
    setSelectedLeft(undefined);
    setSelectedRight(undefined);
  };

  // Handle final 2-tap answer
  const handleAnswer = ({
    leftColorId,
    rightColorId,
  }: {
    leftColorId: string;
    rightColorId: string;
  }) => {
    const correctL = currentMember.penlight.left_color_id;
    const correctR = currentMember.penlight.right_color_id;

    // ADR-0019: Inverted orientation accepted (hand switch)
    const isCorrect =
      (leftColorId === correctL && rightColorId === correctR) ||
      (leftColorId === correctR && rightColorId === correctL);

    if (isCorrect) {
      setScore((s) => s + 1);
      setFeedback('correct');
      // Trigger haptic vibration if supported on mobile
      if (typeof window !== 'undefined' && 'vibrate' in navigator) {
        navigator.vibrate(50);
      }
    } else {
      setFeedback('wrong');
    }

    // Advance to next question after brief 1.2s delay for visual feedback
    setTimeout(() => {
      setFeedback('idle');
      setSelectedLeft(undefined);
      setSelectedRight(undefined);

      if (currentIndex + 1 < members.length) {
        setCurrentIndex((i) => i + 1);
      } else {
        setIsFinished(true);
      }
    }, 1200);
  };

  const handleRestart = () => {
    setCurrentIndex(0);
    setScore(0);
    setIsFinished(false);
    setSelectedLeft(undefined);
    setSelectedRight(undefined);
    setFeedback('idle');
  };

  // Render selected layout component
  const renderLayout = () => {
    const props = {
      target: currentMember,
      costumeTitle: '13th Single 制服',
      selectedLeftColor: selectedLeft,
      selectedRightColor: selectedRight,
      isCorrect: feedback === 'correct',
    };

    switch (layoutMode) {
      case 'overlay':
        return <LayoutOverlay {...props} />;
      case 'compact':
        return <LayoutCompact {...props} />;
      case 'classic':
        return <LayoutClassic {...props} />;
      default:
        return <LayoutClassic {...props} />;
    }
  };

  if (isLoading) {
    return (
      <Container
        size="xs"
        p="xl"
        style={{
          minHeight: '100dvh',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 16,
        }}
      >
        <Loader size="lg" color="blue" />
        <Text size="sm" c="dimmed">
          マスターデータを読み込み中...
        </Text>
      </Container>
    );
  }

  if (loadError || members.length === 0) {
    return (
      <Container
        size="xs"
        p="xl"
        style={{
          minHeight: '100dvh',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          gap: 16,
        }}
      >
        <Text size="md" fw={700} c="red.6">
          データの読み込みに失敗しました
        </Text>
        <Text size="xs" c="dimmed">
          {loadError || '有効な出題メンバーが見つかりません'}
        </Text>
        <Button variant="light" onClick={() => window.location.reload()}>
          再試行
        </Button>
      </Container>
    );
  }

  return (
    <Container
      size="xs"
      p={0}
      style={{ minHeight: '100dvh', display: 'flex', flexDirection: 'column' }}
    >
      {/* 共通ヘッダー */}
      <Header
        layoutMode={layoutMode}
        onLayoutModeChange={setLayoutMode}
        onOpenFilter={() => {}}
      />

      {/* 進行プログレスバー */}
      <Box px="md" pt="xs">
        <Box
          mb={4}
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <Text size="xs" fw={700} c="dimmed">
            第 {currentIndex + 1} 問 / 全 {members.length} 問
          </Text>
          <Badge size="xs" variant="outline" color="blue">
            スコア: {score}
          </Badge>
        </Box>
        <Progress
          value={((currentIndex + 1) / members.length) * 100}
          size="xs"
          radius="xl"
          color="blue"
        />
      </Box>

      {/* メイン出題・解答エリア */}
      <Stack
        gap="md"
        align="center"
        justify="space-between"
        px="md"
        py="sm"
        style={{ flexGrow: 1 }}
      >
        {/* 出題カード (選択中レイアウト) */}
        <Box
          style={{ width: '100%', display: 'flex', justifyContent: 'center' }}
        >
          {renderLayout()}
        </Box>

        {/* 正誤判定フィードバックバナー */}
        {feedback !== 'idle' && (
          <Box
            py={6}
            px={16}
            style={{
              borderRadius: 20,
              backgroundColor: feedback === 'correct' ? '#2b8a3e' : '#c92a2a',
              color: '#fff',
              fontWeight: 700,
              fontSize: '13px',
              animation: 'popIn 0.2s ease',
            }}
          >
            {feedback === 'correct'
              ? '🎉 大正解！ (左右順不同OK)'
              : `😢 不正解... 正解: ${colorMap.get(currentMember.penlight.left_color_id)?.name} × ${colorMap.get(currentMember.penlight.right_color_id)?.name}`}
          </Box>
        )}

        {/* 解答インターフェース (新版 15色パレット) */}
        <PaletteGridInput
          target={currentMember}
          colors={colors}
          onAnswer={handleAnswer}
          disabled={feedback !== 'idle'}
          onColorSelect={handleColorSelect}
          onResetSelection={handleResetSelection}
        />
      </Stack>

      {/* 結果発表モーダル */}
      <Modal
        opened={isFinished}
        onClose={handleRestart}
        title="🏆 クイズ結果発表"
        centered
        radius="md"
      >
        <Stack align="center" gap="md" py="md">
          <IconTrophy size={48} color="#fcc419" />
          <Text size="xl" fw={800}>
            {score} / {members.length} 点
          </Text>
          <Text size="sm" c="dimmed">
            正答率: {((score / members.length) * 100).toFixed(1)}%
          </Text>
          <Button
            leftSection={<IconRotateClockwise size={16} />}
            onClick={handleRestart}
            variant="filled"
            color="blue"
            fullWidth
            mt="sm"
          >
            もう一度挑戦する
          </Button>
        </Stack>
      </Modal>
    </Container>
  );
}
