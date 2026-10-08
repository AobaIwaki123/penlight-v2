'use client';

import {
  Alert,
  Badge,
  Box,
  Button,
  Container,
  Group,
  Modal,
  NumberInput,
  Paper,
  SegmentedControl,
  Stack,
  Table,
  Text,
  TextInput,
} from '@mantine/core';
import {
  IconAlertCircle,
  IconCheck,
  IconChevronLeft,
  IconChevronRight,
  IconClock,
  IconFilter,
  IconPencil,
  IconSearch,
  IconSend,
} from '@tabler/icons-react';
import { useRouter } from 'next/navigation';
import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  fetchBootstrapData,
  submitMetadataEditProposal,
} from '@/features/quiz/api/client';
import {
  FilterModal,
  type QuizFilterCriteria,
} from '@/features/quiz/components/FilterModal';
import { Header } from '@/features/quiz/components/Header';
import { DonutRingModal } from '@/features/quiz/components/inputs/DonutRingModal';
import { LayoutOverlay } from '@/features/quiz/components/layouts/LayoutOverlay';
import { generateMetadataProposalID } from '@/features/quiz/utils/id';
import type {
  BootstrapResponse,
  Member,
  MemberStatus,
  MetadataEditChanges,
  MetadataEditProposal,
  PenlightPair,
} from '@/types/generated';

interface DraftState {
  penlight: PenlightPair;
  generation: number;
  status: MemberStatus;
}

interface SubmissionState {
  id: string;
  bodyKey: string;
  result?: MetadataEditProposal;
  error?: string;
}

const statusLabels: Record<MemberStatus, string> = {
  active: '現役',
  graduated: '卒業',
  hiatus: '休業中',
};

function createDraft(member: Member): DraftState {
  return {
    penlight: { ...member.penlight },
    generation: member.generation,
    status: member.status,
  };
}

function buildChanges(member: Member, draft: DraftState): MetadataEditChanges {
  const changes: MetadataEditChanges = {};

  if (
    member.penlight.left_color_id !== draft.penlight.left_color_id ||
    member.penlight.right_color_id !== draft.penlight.right_color_id ||
    member.penlight.ordered !== draft.penlight.ordered
  ) {
    changes.penlight = {
      before: member.penlight,
      after: draft.penlight,
    };
  }

  if (member.generation !== draft.generation) {
    changes.generation = {
      before: member.generation,
      after: draft.generation,
    };
  }

  if (member.status !== draft.status) {
    changes.status = {
      before: member.status,
      after: draft.status,
    };
  }

  return changes;
}

function changeCount(changes: MetadataEditChanges): number {
  return (
    Number(Boolean(changes.penlight)) +
    Number(Boolean(changes.generation)) +
    Number(Boolean(changes.status))
  );
}

