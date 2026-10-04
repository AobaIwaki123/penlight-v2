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
import {
  FilterModal,
  type QuizFilterCriteria,
} from '@/features/quiz/components/FilterModal';
import { Header } from '@/features/quiz/components/Header';
import { InlineFeedbackBar } from '@/features/quiz/components/InlineFeedbackBar';
import { DonutRingModal } from '@/features/quiz/components/inputs/DonutRingModal';
import { PaletteGridInput } from '@/features/quiz/components/inputs/PaletteGridInput';
import { LayoutClassic } from '@/features/quiz/components/layouts/LayoutClassic';
import { LayoutCompact } from '@/features/quiz/components/layouts/LayoutCompact';
import { LayoutOverlay } from '@/features/quiz/components/layouts/LayoutOverlay';
import type { InputMode, LayoutMode } from '@/features/quiz/types';
import type { Color, Group, Member } from '@/types/generated';

function filterAndShuffleMembers(
  sourceMembers: Member[],
  filter: QuizFilterCriteria,
): Member[] {
  const filtered = sourceMembers.filter((m) => {
    if (m.group_id !== filter.groupId) return false;
    if (!filter.includeGraduated && m.status !== 'active') return false;
    if (!m.penlight?.left_color_id || !m.penlight?.right_color_id) return false;
    return filter.generations.includes(m.generation);
  });
  // Shuffle array using Fisher-Yates
  const shuffled = [...filtered];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
}

