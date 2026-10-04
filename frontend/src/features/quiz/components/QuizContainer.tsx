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
import { InlineFeedbackBar } from '@/features/quiz/components/InlineFeedbackBar';
import { DonutRingModal } from '@/features/quiz/components/inputs/DonutRingModal';
import { PaletteGridInput } from '@/features/quiz/components/inputs/PaletteGridInput';
import { LayoutClassic } from '@/features/quiz/components/layouts/LayoutClassic';
import { LayoutCompact } from '@/features/quiz/components/layouts/LayoutCompact';
import { LayoutOverlay } from '@/features/quiz/components/layouts/LayoutOverlay';
import type { InputMode, LayoutMode } from '@/features/quiz/types';
import type { Color, Member } from '@/types/generated';

export function QuizContainer() {
  // Master data from backend API
  const [members, setMembers] = useState<Member[]>([]);
  const [colors, setColors] = useState<Color[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  // Settings: presentation layout & input interface
  const [layoutMode, setLayoutMode] = useState<LayoutMode>('overlay');
  const [inputMode, setInputMode] = useState<InputMode>('donut');

  // Modal state for Donut Ring Input
  const [isDonutModalOpen, setIsDonutModalOpen] = useState(false);
  const [activeHand, setActiveHand] = useState<'left' | 'right'>('left');

  // Session state
  const [currentIndex, setCurrentIndex] = useState(0);
  const [score, setScore] = useState(0);
  const [isFinished, setIsFinished] = useState(false);

  // Current question selection state
  const [selectedLeft, setSelectedLeft] = useState<Color | undefined>();
  const [selectedRight, setSelectedRight] = useState<Color | undefined>();
  const [confirmed, setConfirmed] = useState({ left: false, right: false });
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
  // 選択色は次の問題でも保持し、各問題で左右とも選び直した時だけ送信対象とする
  const handleColorSelect = (step: 'left' | 'right', color: Color) => {
    if (step === 'left') {
      setSelectedLeft(color);
    } else {
      setSelectedRight(color);
    }
    setConfirmed((c) => ({ ...c, [step]: true }));
  };

  const handleResetSelection = () => {
    setSelectedLeft(undefined);
    setSelectedRight(undefined);
    setConfirmed({ left: false, right: false });
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
      if (typeof window !== 'undefined' && 'vibrate' in navigator) {
        navigator.vibrate(50);
      }
    } else {
      // 不正解でも選択したペンライト色は保持する (正解はフィードバックバーに表示)
      setFeedback('wrong');
    }
  };

  // Move to next question (選択色は保持、確定状態のみリセット)
  const handleNextQuestion = () => {
    setFeedback('idle');
    setConfirmed({ left: false, right: false });
    setActiveHand('left');

    if (currentIndex + 1 < members.length) {
      setCurrentIndex((i) => i + 1);
    } else {
      setIsFinished(true);
    }
  };

  const handleRestart = () => {
    setCurrentIndex(0);
    setScore(0);
    setIsFinished(false);
    setSelectedLeft(undefined);
    setSelectedRight(undefined);
    setConfirmed({ left: false, right: false });
    setFeedback('idle');
  };

  // Render selected layout component
  const renderLayout = () => {
    const handleOpenInput = (hand: 'left' | 'right') => {
      // 「次へ」表示中 (解答済み) はカラーピッカーを開けない
      if (feedback !== 'idle') return;
      setActiveHand(hand);
      setIsDonutModalOpen(true);
    };

    const props = {
      target: currentMember,
      costumeTitle: '13th Single 制服',
      selectedLeftColor: selectedLeft,
      selectedRightColor: selectedRight,
      isCorrect: feedback === 'correct',
      onOpenInput: handleOpenInput,
      isFullscreen: inputMode === 'donut',
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
        inputMode={inputMode}
        onInputModeChange={setInputMode}
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

        {/* 解答インターフェース: グリッド選択時のみ下部に常時表示 */}
        {inputMode === 'grid' && feedback === 'idle' && (
          <PaletteGridInput
            target={currentMember}
            colors={colors}
            onAnswer={handleAnswer}
            disabled={feedback !== 'idle'}
            onColorSelect={handleColorSelect}
            onResetSelection={handleResetSelection}
          />
        )}

        {/* 解答直後の 1行インラインフィードバックバー (中央を邪魔せず最下部に表示) */}
        {feedback !== 'idle' && (
          <InlineFeedbackBar
            isCorrect={feedback === 'correct'}
            correctLeftColor={colorMap.get(
              currentMember.penlight.left_color_id,
            )}
            correctRightColor={colorMap.get(
              currentMember.penlight.right_color_id,
            )}
            onNext={handleNextQuestion}
          />
        )}
      </Stack>

      {/* ドーナツリングカラー選択モーダル (完全透過 ＆ 2本のミニペンライトで左右を視覚化) */}
      <DonutRingModal
        opened={isDonutModalOpen && feedback === 'idle'}
        onClose={() => setIsDonutModalOpen(false)}
        colors={colors}
        selectedLeftColor={selectedLeft}
        selectedRightColor={selectedRight}
        leftConfirmed={confirmed.left}
        rightConfirmed={confirmed.right}
        onAnswer={handleAnswer}
        onColorSelect={handleColorSelect}
        disabled={feedback !== 'idle'}
        initialHand={activeHand}
      />

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
