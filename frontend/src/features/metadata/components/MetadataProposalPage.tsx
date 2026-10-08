'use client';

import {
  Alert,
  Badge,
  Box,
  Button,
  Card,
  Container,
  Divider,
  Group,
  Image,
  Loader,
  NumberInput,
  Paper,
  Select,
  SimpleGrid,
  Stack,
  Text,
  Title,
} from '@mantine/core';
import {
  IconAlertCircle,
  IconArrowLeft,
  IconArrowsLeftRight,
  IconCheck,
  IconCloudUpload,
  IconEdit,
  IconPhoto,
} from '@tabler/icons-react';
import Link from 'next/link';
import { useEffect, useMemo, useRef, useState } from 'react';
import {
  fetchBootstrapData,
  getImageUrl,
  submitMetadataEditProposal,
} from '@/features/quiz/api/client';
import { generateMetadataProposalID } from '@/features/quiz/utils/id';
import type {
  BootstrapResponse,
  Color,
  Group as IdolGroup,
  ImagePhotoTypeChange,
  Member,
  MemberImage,
  MemberStatus,
  MetadataEditChanges,
  MetadataEditProposal,
  PenlightPair,
  PhotoType,
  PrimaryImageChange,
  SubmitMetadataEditProposalRequest,
} from '@/types/generated';

interface DraftState {
  penlight: PenlightPair;
  generation: number;
  status: MemberStatus;
  primaryImageId: string | null;
  photoTypeIDs: Record<string, string>;
}

const statusOptions = [
  { value: 'active', label: '現役' },
  { value: 'graduated', label: '卒業' },
  { value: 'hiatus', label: '休業中' },
];

function getPrimaryImage(member: Member): MemberImage | undefined {
  return member.images?.find((image) => image.is_primary);
}

function createDraft(member: Member): DraftState {
  return {
    penlight: { ...member.penlight },
    generation: member.generation,
    status: member.status,
    primaryImageId: getPrimaryImage(member)?.id || null,
    photoTypeIDs: Object.fromEntries(
      (member.images || []).map((image) => [image.id, image.photo_type_id]),
    ),
  };
}

function memberLabel(member: Member): string {
  const status = member.status === 'active' ? '現役' : member.status;
  return `${member.family_name} ${member.given_name}（${member.generation}期・${status}）`;
}

function colorLabel(color: Color | undefined): string {
  return color ? `${color.name} (${color.hex_code})` : '未選択';
}

function formatPenlight(
  penlight: PenlightPair,
  colorMap: Map<string, Color>,
): string {
  return `${colorLabel(colorMap.get(penlight.left_color_id))} / ${colorLabel(
    colorMap.get(penlight.right_color_id),
  )}`;
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

  const currentPrimary = getPrimaryImage(member)?.id;
  if (draft.primaryImageId && draft.primaryImageId !== currentPrimary) {
    const primaryImageChange: PrimaryImageChange = {
      after: draft.primaryImageId,
    };
    if (currentPrimary) {
      primaryImageChange.before = currentPrimary;
    }
    changes.primary_image_id = primaryImageChange;
  }

  const imagePhotoTypes: ImagePhotoTypeChange[] = (member.images || [])
    .map((image) => ({
      image_id: image.id,
      before: image.photo_type_id,
      after: draft.photoTypeIDs[image.id] || image.photo_type_id,
    }))
    .filter((change) => change.before !== change.after)
    .sort((a, b) => a.image_id.localeCompare(b.image_id));
  if (imagePhotoTypes.length > 0) {
    changes.image_photo_types = imagePhotoTypes;
  }

  return changes;
}

function changeCount(changes: MetadataEditChanges): number {
  return (
    Number(Boolean(changes.penlight)) +
    Number(Boolean(changes.generation)) +
    Number(Boolean(changes.status)) +
    Number(Boolean(changes.primary_image_id)) +
    (changes.image_photo_types?.length || 0)
  );
}

