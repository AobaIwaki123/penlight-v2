'use client';

import {
  Alert,
  Badge,
  Button,
  Checkbox,
  Group,
  Modal,
  SegmentedControl,
  Stack,
  Switch,
  Text,
} from '@mantine/core';
import {
  IconAlertCircle,
  IconFilter,
  IconMusic,
  IconUsers,
} from '@tabler/icons-react';
import { useEffect, useState } from 'react';
import type {
  Group as IdolGroup,
  Member,
  Series,
  Song,
} from '@/types/generated';

export interface QuizFilterCriteria {
  groupId: string;
  generations: number[];
  includeGraduated: boolean;
  songMode: boolean;
}

interface FilterModalProps {
  opened: boolean;
  onClose: () => void;
  series: Series[];
  groups: IdolGroup[];
  allMembers: Member[];
  allSongs: Song[];
  currentFilter: QuizFilterCriteria;
  onApply: (filter: QuizFilterCriteria) => void;
}

export function FilterModal({
  opened,
  onClose,
  series,
  groups,
  allMembers,
  allSongs,
  currentFilter,
  onApply,
}: FilterModalProps) {
  const [selectedGroupId, setSelectedGroupId] = useState(
    currentFilter.groupId || groups[0]?.id || '',
  );
  const [selectedSeriesId, setSelectedSeriesId] = useState(
    groups.find((g) => g.id === (currentFilter.groupId || groups[0]?.id))
      ?.series_id ||
      series[0]?.id ||
      '',
  );
  const [selectedGenerations, setSelectedGenerations] = useState<number[]>(
    currentFilter.generations || [],
  );
  const [includeGraduated, setIncludeGraduated] = useState(
    currentFilter.includeGraduated,
  );
  const [songMode, setSongMode] = useState(currentFilter.songMode ?? false);

  // Sync internal state when modal opens
  useEffect(() => {
    if (opened) {
      const activeGroupId = currentFilter.groupId || groups[0]?.id || '';
      const matchedGroup = groups.find((g) => g.id === activeGroupId);
      const activeSeriesId = matchedGroup?.series_id || series[0]?.id || '';

      setSelectedGroupId(activeGroupId);
      setSelectedSeriesId(activeSeriesId);
      setSelectedGenerations(currentFilter.generations || []);
      setIncludeGraduated(currentFilter.includeGraduated);
      setSongMode(currentFilter.songMode ?? false);
    }
  }, [opened, currentFilter, groups, series]);

  // Groups belonging to the currently selected series
  const filteredGroups = groups.filter((g) => g.series_id === selectedSeriesId);

  // Available generations for the currently selected group
  const groupMembers = allMembers.filter((m) => m.group_id === selectedGroupId);
  const availableGenerations = Array.from(
    new Set(groupMembers.map((m) => m.generation)),
  ).sort((a, b) => a - b);

  // When switching series, automatically select the first group of that series
  const handleSeriesChange = (newSeriesId: string) => {
    setSelectedSeriesId(newSeriesId);
    const newSeriesGroups = groups.filter((g) => g.series_id === newSeriesId);
    const firstGroup = newSeriesGroups[0];
    if (firstGroup) {
      setSelectedGroupId(firstGroup.id);
      const newGroupMembers = allMembers.filter(
        (m) => m.group_id === firstGroup.id,
      );
      const newGens = Array.from(
        new Set(newGroupMembers.map((m) => m.generation)),
      ).sort((a, b) => a - b);
      setSelectedGenerations(newGens);
    } else {
      setSelectedGroupId('');
      setSelectedGenerations([]);
    }
  };

  // When switching groups, reset generations to all for that group
  const handleGroupChange = (newGroupId: string) => {
    setSelectedGroupId(newGroupId);
    const newGroupMembers = allMembers.filter((m) => m.group_id === newGroupId);
    const newGens = Array.from(
      new Set(newGroupMembers.map((m) => m.generation)),
    ).sort((a, b) => a - b);
    setSelectedGenerations(newGens);
  };

  // Quick select actions for generations
  const handleSelectAllGenerations = () => {
    setSelectedGenerations(availableGenerations);
  };

  const handleClearGenerations = () => {
    setSelectedGenerations([]);
  };

  const handleGenerationToggle = (gen: number) => {
    if (selectedGenerations.includes(gen)) {
      setSelectedGenerations(selectedGenerations.filter((g) => g !== gen));
    } else {
      setSelectedGenerations(
        [...selectedGenerations, gen].sort((a, b) => a - b),
      );
    }
  };

  // Compute matching members count in real-time
  const matchingMembers = allMembers.filter((m) => {
    if (m.group_id !== selectedGroupId) return false;
    if (!includeGraduated && m.status !== 'active') return false;
    if (!m.penlight?.left_color_id || !m.penlight?.right_color_id) return false;
    return selectedGenerations.includes(m.generation);
  });

  // Compute matching songs count in real-time
  const matchingSongs = allSongs.filter((s) => s.group_id === selectedGroupId);

  const isInsufficientMembers = !songMode && matchingMembers.length < 4;
  const isInsufficientSongs = songMode && matchingSongs.length === 0;
  const isInsufficient = isInsufficientMembers || isInsufficientSongs;

  const handleApplyClick = () => {
    if (isInsufficient) return;
    onApply({
      groupId: selectedGroupId,
      generations: selectedGenerations,
      includeGraduated,
      songMode,
    });
    onClose();
  };

  return (
    <Modal
      opened={opened}
      onClose={onClose}
      title={
        <Group gap="xs">
          <IconFilter size={20} />
          <Text fw={700}>出題フィルター設定</Text>
        </Group>
      }
      centered
      radius="md"
      size="sm"
    >
      <Stack gap="md">
        {/* シリーズ選択 (拡張性重視: 選択すると配下のグループが表示される) */}
        {series.length > 1 && (
          <Stack gap={4}>
            <Text size="xs" fw={700} c="dimmed">
              シリーズ選択
            </Text>
            <SegmentedControl
              fullWidth
              value={selectedSeriesId}
              onChange={handleSeriesChange}
              data={series.map((s) => ({
                label: s.name,
                value: s.id,
              }))}
            />
          </Stack>
        )}

        {/* 配下のグループ選択 (文字のみ・SegmentedControl) */}
        <Stack gap={4}>
          <Text size="xs" fw={700} c="dimmed">
            グループ選択
          </Text>
          {filteredGroups.length > 0 ? (
            <SegmentedControl
              fullWidth
              value={selectedGroupId}
              onChange={handleGroupChange}
              data={filteredGroups.map((g) => ({
                label: g.name,
                value: g.id,
              }))}
            />
          ) : (
            <Text size="xs" c="dimmed">
              該当グループがありません
            </Text>
          )}
        </Stack>

        {/* クイズ種別切替: メンバー推しメンカラー vs 楽曲カラー */}
        <Stack gap={4}>
          <Text size="xs" fw={700} c="dimmed">
            クイズ種別
          </Text>
          <Switch
            checked={songMode}
            onChange={(event) => setSongMode(event.currentTarget.checked)}
            label={
              <Group gap="xs">
                {songMode ? <IconMusic size={16} /> : <IconUsers size={16} />}
                <Text size="sm" fw={600}>
                  {songMode
                    ? '楽曲カラークイズ'
                    : 'メンバー推しメンカラークイズ'}
                </Text>
              </Group>
            }
            description={
              songMode
                ? 'ライブでの楽曲指定ペンライトカラーを当てるモード'
                : 'メンバー2色の推しメンカラーを当てる通常モード'
            }
          />
        </Stack>

        {/* メンバーモード専用: 期生選択 & 卒業生オプション */}
        {!songMode && (
          <>
            <Stack gap={4}>
              <Group justify="space-between" align="center">
                <Text size="xs" fw={700} c="dimmed">
                  期生選択
                </Text>
                <Group gap={6}>
                  <Button
                    variant="subtle"
                    size="compact-xs"
                    onClick={handleSelectAllGenerations}
                  >
                    全選択
                  </Button>
                  <Button
                    variant="subtle"
                    color="gray"
                    size="compact-xs"
                    onClick={handleClearGenerations}
                  >
                    解除
                  </Button>
                </Group>
              </Group>
              <Group gap="md">
                {availableGenerations.map((gen) => (
                  <Checkbox
                    key={gen}
                    label={`${gen}期生`}
                    checked={selectedGenerations.includes(gen)}
                    onChange={() => handleGenerationToggle(gen)}
                  />
                ))}
              </Group>
            </Stack>

            <Switch
              label="卒業生も含めて出題する"
              checked={includeGraduated}
              onChange={(event) =>
                setIncludeGraduated(event.currentTarget.checked)
              }
            />
          </>
        )}

        {/* 該当件数バッジ & 警告 */}
        <Group justify="space-between" align="center" pt="xs">
          <Text size="sm" fw={600}>
            {songMode ? '該当楽曲数' : '該当メンバー数'}
          </Text>
          <Badge
            size="lg"
            variant="light"
            color={isInsufficient ? 'red' : 'blue'}
          >
            {songMode
              ? `${matchingSongs.length} 曲`
              : `${matchingMembers.length} 名`}
          </Badge>
        </Group>

        {isInsufficientMembers && (
          <Alert
            icon={<IconAlertCircle size={16} />}
            color="red"
            variant="light"
            title="出題メンバー不足"
          >
            出題には最低4名のメンバーが必要です。期生または卒業生設定を変更してください。
          </Alert>
        )}

        {isInsufficientSongs && (
          <Alert
            icon={<IconAlertCircle size={16} />}
            color="red"
            variant="light"
            title="出題楽曲なし"
          >
            選択したグループには出題可能な楽曲データが登録されていません。グループを変更してください。
          </Alert>
        )}

        {/* 確定ボタン */}
        <Button
          fullWidth
          size="md"
          color="blue"
          onClick={handleApplyClick}
          disabled={isInsufficient}
          mt="xs"
        >
          この条件でクイズを開始
        </Button>
      </Stack>
    </Modal>
  );
}