export function MetadataProposalPage() {
  const router = useRouter();
  const [bootstrap, setBootstrap] = useState<BootstrapResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  // フィルター条件 (FilterModal と連動)
  const [filterCriteria, setFilterCriteria] = useState<QuizFilterCriteria>({
    groupId: '',
    generations: [],
    includeGraduated: true,
    songMode: false,
  });

  const [isFilterModalOpen, setIsFilterModalOpen] = useState(false);
  const [currentIndex, setCurrentIndex] = useState(0);

  // 編集下書き & 提案送信状態 (メンバーIDキー)
  const [drafts, setDrafts] = useState<Record<string, DraftState>>({});
  const [submissions, setSubmissions] = useState<
    Record<string, SubmissionState>
  >({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isOnline, setIsOnline] = useState(true);

  // モーダルの開閉
  const [isDonutModalOpen, setIsDonutModalOpen] = useState(false);
  const [activeHand, setActiveHand] = useState<'left' | 'right'>('left');
  const [isInfoModalOpen, setIsInfoModalOpen] = useState(false);
  const [isDiffModalOpen, setIsDiffModalOpen] = useState(false);
  const [isPickerModalOpen, setIsPickerModalOpen] = useState(false);
  const [memberSearchQuery, setMemberSearchQuery] = useState('');

  // オンライン状態検知
  useEffect(() => {
    const updateOnline = () => setIsOnline(navigator.onLine);
    updateOnline();
    window.addEventListener('online', updateOnline);
    window.addEventListener('offline', updateOnline);
    return () => {
      window.removeEventListener('online', updateOnline);
      window.removeEventListener('offline', updateOnline);
    };
  }, []);

  // 初期データ取得 (卒業生も含めて取得)
  useEffect(() => {
    let mounted = true;
    fetchBootstrapData({ includeGraduated: true })
      .then((data) => {
        if (!mounted) return;
        const searchParams = new URLSearchParams(window.location.search);
        const requestedMemberId = searchParams.get('member_id');
        const targetMember = requestedMemberId
          ? data.members.find((m) => m.id === requestedMemberId)
          : null;

        const requestedGroup = targetMember
          ? targetMember.group_id
          : searchParams.get('group_id');

        const defaultGroupId =
          data.groups.find((g) => g.id === requestedGroup)?.id ||
          data.groups[0]?.id ||
          '';

        setFilterCriteria({
          groupId: defaultGroupId,
          generations: [],
          includeGraduated: true,
          songMode: false,
        });

        // 対象メンバーが指定されていた場合、該当メンバーのインデックスをセット
        if (targetMember) {
          const groupMembers = data.members.filter(
            (m) => m.group_id === defaultGroupId,
          );
          const memberIndex = groupMembers.findIndex(
            (m) => m.id === targetMember.id,
          );
          if (memberIndex >= 0) {
            setCurrentIndex(memberIndex);
          }
        }

        setBootstrap(data);
      })
      .catch((err) => {
        if (mounted) setError(err instanceof Error ? err.message : String(err));
      });

    return () => {
      mounted = false;
    };
  }, []);

  // フィルター適用後のメンバー一覧
  const filteredMembers = useMemo(() => {
    if (!bootstrap) return [];
    return bootstrap.members.filter((m) => {
      if (m.group_id !== filterCriteria.groupId) return false;
      if (!filterCriteria.includeGraduated && m.status === 'graduated')
        return false;
      if (
        filterCriteria.generations.length > 0 &&
        !filterCriteria.generations.includes(m.generation)
      ) {
        return false;
      }
      return true;
    });
  }, [bootstrap, filterCriteria]);

  const currentMember = filteredMembers[currentIndex];
  const currentGroup = bootstrap?.groups.find(
    (g) => g.id === filterCriteria.groupId,
  );

  // カラーマップ
  const colorMap = useMemo(() => {
    return new Map((bootstrap?.colors || []).map((c) => [c.id, c]));
  }, [bootstrap]);

  // 現在メンバーのドラフト
  const currentDraft = useMemo(() => {
    if (!currentMember) return null;
    return drafts[currentMember.id] || createDraft(currentMember);
  }, [currentMember, drafts]);

  const updateCurrentDraft = useCallback(
    (update: Partial<DraftState>) => {
      if (!currentMember || !currentDraft) return;
      setDrafts((prev) => ({
        ...prev,
        [currentMember.id]: {
          ...currentDraft,
          ...update,
        },
      }));
    },
    [currentMember, currentDraft],
  );

  // 変更差分と送信ステータス
  const changes = useMemo(() => {
    if (!currentMember || !currentDraft) return {};
    return buildChanges(currentMember, currentDraft);
  }, [currentMember, currentDraft]);

  const changedCount = changeCount(changes);

  const bodyKey = useMemo(() => {
    if (!currentMember) return '';
    return JSON.stringify({
      base_revision: currentMember.metadata_revision,
      changes,
    });
  }, [currentMember, changes]);

  const currentSubmission = useMemo(() => {
    if (!currentMember) return undefined;
    const sub = submissions[currentMember.id];
    return sub?.bodyKey === bodyKey ? sub : undefined;
  }, [currentMember, submissions, bodyKey]);

  const isProposalSubmitted = Boolean(currentSubmission?.result);

  // 左右の色
  const selectedLeftColor = currentDraft
    ? colorMap.get(currentDraft.penlight.left_color_id)
    : undefined;
  const selectedRightColor = currentDraft
    ? colorMap.get(currentDraft.penlight.right_color_id)
    : undefined;

  // 写真 & 衣装名 (表示専用)
  const primaryImage =
    currentMember?.images?.find((img) => img.is_primary) ||
    currentMember?.images?.[0];
  const currentCostumeTitle = primaryImage?.photo_type?.name || '公式衣装';

  // 利用可能なカラー
  const availableColors = useMemo(() => {
    if (!bootstrap || !currentMember) return [];
    return bootstrap.colors
      .filter((c) => !c.group_id || c.group_id === currentMember.group_id)
      .sort((a, b) => a.display_order - b.display_order);
  }, [bootstrap, currentMember]);

  // ナビゲーション
  const handleNavigate = useCallback(
    (delta: number) => {
      setCurrentIndex((prev) => {
        const next = prev + delta;
        return Math.max(0, Math.min(filteredMembers.length - 1, next));
      });
    },
    [filteredMembers.length],
  );

  // キーボードナビゲーション (モーダル非表示時のみ)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (
        isDonutModalOpen ||
        isInfoModalOpen ||
        isDiffModalOpen ||
        isFilterModalOpen ||
        isPickerModalOpen
      ) {
        return;
      }
      if (e.key === 'ArrowLeft') {
        handleNavigate(-1);
      } else if (e.key === 'ArrowRight') {
        handleNavigate(1);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [
    handleNavigate,
    isDonutModalOpen,
    isInfoModalOpen,
    isDiffModalOpen,
    isFilterModalOpen,
    isPickerModalOpen,
  ]);

  // 提案送信
  const handleSubmitProposal = async () => {
    if (
      !currentMember ||
      changedCount === 0 ||
      !isOnline ||
      isSubmitting ||
      isProposalSubmitted
    ) {
      return;
    }

    const attemptId = currentSubmission?.id || generateMetadataProposalID();
    const attempt: SubmissionState = {
      id: attemptId,
      bodyKey,
    };

    setSubmissions((prev) => ({ ...prev, [currentMember.id]: attempt }));
    setIsSubmitting(true);

    try {
      const result = await submitMetadataEditProposal(currentMember.id, {
        id: attemptId,
        base_revision: currentMember.metadata_revision,
        changes,
      });
      setSubmissions((prev) => ({
        ...prev,
        [currentMember.id]: { ...attempt, result },
      }));
      setIsDiffModalOpen(false);
    } catch (err) {
      setSubmissions((prev) => ({
        ...prev,
        [currentMember.id]: {
          ...attempt,
          error: err instanceof Error ? err.message : String(err),
        },
      }));
    } finally {
      setIsSubmitting(false);
    }
  };

  // 公式の回答に戻す
  const handleResetDraft = () => {
    if (!currentMember) return;
    setDrafts((prev) => ({
      ...prev,
      [currentMember.id]: createDraft(currentMember),
    }));
  };

  if (error) {
    return (
      <Container size="xs" py="xl">
        <Alert
          color="red"
          title="読み込みエラー"
          icon={<IconAlertCircle size={16} />}
        >
          {error}
        </Alert>
        <Button mt="md" onClick={() => router.push('/')} variant="default">
          トップへ戻る
        </Button>
      </Container>
    );
  }

  if (!bootstrap) {
    return (
      <Container size="xs" py="xl" ta="center">
        <Text size="sm" c="dimmed">
          マスターデータを読み込んでいます…
        </Text>
      </Container>
    );
  }

  return (
    <Container
      size="xs"
      p={0}
      style={{
        height: '100dvh',
        minHeight: 620,
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* 1. 既存 Header の完全踏襲 (戻る / グループ名 / フィルター) */}
      <Header
        layoutMode="overlay"
        onLayoutModeChange={() => {}}
        inputMode="donut"
        onInputModeChange={() => {}}
        onGoHome={() => router.push('/')}
        groupName={currentGroup?.name}
        groupThemeColor={currentGroup?.theme_color_hex}
        onOpenFilter={() => setIsFilterModalOpen(true)}
      />

      {/* 2. メインカードエリア (クイズ出題画面と同一の世界観) */}
      <Stack
        gap="xs"
        align="center"
        justify="space-between"
        px="md"
        pt="xs"
        pb="xs"
        style={{ flex: 1, minHeight: 0 }}
      >
        {currentMember && currentDraft ? (
          <Box
            style={{
              width: '100%',
              display: 'flex',
              justifyContent: 'center',
              flex: 1,
              minHeight: 0,
            }}
          >
            <LayoutOverlay
              target={{
                ...currentMember,
                generation: currentDraft.generation,
              }}
              costumeTitle={currentCostumeTitle}
              selectedLeftColor={selectedLeftColor}
              selectedRightColor={selectedRightColor}
              isFullscreen
              onOpenInput={(hand) => {
                setActiveHand(hand);
                setIsDonutModalOpen(true);
              }}
              onClickGeneration={() => setIsInfoModalOpen(true)}
              extraBadges={
                <Group gap={6}>
                  {/* 在籍ステータスバッジ (常にタップ可能で、変更時は目立つように表示) */}
                  {(() => {
                    const isStatusChanged =
                      currentDraft.status !== currentMember.status;
                    const isGraduated = currentDraft.status === 'graduated';

                    if (isStatusChanged) {
                      return (
                        <Badge
                          size="sm"
                          color={isGraduated ? 'red' : 'teal'}
                          variant="filled"
                          rightSection={
                            <IconPencil size={10} style={{ marginLeft: 2 }} />
                          }
                          style={{
                            cursor: 'pointer',
                            boxShadow: '0 0 0 1.5px #fff',
                            fontWeight: 800,
                          }}
                          onClick={() => setIsInfoModalOpen(true)}
                          title="タップしてステータスを変更"
                        >
                          {isGraduated ? '卒業へ変更' : '現役へ変更'}
                        </Badge>
                      );
                    }

                    if (isGraduated) {
                      return (
                        <Badge
                          size="sm"
                          color="red.9"
                          variant="filled"
                          rightSection={
                            <IconPencil size={10} style={{ marginLeft: 2 }} />
                          }
                          style={{
                            cursor: 'pointer',
                            boxShadow: '0 0 0 1px rgba(255, 255, 255, 0.4)',
                          }}
                          onClick={() => setIsInfoModalOpen(true)}
                          title="タップしてステータスを変更"
                        >
                          卒業
                        </Badge>
                      );
                    }

                    return (
                      <Badge
                        size="sm"
                        color="gray.4"
                        variant="outline"
                        rightSection={
                          <IconPencil size={10} style={{ marginLeft: 2 }} />
                        }
                        style={{
                          cursor: 'pointer',
                          color: '#fff',
                          borderColor: 'rgba(255, 255, 255, 0.5)',
                        }}
                        onClick={() => setIsInfoModalOpen(true)}
                        title="タップしてステータスを変更"
                      >
                        現役
                      </Badge>
                    );
                  })()}

                  {/* 期生が変更された場合の強調バッジ */}
                  {currentDraft.generation !== currentMember.generation && (
                    <Badge size="sm" color="violet" variant="filled">
                      期生変更
                    </Badge>
                  )}

                  {/* 送信ステータス */}
                  {isProposalSubmitted && (
                    <Badge
                      size="sm"
                      color="teal"
                      variant="filled"
                      leftSection={<IconClock size={12} />}
                    >
                      承認待ち
                    </Badge>
                  )}
                </Group>
              }
              footer={
                <Paper
                  radius="md"
                  p={6}
                  style={{
                    backgroundColor: 'rgba(0, 0, 0, 0.75)',
                    backdropFilter: 'blur(8px)',
                    width: '100%',
                    maxWidth: 400,
                  }}
                >
                  <Group justify="space-between" gap="xs" wrap="nowrap">
                    {/* 変更がある場合のみ「元に戻す」ボタンを表示 */}
                    {changedCount > 0 && !isProposalSubmitted && (
                      <Button
                        variant="subtle"
                        color="gray"
                        size="sm"
                        c="gray.3"
                        onClick={handleResetDraft}
                        disabled={isSubmitting}
                      >
                        元に戻す
                      </Button>
                    )}

                    {/* 提案送信 / 差分確認ボタン */}
                    <Button
                      style={{ flex: 1 }}
                      color={
                        isProposalSubmitted
                          ? 'teal'
                          : changedCount > 0
                            ? 'violet'
                            : 'gray'
                      }
                      variant="filled"
                      loading={isSubmitting}
                      disabled={changedCount === 0 && !isProposalSubmitted}
                      onClick={() => setIsDiffModalOpen(true)}
                      leftSection={
                        isProposalSubmitted ? (
                          <IconCheck size={18} />
                        ) : (
                          <IconSend size={18} />
                        )
                      }
                    >
                      {isProposalSubmitted
                        ? '提案送信済み (承認待ち)'
                        : changedCount > 0
                          ? `${changedCount}項目の修正を提案`
                          : '公式回答と一致'}
                    </Button>
                  </Group>
                </Paper>
              }
            />
          </Box>
        ) : (
          <Paper p="xl" ta="center" w="100%">
            <Text c="dimmed" mb="sm">
              該当するメンバーが見つかりません。
            </Text>
            <Button
              variant="light"
              leftSection={<IconFilter size={16} />}
              onClick={() => setIsFilterModalOpen(true)}
            >
              フィルターを変更
            </Button>
          </Paper>
        )}

        {/* 3. ナビゲーションバー (前へ / カウンター・ジャンプ / 次へ) */}
        {filteredMembers.length > 0 && (
          <Paper
            withBorder
            px="md"
            py={6}
            radius="md"
            style={{ width: '100%', maxWidth: 440 }}
          >
            <Group justify="space-between" align="center" wrap="nowrap">
              <Button
                variant="subtle"
                color="gray"
                size="xs"
                px={6}
                leftSection={<IconChevronLeft size={16} />}
                disabled={currentIndex === 0}
                onClick={() => handleNavigate(-1)}
              >
                前へ
              </Button>

              <Button
                variant="light"
                color="blue"
                size="xs"
                radius="xl"
                leftSection={<IconSearch size={14} />}
                onClick={() => setIsPickerModalOpen(true)}
                title="タップしてメンバー一覧・検索から選択"
              >
                {currentIndex + 1} / {filteredMembers.length} (メンバー一覧)
              </Button>

              <Button
                variant="subtle"
                color="gray"
                size="xs"
                px={6}
                rightSection={<IconChevronRight size={16} />}
                disabled={currentIndex >= filteredMembers.length - 1}
                onClick={() => handleNavigate(1)}
              >
                次へ
              </Button>
            </Group>
          </Paper>
        )}
      </Stack>

      {/* 4. 既存 FilterModal の完全再利用 */}
      <FilterModal
        opened={isFilterModalOpen}
        onClose={() => setIsFilterModalOpen(false)}
        series={bootstrap.series}
        groups={bootstrap.groups}
        allMembers={bootstrap.members}
        allSongs={bootstrap.songs || []}
        currentFilter={filterCriteria}
        onApply={(newFilter) => {
          setFilterCriteria(newFilter);
          setCurrentIndex(0);
          setIsFilterModalOpen(false);
        }}
      />

      {/* 5. 既存 DonutRingModal (カラー選択) */}
      <DonutRingModal
        opened={isDonutModalOpen}
        onClose={() => setIsDonutModalOpen(false)}
        colors={availableColors}
        selectedLeftColor={selectedLeftColor}
        selectedRightColor={selectedRightColor}
        initialHand={activeHand}
        disabled={isSubmitting}
        onColorSelect={(hand, color) => {
          if (!currentDraft) return;
          updateCurrentDraft({
            penlight: {
              ...currentDraft.penlight,
              [hand === 'left' ? 'left_color_id' : 'right_color_id']: color.id,
            },
          });
        }}
        onAnswer={({ leftColorId, rightColorId }) => {
          if (!currentDraft) return;
          updateCurrentDraft({
            penlight: {
              ...currentDraft.penlight,
              left_color_id: leftColorId,
              right_color_id: rightColorId,
            },
          });
        }}
      />

      {/* 6. 期生・ステータス編集モーダル (タップで編集) */}
      {currentMember && currentDraft && (
        <Modal
          opened={isInfoModalOpen}
          onClose={() => setIsInfoModalOpen(false)}
          title={`${currentMember.family_name} ${currentMember.given_name} の情報を編集`}
          centered
          size="sm"
        >
          <Stack gap="md">
            <NumberInput
              label="期生"
              description="加入した期生を選択してください"
              value={currentDraft.generation}
              min={1}
              max={15}
              allowDecimal={false}
              allowNegative={false}
              suffix="期生"
              onChange={(val) => {
                if (typeof val === 'number' && val >= 1) {
                  updateCurrentDraft({ generation: val });
                }
              }}
            />

            <div>
              <Text size="sm" fw={500} mb={4}>
                在籍ステータス
              </Text>
              <SegmentedControl
                fullWidth
                color={currentDraft.status === 'graduated' ? 'red' : 'blue'}
                value={currentDraft.status}
                onChange={(val) =>
                  updateCurrentDraft({ status: val as MemberStatus })
                }
                data={[
                  { label: '現役', value: 'active' },
                  { label: '卒業', value: 'graduated' },
                ]}
              />
            </div>

            <Button fullWidth onClick={() => setIsInfoModalOpen(false)} mt="xs">
              完了
            </Button>
          </Stack>
        </Modal>
      )}

      {/* 7. 変更内容の確認 & 提案送信モーダル */}
      {currentMember && currentDraft && (
        <Modal
          opened={isDiffModalOpen}
          onClose={() => setIsDiffModalOpen(false)}
          title="提案内容の確認"
          centered
          size="md"
        >
          <Stack gap="md">
            {changedCount === 0 ? (
              <Alert color="blue" icon={<IconCheck size={16} />}>
                公式データからの変更はありません。
              </Alert>
            ) : (
              <>
                <Text size="sm">
                  以下の変更内容を編集提案として送信します。
                </Text>

                <Table withTableBorder withColumnBorders>
                  <Table.Thead>
                    <Table.Tr>
                      <Table.Th>変更項目</Table.Th>
                      <Table.Th>現在の公式データ</Table.Th>
                      <Table.Th>あなたの提案</Table.Th>
                    </Table.Tr>
                  </Table.Thead>
                  <Table.Tbody>
                    {changes.penlight && (
                      <Table.Tr>
                        <Table.Td fw={700}>ペンライトカラー</Table.Td>
                        <Table.Td>
                          {colorMap.get(currentMember.penlight.left_color_id)
                            ?.name || '未設定'}{' '}
                          /{' '}
                          {colorMap.get(currentMember.penlight.right_color_id)
                            ?.name || '未設定'}
                        </Table.Td>
                        <Table.Td c="violet.7" fw={700}>
                          {colorMap.get(currentDraft.penlight.left_color_id)
                            ?.name || '未設定'}{' '}
                          /{' '}
                          {colorMap.get(currentDraft.penlight.right_color_id)
                            ?.name || '未設定'}
                        </Table.Td>
                      </Table.Tr>
                    )}
                    {changes.generation && (
                      <Table.Tr>
                        <Table.Td fw={700}>期生</Table.Td>
                        <Table.Td>{currentMember.generation}期生</Table.Td>
                        <Table.Td c="violet.7" fw={700}>
                          {currentDraft.generation}期生
                        </Table.Td>
                      </Table.Tr>
                    )}
                    {changes.status && (
                      <Table.Tr>
                        <Table.Td fw={700}>ステータス</Table.Td>
                        <Table.Td>
                          {statusLabels[currentMember.status]}
                        </Table.Td>
                        <Table.Td c="violet.7" fw={700}>
                          {statusLabels[currentDraft.status]}
                        </Table.Td>
                      </Table.Tr>
                    )}
                  </Table.Tbody>
                </Table>
              </>
            )}

            {currentSubmission?.error && (
              <Alert color="red" icon={<IconAlertCircle size={16} />}>
                {currentSubmission.error}
              </Alert>
            )}

            {isProposalSubmitted && (
              <Alert color="teal" icon={<IconCheck size={16} />}>
                この内容はすでに送信され、承認待ちです。管理者が確認次第、公式データに反映されます。
              </Alert>
            )}

            {!isOnline && (
              <Alert color="yellow" icon={<IconAlertCircle size={16} />}>
                提案の送信にはインターネット接続が必要です。
              </Alert>
            )}

            <Group justify="space-between" mt="xs">
              {changedCount > 0 && !isProposalSubmitted ? (
                <Button
                  variant="subtle"
                  color="red"
                  size="sm"
                  onClick={() => {
                    handleResetDraft();
                    setIsDiffModalOpen(false);
                  }}
                >
                  変更を破棄して戻す
                </Button>
              ) : (
                <Box />
              )}
              <Group gap="xs">
                <Button
                  variant="default"
                  onClick={() => setIsDiffModalOpen(false)}
                >
                  閉じる
                </Button>
                {changedCount > 0 && !isProposalSubmitted && (
                  <Button
                    color="violet"
                    leftSection={<IconSend size={16} />}
                    loading={isSubmitting}
                    disabled={!isOnline}
                    onClick={handleSubmitProposal}
                  >
                    この内容で提案を送信
                  </Button>
                )}
              </Group>
            </Group>
          </Stack>
        </Modal>
      )}

      {/* 8. メンバー直接選択・検索モーダル (クイックジャンプ) */}
      <Modal
        opened={isPickerModalOpen}
        onClose={() => {
          setIsPickerModalOpen(false);
          setMemberSearchQuery('');
        }}
        title="メンバーを選択"
        centered
        size="sm"
      >
        <Stack gap="sm">
          <TextInput
            placeholder="メンバー名で検索…"
            leftSection={<IconSearch size={16} />}
            value={memberSearchQuery}
            onChange={(e) => setMemberSearchQuery(e.currentTarget.value)}
            autoFocus
          />

          <Box style={{ maxHeight: 360, overflowY: 'auto' }}>
            <Stack gap={4}>
              {filteredMembers
                .map((m, idx) => ({ member: m, index: idx }))
                .filter(({ member }) => {
                  if (!memberSearchQuery.trim()) return true;
                  const q = memberSearchQuery.trim().toLowerCase();
                  const fullKanji =
                    `${member.family_name}${member.given_name}`.toLowerCase();
                  const fullKana =
                    `${member.family_name_kana}${member.given_name_kana}`.toLowerCase();
                  return fullKanji.includes(q) || fullKana.includes(q);
                })
                .map(({ member, index }) => (
                  <Button
                    key={member.id}
                    variant={index === currentIndex ? 'filled' : 'subtle'}
                    color={index === currentIndex ? 'blue' : 'gray'}
                    justify="space-between"
                    fullWidth
                    size="sm"
                    onClick={() => {
                      setCurrentIndex(index);
                      setIsPickerModalOpen(false);
                      setMemberSearchQuery('');
                    }}
                  >
                    <Group gap={8}>
                      <Text size="sm">
                        {member.family_name} {member.given_name}
                      </Text>
                      <Badge size="xs" color="orange" variant="light">
                        {member.generation}期
                      </Badge>
                      {member.status !== 'active' && (
                        <Badge size="xs" color="gray" variant="outline">
                          {statusLabels[member.status]}
                        </Badge>
                      )}
                    </Group>
                    {drafts[member.id] &&
                      changeCount(buildChanges(member, drafts[member.id])) >
                        0 && (
                        <Badge size="xs" color="violet" variant="filled">
                          編集中
                        </Badge>
                      )}
                  </Button>
                ))}
            </Stack>
          </Box>
        </Stack>
      </Modal>
    </Container>
  );
}