function ProposalDiff({
  member,
  changes,
  colors,
  photoTypes,
}: {
  member: Member;
  changes: MetadataEditChanges;
  colors: Color[];
  photoTypes: PhotoType[];
}) {
  const colorMap = new Map(colors.map((color) => [color.id, color]));
  const photoTypeMap = new Map(
    photoTypes.map((photoType) => [photoType.id, photoType]),
  );
  const imageMap = new Map(
    (member.images || []).map((image) => [image.id, image]),
  );
  const rows: Array<{ label: string; before: string; after: string }> = [];

  if (changes.penlight) {
    rows.push({
      label: 'ペンライト色・左右順序',
      before: formatPenlight(changes.penlight.before, colorMap),
      after: formatPenlight(changes.penlight.after, colorMap),
    });
  }
  if (changes.generation) {
    rows.push({
      label: '期生',
      before: `${changes.generation.before}期生`,
      after: `${changes.generation.after}期生`,
    });
  }
  if (changes.status) {
    rows.push({
      label: '状態',
      before: changes.status.before,
      after: changes.status.after,
    });
  }
  if (changes.primary_image_id) {
    rows.push({
      label: '代表写真',
      before: changes.primary_image_id.before
        ? imageMap.get(changes.primary_image_id.before)?.image_key || '設定済み'
        : '未設定',
      after:
        imageMap.get(changes.primary_image_id.after)?.image_key || '選択写真',
    });
  }
  for (const change of changes.image_photo_types || []) {
    rows.push({
      label: `${imageMap.get(change.image_id)?.image_key || change.image_id} の衣装タグ`,
      before: photoTypeMap.get(change.before)?.name || change.before,
      after: photoTypeMap.get(change.after)?.name || change.after,
    });
  }

  if (rows.length === 0) {
    return (
      <Text size="sm" c="dimmed">
        変更項目を選択すると、ここに変更前後が表示されます。
      </Text>
    );
  }

  return (
    <Stack gap="xs">
      {rows.map((row) => (
        <Paper key={row.label} withBorder p="sm" radius="sm">
          <Text size="xs" fw={700} c="dimmed" mb={4}>
            {row.label}
          </Text>
          <SimpleGrid cols={{ base: 1, xs: 2 }} spacing="xs">
            <Box>
              <Text size="xs" c="dimmed">
                変更前
              </Text>
              <Text size="sm">{row.before}</Text>
            </Box>
            <Box>
              <Text size="xs" c="violet.7">
                変更後
              </Text>
              <Text size="sm" fw={600} c="violet.8">
                {row.after}
              </Text>
            </Box>
          </SimpleGrid>
        </Paper>
      ))}
    </Stack>
  );
}