export function QuizContainer() {
  // Master data from backend API
  const [allGroups, setAllGroups] = useState<Group[]>([]);
  const [allMembers, setAllMembers] = useState<Member[]>([]);
  const [allColors, setAllColors] = useState<Color[]>([]);

  // Active quiz pool and colors
  const [members, setMembers] = useState<Member[]>([]);
  const [colors, setColors] = useState<Color[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  // Settings: presentation layout & input interface
  const [layoutMode, setLayoutMode] = useState<LayoutMode>('overlay');
  const [inputMode, setInputMode] = useState<InputMode>('donut');

  // Filter criteria and modal state
  const [isFilterModalOpen, setIsFilterModalOpen] = useState(false);
  const [filterCriteria, setFilterCriteria] = useState<QuizFilterCriteria>({
    groupId: '',
    generations: [],
    includeGraduated: false,
  });

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
  const [feedback, setFeedback] = useState<'idle' | 'correct' | 'wrong'>(
    'idle',
  );

  useEffect(() => {
    fetchBootstrapData()
      .then((data) => {
        const groups = data.groups || [];
        const rawMembers = data.members || [];
        const rawColors = data.colors || [];

        setAllGroups(groups);
        setAllMembers(rawMembers);
        setAllColors(rawColors);

        // Default to Hinatazaka46 if available, or first group
        const defaultGroup =
          groups.find((g) => g.slug === 'hinatazaka46') || groups[0];
        const defaultGroupId = defaultGroup ? defaultGroup.id : '';

        // Extract generations for default group
        const groupMembers = rawMembers.filter(
          (m) => m.group_id === defaultGroupId,
        );
        const gens = Array.from(
          new Set(groupMembers.map((m) => m.generation)),
        ).sort((a, b) => a - b);

        const initialCriteria: QuizFilterCriteria = {
          groupId: defaultGroupId,
          generations: gens,
          includeGraduated: false,
        };
        setFilterCriteria(initialCriteria);

        const initialMembers = filterAndShuffleMembers(
          rawMembers,
          initialCriteria,
        );
        setMembers(
          initialMembers.length > 0
            ? initialMembers
            : rawMembers.filter(
                (m) =>
                  m.status === 'active' &&
                  m.penlight?.left_color_id &&
                  m.penlight?.right_color_id,
              ),
        );

        const groupColors = rawColors.filter(
          (c) => c.group_id === defaultGroupId,
        );
        setColors(
          groupColors.length > 0 ? groupColors : rawColors.slice(0, 15),
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
  const currentGroup = allGroups.find((g) => g.id === filterCriteria.groupId);

  // Handle filter submission: rebuild deck and reset quiz progress
  const handleApplyFilter = (newCriteria: QuizFilterCriteria) => {
    setFilterCriteria(newCriteria);
    const filtered = filterAndShuffleMembers(allMembers, newCriteria);
    setMembers(filtered);

    const groupColors = allColors.filter(
      (c) => c.group_id === newCriteria.groupId,
    );
    setColors(groupColors.length > 0 ? groupColors : allColors.slice(0, 15));

    setCurrentIndex(0);
    setScore(0);
    setIsFinished(false);
    setSelectedLeft(undefined);
    setSelectedRight(undefined);
    setFeedback('idle');
  };

  // Handle color selection updates for live preview
  const handleColorSelect = (step: 'left' | 'right', color: Color) => {
    if (step === 'left') {
      setSelectedLeft(color);
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
      if (typeof window !== 'undefined' && 'vibrate' in navigator) {
        navigator.vibrate(50);
      }
    } else {
      setFeedback('wrong');
    }
  };

  // Move to next question (次へ押下でペンライト色をリセット)
  const handleNextQuestion = () => {
    setFeedback('idle');
    setSelectedLeft(undefined);
    setSelectedRight(undefined);
    setActiveHand('left');

    if (currentIndex + 1 < members.length) {
      setCurrentIndex((i) => i + 1);
    } else {
      setIsFinished(true);
    }
  };

  const handleRestart = () => {
    // Reshuffle current member pool for next round
    setMembers((prev) => {
      const reshuffled = [...prev];
      for (let i = reshuffled.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1));
        [reshuffled[i], reshuffled[j]] = [reshuffled[j], reshuffled[i]];
      }
      return reshuffled;
    });
    setCurrentIndex(0);
    setScore(0);
    setIsFinished(false);
    setSelectedLeft(undefined);
    setSelectedRight(undefined);
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
      isFullscreen: fillScreen,
    };

    switch (layoutMode) {
      case 'overlay':
        return (
          <LayoutOverlay
            {...props}
            footer={feedback !== 'idle' ? feedbackBar : undefined}
          />
        );
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

  // 写真を下端まで敷き詰める全画面構成 (オーバーレイ＋ドーナツ)
  const fillScreen = layoutMode === 'overlay' && inputMode === 'donut';

  const feedbackBar = (
    <InlineFeedbackBar
      isCorrect={feedback === 'correct'}
      correctLeftColor={colorMap.get(currentMember.penlight.left_color_id)}
      correctRightColor={colorMap.get(currentMember.penlight.right_color_id)}
      onNext={handleNextQuestion}
    />
  );

  return (
    <Container
      size="xs"
      p={0}
      style={{
        // 全画面レイアウト時は画面高にぴったり収め、写真を下端まで敷き詰める
        ...(fillScreen ? { height: '100dvh' } : { minHeight: '100dvh' }),
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* 共通ヘッダー */}
      <Header
        layoutMode={layoutMode}
        onLayoutModeChange={setLayoutMode}
        inputMode={inputMode}
        onInputModeChange={setInputMode}
        onOpenFilter={() => setIsFilterModalOpen(true)}
        groupThemeColor={currentGroup?.theme_color_hex}
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
        pt="sm"
        pb="sm"
        style={{ flexGrow: 1, minHeight: 0 }}
      >
        {/* 出題カード (選択中レイアウト) */}
        <Box
          style={{
            width: '100%',
            display: 'flex',
            justifyContent: 'center',
            flex: fillScreen ? 1 : undefined,
            minHeight: fillScreen ? 0 : undefined,
          }}
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
        {feedback !== 'idle' && !fillScreen && feedbackBar}
      </Stack>

      {/* ドーナツリングカラー選択モーダル (完全透過 ＆ 2本のミニペンライトで左右を視覚化) */}
      <DonutRingModal
        opened={isDonutModalOpen && feedback === 'idle'}
        onClose={() => setIsDonutModalOpen(false)}
        colors={colors}
        selectedLeftColor={selectedLeft}
        selectedRightColor={selectedRight}
        onAnswer={handleAnswer}
        onColorSelect={handleColorSelect}
        disabled={feedback !== 'idle'}
        initialHand={activeHand}
      />

      {/* 絞り込みフィルターモーダル */}
      <FilterModal
        opened={isFilterModalOpen}
        onClose={() => setIsFilterModalOpen(false)}
        groups={allGroups}
        allMembers={allMembers}
        currentFilter={filterCriteria}
        onApply={handleApplyFilter}
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
