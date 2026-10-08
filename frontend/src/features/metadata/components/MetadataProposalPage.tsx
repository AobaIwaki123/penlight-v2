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
  Modal,
  NumberInput,
  Paper,
  Progress,
  Select,
  SimpleGrid,
  Stack,
  Text,
  Title,
} from '@mantine/core';
import {
  IconAlertCircle,
  IconArrowLeft,
  IconArrowRight,
  IconArrowsLeftRight,
  IconCheck,
  IconCloudUpload,
  IconEdit,
  IconPhoto,
  IconRotateClockwise,
} from '@tabler/icons-react';
import Link from 'next/link';
import {
  type PointerEvent as ReactPointerEvent,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import {
  fetchBootstrapData,
  getImageUrl,
  submitMetadataEditProposal,
} from '@/features/quiz/api/client';
import { PenlightStick } from '@/features/quiz/components/PenlightStick';
import { generateMetadataProposalID } from '@/features/quiz/utils/id';
import type {
  BootstrapResponse,
  Color,
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

function statusLabel(status: MemberStatus): string {
  switch (status) {
    case 'active':
      return '現役';
    case 'graduated':
      return '卒業';
    case 'hiatus':
      return '休業中';
    default:
      return status;
  }
}

function statusColor(status: MemberStatus): string {
  switch (status) {
    case 'active':
      return 'teal';
    case 'graduated':
      return 'gray';
    default:
      return 'yellow';
  }
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
      before: statusLabel(changes.status.before),
      after: statusLabel(changes.status.after),
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

function MetadataProposalEditorModal({
  opened,
  onClose,
  member,
  colors,
  photoTypes,
  onSubmitted,
}: {
  opened: boolean;
  onClose: () => void;
  member: Member;
  colors: Color[];
  photoTypes: PhotoType[];
  onSubmitted: (proposal: MetadataEditProposal) => void;
}) {
  const [draft, setDraft] = useState<DraftState>(() => createDraft(member));
  const [proposal, setProposal] = useState<MetadataEditProposal | null>(null);
  const [proposalId, setProposalId] = useState<string | null>(null);
  const [isOnline, setIsOnline] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const proposalBodyKey = useRef<string | null>(null);

  useEffect(() => {
    if (!opened) return;
    setDraft(createDraft(member));
    setProposal(null);
    setProposalId(null);
    setSubmitError(null);
    proposalBodyKey.current = null;
  }, [opened, member]);

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

  const availableColors = useMemo(
    () =>
      colors.filter(
        (color) => !color.group_id || color.group_id === member.group_id,
      ),
    [colors, member.group_id],
  );
  const availablePhotoTypes = useMemo(
    () =>
      photoTypes.filter((photoType) => photoType.group_id === member.group_id),
    [member.group_id, photoTypes],
  );
  const colorMap = useMemo(
    () => new Map(colors.map((color) => [color.id, color])),
    [colors],
  );
  const changes = useMemo(() => buildChanges(member, draft), [draft, member]);
  const changedItems = changeCount(changes);
  const bodyKey = JSON.stringify({
    member_id: member.id,
    base_revision: member.metadata_revision,
    changes,
  });
  const proposalIsCurrent =
    proposal !== null && proposalBodyKey.current === bodyKey;
  const reusableProposalID =
    proposalId !== null && proposalBodyKey.current === bodyKey;

  const updateDraft = (update: Partial<DraftState>) => {
    setDraft((current) => ({ ...current, ...update }));
  };

  const updatePhotoType = (imageId: string, photoTypeId: string | null) => {
    if (!photoTypeId) return;
    setDraft((current) => ({
      ...current,
      photoTypeIDs: {
        ...current.photoTypeIDs,
        [imageId]: photoTypeId,
      },
    }));
  };

  const handleSubmit = async () => {
    if (changedItems === 0 || !isOnline) return;

    let currentProposalId = proposalId;
    if (!currentProposalId || proposalBodyKey.current !== bodyKey) {
      currentProposalId = generateMetadataProposalID();
      setProposalId(currentProposalId);
      proposalBodyKey.current = bodyKey;
      setProposal(null);
    }

    const request: SubmitMetadataEditProposalRequest = {
      id: currentProposalId,
      base_revision: member.metadata_revision,
      changes,
    };

    setIsSubmitting(true);
    setSubmitError(null);
    try {
      const result = await submitMetadataEditProposal(member.id, request);
      setProposal(result);
      onSubmitted(result);
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : String(err));
    } finally {
      setIsSubmitting(false);
    }
  };

  const primaryImage = getPrimaryImage(member);
  const colorOptions = availableColors.map((color) => ({
    value: color.id,
    label: colorLabel(color),
  }));
  const photoTypeOptions = availablePhotoTypes.map((photoType) => ({
    value: photoType.id,
    label: photoType.name,
  }));

  return (
    <Modal
      opened={opened}
      onClose={onClose}
      title={
        <Group gap="xs">
          <IconEdit size={20} />
          <Text fw={700}>このメンバーの修正提案</Text>
        </Group>
      }
      centered
      radius="md"
      size="lg"
    >
      <Stack gap="md">
        <Paper
          radius="md"
          p="sm"
          withBorder
          style={{
            background:
              'linear-gradient(135deg, var(--mantine-color-violet-0), transparent)',
          }}
        >
          <Group justify="space-between" align="center" wrap="nowrap">
            <Group gap="sm" wrap="nowrap">
              <Image
                src={getImageUrl(primaryImage?.image_key)}
                alt={`${member.family_name} ${member.given_name}`}
                w={64}
                h={64}
                radius="sm"
                fit="cover"
                fallbackSrc="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='64' height='64'%3E%3Crect width='100%25' height='100%25' fill='%23f1f3f5'/%3E%3C/svg%3E"
              />
              <Box>
                <Text fw={800}>
                  {member.family_name} {member.given_name}
                </Text>
                <Text size="xs" c="dimmed">
                  {member.generation}期生 / revision {member.metadata_revision}
                </Text>
              </Box>
            </Group>
            <Group gap="xs" wrap="nowrap">
              <PenlightStick
                color={colorMap.get(member.penlight.left_color_id)}
                label="公式 左"
                height={42}
                width={18}
              />
              <PenlightStick
                color={colorMap.get(member.penlight.right_color_id)}
                label="公式 右"
                height={42}
                width={18}
              />
            </Group>
          </Group>
        </Paper>

        <Alert color="violet" icon={<IconCloudUpload size={18} />}>
          公開データは直接変更されません。変更前後を含む提案として保存され、承認後に反映されます。
        </Alert>

        {!isOnline && (
          <Alert color="yellow" icon={<IconAlertCircle size={18} />}>
            オフライン中は提案を送信できません。オンライン復帰後に再送してください。
          </Alert>
        )}

        <Divider label="変更内容" labelPosition="left" />
        <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="sm">
          <Select
            label="左手の公式色"
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
            label="右手の公式色"
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
        {(member.images || []).length === 0 ? (
          <Text size="sm" c="dimmed">
            登録写真がありません。
          </Text>
        ) : (
          <SimpleGrid cols={{ base: 2, sm: 3 }} spacing="sm">
            {(member.images || []).map((image) => {
              const isSelected = draft.primaryImageId === image.id;
              return (
                <Card
                  key={image.id}
                  withBorder
                  p="xs"
                  radius="sm"
                  role="button"
                  tabIndex={0}
                  style={{
                    cursor: 'pointer',
                    borderColor: isSelected
                      ? 'var(--mantine-color-violet-6)'
                      : undefined,
                  }}
                  onClick={() => updateDraft({ primaryImageId: image.id })}
                >
                  <Image
                    src={getImageUrl(image.image_key)}
                    alt={image.image_key}
                    h={110}
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
        {(member.images || []).length === 0 ? (
          <Text size="sm" c="dimmed">
            登録写真がありません。
          </Text>
        ) : availablePhotoTypes.length === 0 ? (
          <Text size="sm" c="dimmed">
            このグループに利用可能な衣装種別がありません。
          </Text>
        ) : (
          <Stack gap="sm">
            {(member.images || []).map((image) => (
              <Group key={image.id} align="end" wrap="nowrap">
                <IconPhoto size={18} color="gray" />
                <Select
                  style={{ flex: 1 }}
                  label={image.image_key}
                  data={photoTypeOptions}
                  value={draft.photoTypeIDs[image.id] || image.photo_type_id}
                  onChange={(value) => updatePhotoType(image.id, value)}
                  searchable
                />
              </Group>
            ))}
          </Stack>
        )}

        <Card withBorder radius="md" p="sm">
          <Stack gap="sm">
            <Group justify="space-between">
              <Text fw={700}>変更前後を確認</Text>
              <Badge
                color={changedItems > 0 ? 'violet' : 'gray'}
                variant="light"
              >
                {changedItems > 0 ? `${changedItems}件の変更` : '変更なし'}
              </Badge>
            </Group>
            <ProposalDiff
              member={member}
              changes={changes}
              colors={colors}
              photoTypes={photoTypes}
            />
            {proposalIsCurrent && proposal && (
              <Alert
                color="teal"
                icon={<IconCheck size={18} />}
                title="承認待ちとして保存しました"
              >
                提案 ID: {proposal.id}
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
                    同じ内容の再送では {proposalId} を再利用します。
                  </Text>
                )}
              </Alert>
            )}
            <Group grow>
              <Button variant="default" onClick={onClose}>
                閉じる
              </Button>
              <Button
                color="violet"
                leftSection={
                  reusableProposalID ? (
                    <IconRotateClockwise size={18} />
                  ) : (
                    <IconCheck size={18} />
                  )
                }
                loading={isSubmitting}
                disabled={changedItems === 0 || !isOnline}
                onClick={handleSubmit}
              >
                {reusableProposalID ? '同じ提案を再送信' : '提案を送信'}
              </Button>
            </Group>
          </Stack>
        </Card>
      </Stack>
    </Modal>
  );
}

function MemberTriageCard({
  member,
  colors,
  pending,
  onSwipe,
}: {
  member: Member;
  colors: Color[];
  pending?: MetadataEditProposal;
  onSwipe: () => void;
}) {
  const gestureStart = useRef<{ x: number; y: number } | null>(null);
  const colorMap = new Map(colors.map((color) => [color.id, color]));
  const primaryImage = getPrimaryImage(member) || member.images?.[0];
  const imageSrc =
    getImageUrl(primaryImage?.image_key) ||
    'https://placehold.co/400x560/7cc7e8/ffffff?text=Penlight+Quiz';

  const handlePointerDown = (event: ReactPointerEvent<HTMLDivElement>) => {
    if (event.pointerType === 'mouse' && event.button !== 0) return;
    gestureStart.current = { x: event.clientX, y: event.clientY };
    event.currentTarget.setPointerCapture(event.pointerId);
  };

  const handlePointerUp = (event: ReactPointerEvent<HTMLDivElement>) => {
    const start = gestureStart.current;
    gestureStart.current = null;
    if (!start) return;
    const deltaX = event.clientX - start.x;
    const deltaY = event.clientY - start.y;
    if (Math.abs(deltaX) >= 56 && Math.abs(deltaX) > Math.abs(deltaY)) {
      onSwipe();
    }
  };

  return (
    <Paper
      radius="lg"
      shadow="md"
      onPointerDown={handlePointerDown}
      onPointerUp={handlePointerUp}
      onPointerCancel={() => {
        gestureStart.current = null;
      }}
      style={{
        position: 'relative',
        width: '100%',
        height: 'min(62dvh, 520px)',
        minHeight: 360,
        overflow: 'hidden',
        backgroundColor: '#000',
        cursor: 'grab',
        touchAction: 'pan-y',
        userSelect: 'none',
      }}
    >
      <Image
        src={imageSrc}
        alt={`${member.family_name} ${member.given_name}`}
        fit="cover"
        style={{
          position: 'absolute',
          inset: 0,
          width: '100%',
          height: '100%',
        }}
      />
      <Box
        style={{
          position: 'absolute',
          inset: 0,
          background:
            'linear-gradient(to top, rgba(0,0,0,0.94) 0%, rgba(0,0,0,0.48) 56%, transparent 100%)',
          pointerEvents: 'none',
        }}
      />
      <Group
        justify="space-between"
        align="flex-start"
        style={{ position: 'absolute', top: 14, left: 14, right: 14 }}
      >
        <Badge color="violet" variant="filled">
          公式データを表示中
        </Badge>
        {pending && (
          <Badge
            color="teal"
            variant="filled"
            leftSection={<IconCheck size={12} />}
          >
            提案済み・承認待ち
          </Badge>
        )}
      </Group>
      <Box
        style={{
          position: 'absolute',
          bottom: 16,
          left: 16,
          right: 16,
        }}
      >
        <Group justify="space-between" align="flex-end" wrap="nowrap">
          <Box
            style={{ color: '#fff', textShadow: '0 2px 4px rgba(0,0,0,0.6)' }}
          >
            <Group gap="xs" align="center" mb={4} wrap="nowrap">
              <Text size="xl" fw={800} c="white">
                {member.family_name} {member.given_name}
              </Text>
              <Badge
                size="sm"
                color={statusColor(member.status)}
                variant="filled"
              >
                {member.generation}期生 / {statusLabel(member.status)}
              </Badge>
            </Group>
            <Text size="xs" c="gray.3">
              ペンライト正解: 左右の色を表示しています
            </Text>
          </Box>
          <Group gap={12} align="flex-end" wrap="nowrap">
            <PenlightStick
              color={colorMap.get(member.penlight.left_color_id)}
              label="公式 左"
              height={74}
              width={28}
              textColor="#fff"
            />
            <PenlightStick
              color={colorMap.get(member.penlight.right_color_id)}
              label="公式 右"
              height={74}
              width={28}
              textColor="#fff"
            />
          </Group>
        </Group>
      </Box>
    </Paper>
  );
}

export function MetadataProposalPage() {
  const [bootstrap, setBootstrap] = useState<BootstrapResponse | null>(null);
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [currentIndex, setCurrentIndex] = useState(0);
  const [history, setHistory] = useState<number[]>([]);
  const [isComplete, setIsComplete] = useState(false);
  const [isEditorOpen, setIsEditorOpen] = useState(false);
  const [pendingProposals, setPendingProposals] = useState<
    Record<string, MetadataEditProposal>
  >({});
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    fetchBootstrapData({ includeGraduated: true })
      .then((data) => {
        if (!mounted) return;
        const requestedGroup =
          typeof window !== 'undefined'
            ? new URLSearchParams(window.location.search).get('group_id') || ''
            : '';
        const groupID = data.groups.some((group) => group.id === requestedGroup)
          ? requestedGroup
          : data.groups[0]?.id || '';
        setBootstrap(data);
        setSelectedGroupId(groupID);
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

  const groups = bootstrap?.groups || [];
  const members = bootstrap?.members || [];
  const colors = bootstrap?.colors || [];
  const photoTypes = bootstrap?.photo_types || [];
  const groupMembers = useMemo(
    () => members.filter((member) => member.group_id === selectedGroupId),
    [members, selectedGroupId],
  );
  const currentMember = groupMembers[currentIndex];
  const currentPending = currentMember
    ? pendingProposals[currentMember.id]
    : undefined;

  const advanceMember = () => {
    if (!currentMember) return;
    setHistory((current) => [...current, currentIndex]);
    if (currentIndex + 1 >= groupMembers.length) {
      setIsComplete(true);
      return;
    }
    setCurrentIndex((current) => current + 1);
  };

  const undoMember = () => {
    const previous = history[history.length - 1];
    if (previous === undefined) return;
    setHistory((current) => current.slice(0, -1));
    setCurrentIndex(previous);
    setIsComplete(false);
  };

  const restartMembers = () => {
    setCurrentIndex(0);
    setHistory([]);
    setIsComplete(false);
  };

  const handleGroupChange = (groupID: string | null) => {
    if (!groupID) return;
    setSelectedGroupId(groupID);
    restartMembers();
  };

  if (isLoading) {
    return (
      <Container size="xs" py="xl">
        <Stack align="center" gap="sm">
          <Text size="sm" c="dimmed">
            修正提案モードを読み込んでいます...
          </Text>
        </Stack>
      </Container>
    );
  }

  if (error) {
    return (
      <Container size="xs" py="xl">
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

  const selectedGroup = groups.find((group) => group.id === selectedGroupId);
  const groupOptions = groups.map((group) => ({
    value: group.id,
    label: group.name,
  }));

  return (
    <Container
      size="xs"
      p={0}
      style={{ minHeight: '100dvh', display: 'flex', flexDirection: 'column' }}
    >
      <Box
        px="md"
        py="xs"
        style={{
          borderBottom: '1px solid var(--mantine-color-default-border)',
        }}
      >
        <Group justify="space-between" align="center" mb="xs">
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
              修正提案
            </Title>
          </Group>
          <IconEdit size={22} color="var(--mantine-color-violet-6)" />
        </Group>
        <Select
          aria-label="対象グループ"
          data={groupOptions}
          value={selectedGroupId}
          onChange={handleGroupChange}
          size="xs"
          allowDeselect={false}
        />
      </Box>

      <Stack gap="sm" px="md" pt="sm" pb="md" style={{ flexGrow: 1 }}>
        <Group justify="space-between" align="center">
          <Box>
            <Text size="xs" fw={700} c="dimmed">
              {selectedGroup?.name || 'メンバー'} の公式データ確認
            </Text>
            <Text size="xs" c="dimmed">
              左右にスワイプしてメンバーを切り替えます
            </Text>
          </Box>
          <Badge color="blue" variant="light">
            {groupMembers.length > 0 && !isComplete
              ? `${currentIndex + 1} / ${groupMembers.length}`
              : `${groupMembers.length}人`}
          </Badge>
        </Group>

        <Progress
          value={
            groupMembers.length > 0 && !isComplete
              ? ((currentIndex + 1) / groupMembers.length) * 100
              : isComplete
                ? 100
                : 0
          }
          size="xs"
          radius="xl"
          color="violet"
        />

        {groupMembers.length === 0 ? (
          <Alert color="gray" icon={<IconAlertCircle size={18} />}>
            このグループには表示できるメンバーがいません。
          </Alert>
        ) : isComplete ? (
          <Paper withBorder radius="lg" p="xl" style={{ textAlign: 'center' }}>
            <Stack align="center" gap="sm">
              <IconCheck size={44} color="var(--mantine-color-teal-6)" />
              <Text fw={800}>このグループを確認しました</Text>
              <Text size="sm" c="dimmed">
                提案済みの内容は承認待ちとして保存されています。
              </Text>
              <Button
                variant="light"
                color="violet"
                leftSection={<IconRotateClockwise size={16} />}
                onClick={restartMembers}
              >
                先頭からもう一度見る
              </Button>
            </Stack>
          </Paper>
        ) : currentMember ? (
          <>
            <MemberTriageCard
              member={currentMember}
              colors={colors}
              pending={currentPending}
              onSwipe={advanceMember}
            />
            <Text size="xs" c="dimmed" ta="center">
              正解（公式ペンライト色）は表示済みです。必要なメンバーだけ修正を提案できます。
            </Text>
            <Button
              fullWidth
              size="md"
              color="violet"
              variant="light"
              leftSection={<IconEdit size={18} />}
              onClick={() => setIsEditorOpen(true)}
            >
              {currentPending
                ? '提案内容を確認・再送'
                : 'このメンバーの修正を提案'}
            </Button>
            <Group grow>
              <Button
                variant="subtle"
                color="gray"
                disabled={history.length === 0}
                onClick={undoMember}
              >
                ひとつ戻る
              </Button>
              <Button
                variant="light"
                rightSection={<IconArrowRight size={16} />}
                onClick={advanceMember}
              >
                次のメンバー
              </Button>
            </Group>
          </>
        ) : null}
      </Stack>

      {currentMember && (
        <MetadataProposalEditorModal
          opened={isEditorOpen}
          onClose={() => setIsEditorOpen(false)}
          member={currentMember}
          colors={colors}
          photoTypes={photoTypes}
          onSubmitted={(proposal) =>
            setPendingProposals((current) => ({
              ...current,
              [currentMember.id]: proposal,
            }))
          }
        />
      )}
    </Container>
  );
}