export function MetadataProposalPage() {
  const [bootstrap, setBootstrap] = useState<BootstrapResponse | null>(null);
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [selectedMemberId, setSelectedMemberId] = useState('');
  const [draft, setDraft] = useState<DraftState | null>(null);
  const [proposal, setProposal] = useState<MetadataEditProposal | null>(null);
  const [proposalId, setProposalId] = useState<string | null>(null);
  const [isOnline, setIsOnline] = useState(true);
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const proposalBodyKey = useRef<string | null>(null);

  useEffect(() => {
    let mounted = true;
    fetchBootstrapData({ includeGraduated: true })
      .then((data) => {
        if (!mounted) return;
        setBootstrap(data);
        setSelectedGroupId(data.groups[0]?.id || '');
        setIsLoading(false);
      })
      .catch((err) => {
        if (!mounted) return;
        setError(err instanceof Error ? err.message : String(err));
        setIsLoading(false);
      });

    return () => {
      mounted = false;
    };
  }, []);

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

  const groups = bootstrap?.groups || [];
  const members = bootstrap?.members || [];
  const colors = bootstrap?.colors || [];
  const photoTypes = bootstrap?.photo_types || [];

  const selectedGroup = groups.find((group) => group.id === selectedGroupId);
  const groupMembers = useMemo(
    () => members.filter((member) => member.group_id === selectedGroupId),
    [members, selectedGroupId],
  );
  const selectedMember = members.find(
    (member) => member.id === selectedMemberId,
  );

  useEffect(() => {
    if (!groupMembers.some((member) => member.id === selectedMemberId)) {
      setSelectedMemberId(groupMembers[0]?.id || '');
    }
  }, [groupMembers, selectedMemberId]);

  useEffect(() => {
    if (!selectedMember) {
      setDraft(null);
      return;
    }
    setDraft(createDraft(selectedMember));
    setProposal(null);
    setProposalId(null);
    setSubmitError(null);
    proposalBodyKey.current = null;
  }, [selectedMember]);

  const availableColors = useMemo(
    () =>
      colors.filter(
        (color) =>
          !color.group_id || color.group_id === selectedMember?.group_id,
      ),
    [colors, selectedMember?.group_id],
  );
  const availablePhotoTypes = useMemo(
    () =>
      photoTypes.filter(
        (photoType) => photoType.group_id === selectedMember?.group_id,
      ),
    [photoTypes, selectedMember?.group_id],
  );
  const changes = useMemo(
    () =>
      selectedMember && draft
        ? buildChanges(selectedMember, draft)
        : ({} satisfies MetadataEditChanges),
    [draft, selectedMember],
  );
  const changedItems = changeCount(changes);

  const groupOptions = groups.map((group: IdolGroup) => ({
    value: group.id,
    label: group.name,
  }));
  const memberOptions = groupMembers.map((member) => ({
    value: member.id,
    label: memberLabel(member),
  }));
  const colorOptions = availableColors.map((color) => ({
    value: color.id,
    label: colorLabel(color),
  }));
  const photoTypeOptions = availablePhotoTypes.map((photoType: PhotoType) => ({
    value: photoType.id,
    label: photoType.name,
  }));

  const updateDraft = (update: Partial<DraftState>) => {
    setDraft((current) => (current ? { ...current, ...update } : current));
  };

  const handleSubmit = async () => {
    if (!selectedMember || !draft || changedItems === 0 || !isOnline) return;

    const bodyKey = JSON.stringify({
      member_id: selectedMember.id,
      base_revision: selectedMember.metadata_revision,
      changes,
    });
    let currentProposalId = proposalId;
    if (!currentProposalId || proposalBodyKey.current !== bodyKey) {
      currentProposalId = generateMetadataProposalID();
      setProposalId(currentProposalId);
      proposalBodyKey.current = bodyKey;
    }

    const request: SubmitMetadataEditProposalRequest = {
      id: currentProposalId,
      base_revision: selectedMember.metadata_revision,
      changes,
    };

    setIsSubmitting(true);
    setSubmitError(null);
    try {
      const result = await submitMetadataEditProposal(
        selectedMember.id,
        request,
      );
      setProposal(result);
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : String(err));
    } finally {
      setIsSubmitting(false);
    }
  };

  const updatePhotoType = (imageId: string, photoTypeId: string | null) => {
    if (!photoTypeId) return;
    setDraft((current) =>
      current
        ? {
            ...current,
            photoTypeIDs: {
              ...current.photoTypeIDs,
              [imageId]: photoTypeId,
            },
          }
        : current,
    );
  };

  if (isLoading) {
    return (
      <Container size="sm" py="xl">
        <Stack align="center" gap="sm">
          <Loader />
          <Text size="sm" c="dimmed">
            編集対象を読み込んでいます...
          </Text>
        </Stack>
      </Container>
    );
  }

  if (error) {
    return (
      <Container size="sm" py="xl">
        <Alert
          color="red"
          title="読み込みに失敗しました"
          icon={<IconAlertCircle />}
        >
          {error}
        </Alert>
        <Button
          component={Link}
          href="/"
          mt="md"
          leftSection={<IconArrowLeft size={16} />}
        >
          クイズへ戻る
        </Button>
      </Container>
    );
  }

  return (
    <Container size="md" py="md" style={{ minHeight: '100dvh' }}>
      <Stack gap="md">
        <Group justify="space-between" align="center">
          <Group gap="xs">
            <Button
              component={Link}
              href="/"
              variant="subtle"
              color="gray"
              px="xs"
              leftSection={<IconArrowLeft size={18} />}
            >
              戻る
            </Button>
            <Title order={2} size="h3">
              データ修正を提案
            </Title>
          </Group>
          <IconEdit size={24} color="var(--mantine-color-violet-6)" />
        </Group>

        <Alert color="violet" icon={<IconCloudUpload size={18} />}>
          変更は公開データへ直接反映されず、変更前後を含む提案として保存されます。管理者の承認後にクイズへ反映されます。
        </Alert>

        {!isOnline && (
          <Alert color="yellow" icon={<IconAlertCircle size={18} />}>
            現在オフラインです。提案の送信はオンライン復帰後に行ってください。
          </Alert>
        )}

        <Card withBorder radius="md" p="md">
          <Stack gap="sm">
            <Text size="sm" fw={700}>
              対象を選択
            </Text>
            <Select
              label="グループ"
              data={groupOptions}
              value={selectedGroupId}
              onChange={(value) => value && setSelectedGroupId(value)}
              searchable
            />
            <Select
              label="メンバー"
              data={memberOptions}
              value={selectedMemberId}
              onChange={(value) => value && setSelectedMemberId(value)}
              searchable
              nothingFoundMessage="該当するメンバーがいません"
            />
            {selectedGroup && (
              <Text size="xs" c="dimmed">
                {selectedGroup.name} の公開マスタを基準に提案を作成します。
              </Text>
            )}
          </Stack>
        </Card>

        {selectedMember && draft && (
          <>
            <Card withBorder radius="md" p="md">
              <Stack gap="md">
                <Group justify="space-between" align="flex-start">
                  <Box>
                    <Text fw={800}>{memberLabel(selectedMember)}</Text>
                    <Text size="xs" c="dimmed" mt={2}>
                      公開 revision: {selectedMember.metadata_revision}
                    </Text>
                  </Box>
                  <Badge
                    color={changedItems > 0 ? 'violet' : 'gray'}
                    variant="light"
                  >
                    {changedItems > 0 ? `${changedItems}件の変更` : '変更なし'}
                  </Badge>
                </Group>

                <Divider label="ペンライト" labelPosition="left" />
                <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="sm">
                  <Select
                    label="左手"
                    data={colorOptions}
                    value={draft.penlight.left_color_id}
                    onChange={(value) =>
                      value &&
                      updateDraft({
                        penlight: { ...draft.penlight, left_color_id: value },
                      })
                    }
                    searchable
                  />
                  <Select
                    label="右手"
                    data={colorOptions}
                    value={draft.penlight.right_color_id}
                    onChange={(value) =>
                      value &&
                      updateDraft({
                        penlight: { ...draft.penlight, right_color_id: value },
                      })
                    }
                    searchable
                  />
                </SimpleGrid>
                <Button
                  variant="light"
                  color="violet"
                  leftSection={<IconArrowsLeftRight size={16} />}
                  onClick={() =>
                    updateDraft({
                      penlight: {
                        ...draft.penlight,
                        left_color_id: draft.penlight.right_color_id,
                        right_color_id: draft.penlight.left_color_id,
                      },
                    })
                  }
                >
                  左右を入れ替える
                </Button>

                <Divider label="基本情報" labelPosition="left" />
                <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="sm">
                  <NumberInput
                    label="期生"
                    min={1}
                    value={draft.generation}
                    onChange={(value) =>
                      updateDraft({
                        generation:
                          typeof value === 'number' && Number.isFinite(value)
                            ? value
                            : draft.generation,
                      })
                    }
                  />
                  <Select
                    label="状態"
                    data={statusOptions}
                    value={draft.status}
                    onChange={(value) =>
                      value && updateDraft({ status: value as MemberStatus })
                    }
                  />
                </SimpleGrid>

                <Divider label="代表写真" labelPosition="left" />
                {(selectedMember.images || []).length === 0 ? (
                  <Text size="sm" c="dimmed">
                    登録写真がありません。
                  </Text>
                ) : (
                  <SimpleGrid cols={{ base: 2, sm: 3 }} spacing="sm">
                    {(selectedMember.images || []).map((image) => {
                      const isSelected = draft.primaryImageId === image.id;
                      return (
                        <Card
                          key={image.id}
                          withBorder
                          p="xs"
                          radius="sm"
                          style={{
                            cursor: 'pointer',
                            borderColor: isSelected
                              ? 'var(--mantine-color-violet-6)'
                              : undefined,
                          }}
                          onClick={() =>
                            updateDraft({ primaryImageId: image.id })
                          }
                        >
                          <Image
                            src={getImageUrl(image.image_key)}
                            alt={image.image_key}
                            h={130}
                            fit="cover"
                            fallbackSrc="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='240' height='160'%3E%3Crect width='100%25' height='100%25' fill='%23f1f3f5'/%3E%3C/svg%3E"
                          />
                          <Group justify="space-between" gap={4} mt="xs">
                            <Text size="xs" truncate>
                              {image.photo_type?.name || image.image_key}
                            </Text>
                            {isSelected && (
                              <IconCheck
                                size={16}
                                color="var(--mantine-color-violet-6)"
                              />
                            )}
                          </Group>
                        </Card>
                      );
                    })}
                  </SimpleGrid>
                )}

                <Divider label="衣装タグ" labelPosition="left" />
                {(selectedMember.images || []).length === 0 ? (
                  <Text size="sm" c="dimmed">
                    登録写真がありません。
                  </Text>
                ) : availablePhotoTypes.length === 0 ? (
                  <Text size="sm" c="dimmed">
                    このグループに利用可能な衣装種別がありません。
                  </Text>
                ) : (
                  <Stack gap="sm">
                    {(selectedMember.images || []).map((image) => (
                      <Group key={image.id} align="end" wrap="nowrap">
                        <IconPhoto size={18} color="gray" />
                        <Select
                          style={{ flex: 1 }}
                          label={image.image_key}
                          data={photoTypeOptions}
                          value={
                            draft.photoTypeIDs[image.id] || image.photo_type_id
                          }
                          onChange={(value) => updatePhotoType(image.id, value)}
                          searchable
                        />
                      </Group>
                    ))}
                  </Stack>
                )}
              </Stack>
            </Card>

            <Card withBorder radius="md" p="md">
              <Stack gap="sm">
                <Group justify="space-between">
                  <Text fw={700}>変更前後を確認</Text>
                  <Badge color="violet" variant="light">
                    承認待ちで保存
                  </Badge>
                </Group>
                <ProposalDiff
                  member={selectedMember}
                  changes={changes}
                  colors={colors}
                  photoTypes={photoTypes}
                />
                {proposal && (
                  <Alert
                    color="teal"
                    icon={<IconCheck size={18} />}
                    title="提案を送信しました"
                  >
                    承認待ちとして保存されています。提案 ID: {proposal.id}
                  </Alert>
                )}
                {submitError && (
                  <Alert
                    color="red"
                    icon={<IconAlertCircle size={18} />}
                    title="送信できませんでした"
                  >
                    {submitError}
                    {proposalId && (
                      <Text size="xs" mt={4}>
                        同じ内容の再送では提案 ID {proposalId} を再利用します。
                      </Text>
                    )}
                  </Alert>
                )}
                <Button
                  fullWidth
                  size="md"
                  color="violet"
                  leftSection={
                    proposalId ? (
                      <IconCloudUpload size={18} />
                    ) : (
                      <IconCheck size={18} />
                    )
                  }
                  loading={isSubmitting}
                  disabled={changedItems === 0 || !isOnline}
                  onClick={handleSubmit}
                >
                  {proposalId ? '同じ提案を再送信' : '変更提案を送信'}
                </Button>
              </Stack>
            </Card>
          </>
        )}
      </Stack>
    </Container>
  );
}
