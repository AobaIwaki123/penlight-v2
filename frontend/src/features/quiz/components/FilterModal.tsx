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
import { IconAlertCircle, IconFilter } from '@tabler/icons-react';
import { useEffect, useState } from 'react';
import type { Group as IdolGroup, Member } from '@/types/generated';

export interface QuizFilterCriteria {
  groupId: string;
  generations: number[];
  includeGraduated: boolean;
}

interface FilterModalProps {
  opened: boolean;
  onClose: () => void;
  groups: IdolGroup[];
  allMembers: Member[];
  currentFilter: QuizFilterCriteria;
  onApply: (filter: QuizFilterCriteria) => void;
}

export function FilterModal({
  opened,
  onClose,
  groups,
  allMembers,
  currentFilter,
  onApply,
}: FilterModalProps) {
  const [selectedGroupId, setSelectedGroupId] = useState(
    currentFilter.groupId || groups[0]?.id || '',
  );
  const [selectedGenerations, setSelectedGenerations] = useState<number[]>(
    currentFilter.generations || [],
  );
  const [includeGraduated, setIncludeGraduated] = useState(
    currentFilter.includeGraduated,
  );

  // Sync internal state when modal opens
  useEffect(() => {
    if (opened) {
      setSelectedGroupId(currentFilter.groupId || groups[0]?.id || '');
      setSelectedGenerations(currentFilter.generations || []);
      setIncludeGraduated(currentFilter.includeGraduated);
    }
  }, [opened, currentFilter, groups]);

  // Available generations for the currently selected group
  const groupMembers = allMembers.filter((m) => m.group_id === selectedGroupId);
  const availableGenerations = Array.from(
    new Set(groupMembers.map((m) => m.generation)),
  ).sort((a, b) => a - b);

  // When switching groups, select all available generations for that group if none matched
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

  const isInsufficient = matchingMembers.length < 4;

  const handleApplyClick = () => {
    if (isInsufficient) return;
    onApply({
      groupId: selectedGroupId,
      generations: selectedGenerations,
      includeGraduated,
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
        {/* グループ選択 */}
        <Stack gap={4}>
          <Text size="xs" fw={700} c="dimmed">
            グループ選択
          </Text>
          {groups.length > 0 && (
            <SegmentedControl
              fullWidth
              value={selectedGroupId}
              onChange={handleGroupChange}
              data={groups.map((g) => ({
                label: g.name,
                value: g.id,
              }))}
            />
          )}
        </Stack>

        {/* 期生選択 */}
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

        {/* 卒業生オプション */}
        <Switch
          label="卒業生も含めて出題する"
          checked={includeGraduated}
          onChange={(event) => setIncludeGraduated(event.currentTarget.checked)}
        />

        {/* 該当人数バッジ & 警告 */}
        <Group justify="space-between" align="center" pt="xs">
          <Text size="sm" fw={600}>
            該当メンバー数
          </Text>
          <Badge
            size="lg"
            variant="light"
            color={isInsufficient ? 'red' : 'blue'}
          >
            {matchingMembers.length} 名
          </Badge>
        </Group>

        {isInsufficient && (
          <Alert
            icon={<IconAlertCircle size={16} />}
            color="red"
            variant="light"
            title="出題メンバー不足"
          >
            出題には最低4名のメンバーが必要です。期生または卒業生設定を変更してください。
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
