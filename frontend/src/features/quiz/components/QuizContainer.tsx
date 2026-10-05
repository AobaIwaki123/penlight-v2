'use client';

import {
  Badge,
  Box,
  Button,
  Container,
  Loader,
  Group as MantineGroup,
  Modal,
  Progress,
  Stack,
  Text,
} from '@mantine/core';
import { IconHome, IconRotateClockwise, IconTrophy } from '@tabler/icons-react';
import { useEffect, useState } from 'react';
import { PortalView } from '@/features/portal/components/PortalView';
import { fetchBootstrapData } from '@/features/quiz/api/client';
import { Header } from '@/features/quiz/components/Header';
import { InlineFeedbackBar } from '@/features/quiz/components/InlineFeedbackBar';
import { DonutRingModal } from '@/features/quiz/components/inputs/DonutRingModal';
import { PaletteGridInput } from '@/features/quiz/components/inputs/PaletteGridInput';
import { LayoutClassic } from '@/features/quiz/components/layouts/LayoutClassic';
import { LayoutCompact } from '@/features/quiz/components/layouts/LayoutCompact';
import { LayoutOverlay } from '@/features/quiz/components/layouts/LayoutOverlay';
import { SongQuizArea } from '@/features/quiz/components/SongQuizArea';
import type { InputMode, LayoutMode } from '@/features/quiz/types';
import {
  loadSavedInputMode,
  loadSavedLayoutMode,
  loadSavedSettings,
  saveInputMode,
  saveLayoutMode,
  saveSettings,
} from '@/features/quiz/utils/storage';
import type { Color, Group, Member, Series, Song } from '@/types/generated';

function filterAndShuffleMembers(
  sourceMembers: Member[],
  groupId: string,
): Member[] {
  const filtered = sourceMembers.filter(
    (m) =>
      m.group_id === groupId &&
      m.status === 'active' &&
      m.penlight?.left_color_id &&
      m.penlight?.right_color_id,
  );
  const shuffled = [...filtered];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
}

function filterAndShuffleSongs(sourceSongs: Song[], groupId: string): Song[] {
  const filtered = sourceSongs.filter((s) => s.group_id === groupId);
  const shuffled = [...filtered];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }
  return shuffled;
}

function getRelevantColors(
  allColors: Color[],
  groupId: string,
  seriesId: string,
  groups: Group[],
): Color[] {
  const groupColors = allColors.filter((c) => c.group_id === groupId);
  if (groupColors.length > 0) return groupColors;

  const seriesGroupIds = new Set(
    groups.filter((g) => g.series_id === seriesId).map((g) => g.id),
  );
  const seriesColors = allColors.filter(
    (c) => c.group_id && seriesGroupIds.has(c.group_id),
  );
  if (seriesColors.length > 0) return seriesColors;

  return allColors.slice(0, 15);
}

export function QuizContainer() {
  // Master data from backend API
  const [allSeries, setAllSeries] = useState<Series[]>([]);
  const [allGroups, setAllGroups] = useState<Group[]>([]);
  const [allMembers, setAllMembers] = useState<Member[]>([]);
  const [allSongs, setAllSongs] = useState<Song[]>([]);
  const [allColors, setAllColors] = useState<Color[]>([]);

  // Selected quiz criteria
  const [selectedGroupId, setSelectedGroupId] = useState<string>('');
  const [songMode, setSongMode] = useState<boolean>(false);

  // Active quiz pool and colors
  const [members, setMembers] = useState<Member[]>([]);
  const [songs, setSongs] = useState<Song[]>([]);
  const [colors, setColors] = useState<Color[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [loadError, setLoadError] = useState<string | null>(null);

  // View state: 'portal' or 'quiz'
  const [viewMode, setViewMode] = useState<'portal' | 'quiz'>('portal');

  // Presentation layout & input interface
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
  const [feedback, setFeedback] = useState<'idle' | 'correct' | 'wrong'>(
    'idle',
  );

  useEffect(() => {
    // Restore saved layout and input mode
    const savedLayout = loadSavedLayoutMode();
    if (savedLayout) setLayoutMode(savedLayout);
    const savedInput = loadSavedInputMode();
    if (savedInput) setInputMode(savedInput);

    fetchBootstrapData()
      .then((data) => {
        const rawSeries = data.series || [];
        const rawGroups = data.groups || [];
        const rawMembers = data.members || [];
        const rawSongs = data.songs || [];
        const rawColors = data.colors || [];

        setAllSeries(rawSeries);
        setAllGroups(rawGroups);
        setAllMembers(rawMembers);
        setAllSongs(rawSongs);
        setAllColors(rawColors);

        // Load saved settings or fallback to Hinatazaka46
        const saved = loadSavedSettings();
        const defaultGroup =
          rawGroups.find((g) => g.id === saved?.groupId) ||
          rawGroups.find((g) => g.slug === 'hinatazaka46') ||
          rawGroups[0];
        const defaultGroupId = defaultGroup ? defaultGroup.id : '';

        setSelectedGroupId(defaultGroupId);
        setSongMode(saved?.songMode ?? false);

        if (defaultGroup) {
          const relevantColors = getRelevantColors(
            rawColors,
            defaultGroupId,
            defaultGroup.series_id,
            rawGroups,
          );
          setColors(relevantColors);
        }

        setIsLoading(false);
      })
      .catch((err) => {
        setLoadError(err instanceof Error ? err.message : String(err));
        setIsLoading(false);
      });
  }, []);

  const currentGroup = allGroups.find((g) => g.id === selectedGroupId);
  const currentSeries = allSeries.find((s) => s.id === currentGroup?.series_id);
  const colorMap = new Map<string, Color>(allColors.map((c) => [c.id, c]));

  // Total questions count depending on mode
  const totalQuestions = songMode ? songs.length : members.length;
  const currentMember = members[currentIndex] || members[0];
  const currentSong = songs[currentIndex] || songs[0];

  // Change group from Portal
  const handleGroupChange = (groupId: string) => {
    setSelectedGroupId(groupId);
    saveSettings({ groupId, songMode });

    const targetGroup = allGroups.find((g) => g.id === groupId);
    if (targetGroup) {
      setColors(
        getRelevantColors(allColors, groupId, targetGroup.series_id, allGroups),
      );
    }
  };

  // Change songMode from Portal
  const handleSongModeChange = (newSongMode: boolean) => {
    setSongMode(newSongMode);
    saveSettings({ groupId: selectedGroupId, songMode: newSongMode });
  };

  // Start quiz session directly with current group and mode
  const startQuizSession = () => {
    if (songMode) {
      const shuffledSongs = filterAndShuffleSongs(allSongs, selectedGroupId);
      setSongs(shuffledSongs);
    } else {
      const shuffledMembers = filterAndShuffleMembers(
        allMembers,
        selectedGroupId,
      );
      setMembers(shuffledMembers);
    }

    setCurrentIndex(0);
    setScore(0);
    setIsFinished(false);
    setSelectedLeft(undefined);
    setSelectedRight(undefined);
    setFeedback('idle');
    setViewMode('quiz');
  };

  const handleReturnToPortal = () => {
    setViewMode('portal');
    setIsFinished(false);
    setFeedback('idle');
    setSelectedLeft(undefined);
    setSelectedRight(undefined);
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

  // Handle final answer (both Member & Song)
  const handleAnswer = ({
    leftColorId,
    rightColorId,
  }: {
    leftColorId: string;
    rightColorId: string;
  }) => {
    let isCorrect = false;

    if (songMode) {
      if (!currentSong) return;
      const c1 = currentSong.color1_id;
      const c2 = currentSong.color2_id;

      if (!c2) {
        // 1-color song: exact match with leftColorId
        isCorrect = leftColorId === c1;
      } else {
        // 2-color song: set-equality (unordered)
        isCorrect =
          (leftColorId === c1 && rightColorId === c2) ||
          (leftColorId === c2 && rightColorId === c1);
      }
    } else {
      if (!currentMember) return;
      const correctL = currentMember.penlight?.left_color_id;
      const correctR = currentMember.penlight?.right_color_id;

      isCorrect =
        (leftColorId === correctL && rightColorId === correctR) ||
        (leftColorId === correctR && rightColorId === correctL);
    }

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

  // Move to next question
  const handleNextQuestion = () => {
    setFeedback('idle');
    setSelectedLeft(undefined);
    setSelectedRight(undefined);
    setActiveHand('left');

    if (currentIndex + 1 < totalQuestions) {
      setCurrentIndex((i) => i + 1);
    } else {
      setIsFinished(true);
    }
  };

  const handleRestart = () => {
    startQuizSession();
  };

  const handleLayoutModeChange = (mode: LayoutMode) => {
    setLayoutMode(mode);
    saveLayoutMode(mode);
  };

  const handleInputModeChange = (mode: InputMode) => {
    setInputMode(mode);
    saveInputMode(mode);
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

  if (loadError) {
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
          {loadError}
        </Text>
        <Button variant="light" onClick={() => window.location.reload()}>
          再試行
        </Button>
      </Container>
    );
  }

  // Render Portal View when viewMode is 'portal' (Simple & Clean)
  if (viewMode === 'portal') {
    return (
      <PortalView
        groups={allGroups}
        members={allMembers}
        songs={allSongs}
        colors={allColors}
        selectedGroupId={selectedGroupId}
        songMode={songMode}
        onGroupChange={handleGroupChange}
        onSongModeChange={handleSongModeChange}
        onStartQuiz={startQuizSession}
      />
    );
  }

  // Quiz View
  const fillScreen =
    !songMode && layoutMode === 'overlay' && inputMode === 'donut';

  const requiredColorsCount: 1 | 2 =
    songMode && currentSong ? (currentSong.color2_id ? 2 : 1) : 2;

  const correctLeft = songMode
    ? colorMap.get(currentSong?.color1_id || '')
    : colorMap.get(currentMember?.penlight?.left_color_id || '');

  const correctRight = songMode
    ? currentSong?.color2_id
      ? colorMap.get(currentSong.color2_id)
      : undefined
    : colorMap.get(currentMember?.penlight?.right_color_id || '');

  const feedbackBar = (
    <InlineFeedbackBar
      isCorrect={feedback === 'correct'}
      correctLeftColor={correctLeft}
      correctRightColor={correctRight}
      onNext={handleNextQuestion}
    />
  );

  // Render member layout
  const renderMemberLayout = () => {
    const handleOpenInput = (hand: 'left' | 'right') => {
      if (feedback !== 'idle') return;
      setActiveHand(hand);
      setIsDonutModalOpen(true);
    };

    const props = {
      target: currentMember,
      costumeTitle: currentMember?.images?.[0]?.photo_type?.name || '公式衣装',
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

  return (
    <Container
      size="xs"
      p={0}
      style={{
        ...(fillScreen ? { height: '100dvh' } : { minHeight: '100dvh' }),
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* 共通ヘッダー (ポータルへ戻るボタン・グループ名表示) */}
      <Header
        layoutMode={layoutMode}
        onLayoutModeChange={handleLayoutModeChange}
        inputMode={inputMode}
        onInputModeChange={handleInputModeChange}
        onGoHome={handleReturnToPortal}
        groupThemeColor={currentGroup?.theme_color_hex}
        seriesName={currentSeries?.name}
        groupName={currentGroup?.name}
        isSongMode={songMode}
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
            第 {currentIndex + 1} 問 / 全 {totalQuestions} 問
          </Text>
          <Badge size="xs" variant="outline" color="blue">
            スコア: {score}
          </Badge>
        </Box>
        <Progress
          value={
            totalQuestions > 0 ? ((currentIndex + 1) / totalQuestions) * 100 : 0
          }
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
        {/* 出題カード (楽曲クイズ vs メンバークイズ) */}
        <Box
          style={{
            width: '100%',
            display: 'flex',
            justifyContent: 'center',
            flex: fillScreen ? 1 : undefined,
            minHeight: fillScreen ? 0 : undefined,
          }}
        >
          {songMode ? (
            currentSong ? (
              <SongQuizArea
                song={currentSong}
                group={currentGroup}
                selectedColor1={selectedLeft}
                selectedColor2={selectedRight}
                isCorrect={feedback === 'correct'}
                onOpenInput={
                  feedback === 'idle'
                    ? (slot) => {
                        setActiveHand(slot === 1 ? 'left' : 'right');
                        setIsDonutModalOpen(true);
                      }
                    : undefined
                }
                footer={feedback !== 'idle' ? feedbackBar : undefined}
              />
            ) : (
              <Text c="dimmed">出題可能な楽曲がありません</Text>
            )
          ) : (
            currentMember && renderMemberLayout()
          )}
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
            requiredColorsCount={requiredColorsCount}
          />
        )}

        {/* 解答直後のインラインフィードバックバー (通常モード) */}
        {feedback !== 'idle' && !fillScreen && !songMode && feedbackBar}
      </Stack>

      {/* ドーナツリングカラー選択モーダル */}
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
        requiredColorsCount={requiredColorsCount}
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
            {score} / {totalQuestions} 点
          </Text>
          <Text size="sm" c="dimmed">
            正答率:{' '}
            {totalQuestions > 0
              ? ((score / totalQuestions) * 100).toFixed(1)
              : 0}
            %
          </Text>
          <MantineGroup gap="sm" style={{ width: '100%' }}>
            <Button
              leftSection={<IconRotateClockwise size={16} />}
              onClick={handleRestart}
              variant="filled"
              color="blue"
              style={{ flex: 1 }}
            >
              もう一度挑戦
            </Button>
            <Button
              leftSection={<IconHome size={16} />}
              onClick={handleReturnToPortal}
              variant="light"
              color="gray"
              style={{ flex: 1 }}
            >
              トップへ戻る
            </Button>
          </MantineGroup>
        </Stack>
      </Modal>
    </Container>
  );
}
