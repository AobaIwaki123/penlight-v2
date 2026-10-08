'use client';

import {
  ActionIcon,
  Alert,
  Badge,
  Box,
  Button,
  Card,
  Container,
  Group,
  Image,
  NumberInput,
  Popover,
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
  IconArrowsLeftRight,
  IconCheck,
  IconPhoto,
  IconRestore,
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
import { DonutRingModal } from '@/features/quiz/components/inputs/DonutRingModal';
import { LayoutOverlay } from '@/features/quiz/components/layouts/LayoutOverlay';
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

interface SubmissionState {
  id: string;
  bodyKey: string;
  result?: MetadataEditProposal;
  error?: string;
}

function MemberAnswerEditor({
  member,
  colors,
  photoTypes,
  draft,
  submission,
  onDraftChange,
  onSubmissionChange,
  onNavigate,
}: {
  member: Member;
  colors: Color[];
  photoTypes: PhotoType[];
  draft: DraftState;
  submission?: SubmissionState;
  onDraftChange: (draft: DraftState) => void;
  onSubmissionChange: (submission: SubmissionState) => void;
  onNavigate: (direction: -1 | 1) => void;
}) {
  const [activeHand, setActiveHand] = useState<'left' | 'right'>('left');
  const [donutOpened, setDonutOpened] = useState(false);
  const [photosOpened, setPhotosOpened] = useState(false);
  const [isOnline, setIsOnline] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const gesture = useRef<{ x: number; y: number; pointerId: number } | null>(
    null,
  );
  const swiped = useRef(false);
  const colorMap = new Map(colors.map((color) => [color.id, color]));
  const availableColors = colors
    .filter((color) => !color.group_id || color.group_id === member.group_id)
    .sort((a, b) => a.display_order - b.display_order);
  const availablePhotoTypes = photoTypes.filter(
    (type) => type.group_id === member.group_id,
  );
  const primaryImage =
    member.images?.find((image) => image.id === draft.primaryImageId) ||
    getPrimaryImage(member) ||
    member.images?.[0];
  const changes = buildChanges(member, draft);
  const changedItems = changeCount(changes);
  const bodyKey = JSON.stringify({
    base_revision: member.metadata_revision,
    changes,
  });
  const currentSubmission =
    submission?.bodyKey === bodyKey ? submission : undefined;
  const submitted = Boolean(currentSubmission?.result);
  const selectedPhotoType = primaryImage
    ? draft.photoTypeIDs[primaryImage.id] || primaryImage.photo_type_id
    : null;

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

  const updateDraft = (update: Partial<DraftState>) => {
    onDraftChange({ ...draft, ...update });
  };

  const handleSubmit = async () => {
    if (changedItems === 0 || !isOnline || isSubmitting || submitted) return;
    const attempt: SubmissionState = {
      id: currentSubmission?.id || generateMetadataProposalID(),
      bodyKey,
    };
    onSubmissionChange(attempt);
    setIsSubmitting(true);
    try {
      const result = await submitMetadataEditProposal(member.id, {
        id: attempt.id,
        base_revision: member.metadata_revision,
        changes,
      });
      onSubmissionChange({ ...attempt, result });
    } catch (error) {
      onSubmissionChange({
        ...attempt,
        error: error instanceof Error ? error.message : String(error),
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handlePointerDown = (event: ReactPointerEvent<HTMLDivElement>) => {
    swiped.current = false;
    gesture.current = null;
    if (isSubmitting || (event.pointerType === 'mouse' && event.button !== 0))
      return;
    // Input gestures belong to the input. Only the photo surface starts a swipe.
    if (
      !event.currentTarget.contains(event.target as Node) ||
      (event.target as HTMLElement).closest(
        'button, input, select, [role="button"], [role="combobox"]',
      )
    )
      return;
    gesture.current = {
      x: event.clientX,
      y: event.clientY,
      pointerId: event.pointerId,
    };
  };

  const handlePointerMove = (event: ReactPointerEvent<HTMLDivElement>) => {
    const start = gesture.current;
    if (!start || start.pointerId !== event.pointerId) return;
    const dx = event.clientX - start.x;
    const dy = event.clientY - start.y;
    if (Math.abs(dx) > 12 && Math.abs(dx) > Math.abs(dy)) {
      swiped.current = true;
      if (!event.currentTarget.hasPointerCapture(event.pointerId)) {
        event.currentTarget.setPointerCapture(event.pointerId);
      }
    }
  };

  const handlePointerUp = (event: ReactPointerEvent<HTMLDivElement>) => {
    const start = gesture.current;
    gesture.current = null;
    if (!start || start.pointerId !== event.pointerId) return;
    const dx = event.clientX - start.x;
    const dy = event.clientY - start.y;
    if (Math.abs(dx) >= 56 && Math.abs(dx) > Math.abs(dy)) {
      swiped.current = true;
      onNavigate(dx < 0 ? 1 : -1);
    }
  };

  return (
    <>
      <Box
        data-testid="member-swipe-surface"
        onPointerDown={handlePointerDown}
        onPointerMove={handlePointerMove}
        onPointerUp={handlePointerUp}
        onPointerCancel={() => {
          gesture.current = null;
        }}
        onClickCapture={(event) => {
          if (swiped.current) {
            event.preventDefault();
            event.stopPropagation();
            swiped.current = false;
          }
        }}
        onDragStart={(event) => event.preventDefault()}
        style={{
          width: '100%',
          display: 'flex',
          flex: 1,
          minHeight: 0,
          touchAction: 'pan-y',
        }}
      >
        <LayoutOverlay
          target={{ ...member, images: primaryImage ? [primaryImage] : [] }}
          costumeTitle={primaryImage?.photo_type?.name || ''}
          selectedLeftColor={colorMap.get(draft.penlight.left_color_id)}
          selectedRightColor={colorMap.get(draft.penlight.right_color_id)}
          onOpenInput={(hand) => {
            setActiveHand(hand);
            setDonutOpened(true);
          }}
          isFullscreen
          photoControl={
            <Popover
              opened={photosOpened}
              onChange={setPhotosOpened}
              width={300}
              position="bottom-end"
              withinPortal
            >
              <Popover.Target>
                <Button
                  size="xs"
                  color="dark"
                  leftSection={<IconPhoto size={16} />}
                  onClick={() => setPhotosOpened(!photosOpened)}
                  disabled={isSubmitting}
                >
                  写真を選ぶ
                </Button>
              </Popover.Target>
              <Popover.Dropdown>
                <Text size="sm" fw={700} mb="xs">
                  代表写真
                </Text>
                {(member.images || []).length === 0 ? (
                  <Text size="sm" c="dimmed">
                    登録写真がありません
                  </Text>
                ) : (
                  <SimpleGrid cols={2} spacing="xs">
                    {(member.images || []).map((image) => (
                      <Card
                        key={image.id}
                        component="button"
                        type="button"
                        aria-label={`代表写真: ${image.photo_type?.name || '登録写真'}`}
                        aria-pressed={draft.primaryImageId === image.id}
                        withBorder
                        p={4}
                        onClick={() => {
                          updateDraft({ primaryImageId: image.id });
                          setPhotosOpened(false);
                        }}
                        style={{
                          cursor: 'pointer',
                          borderColor:
                            draft.primaryImageId === image.id
                              ? 'var(--mantine-color-violet-6)'
                              : undefined,
                        }}
                      >
                        <Image
                          src={getImageUrl(image.image_key)}
                          alt={image.photo_type?.name || '登録写真'}
                          h={96}
                          fit="cover"
                        />
                        <Text size="xs" mt={4}>
                          {image.photo_type?.name || '登録写真'}
                        </Text>
                      </Card>
                    ))}
                  </SimpleGrid>
                )}
              </Popover.Dropdown>
            </Popover>
          }
          memberDetails={
            <Stack gap={6} w="100%" style={{ maxWidth: 210 }}>
              <Text
                size="xl"
                fw={800}
                c="white"
                data-testid="current-member-name"
              >
                {member.family_name} {member.given_name}
              </Text>
              <Group gap={6} wrap="nowrap">
                <NumberInput
                  aria-label="期生"
                  value={draft.generation}
                  min={1}
                  allowDecimal={false}
                  allowNegative={false}
                  suffix="期生"
                  size="xs"
                  style={{ flex: 1, minWidth: 0 }}
                  disabled={isSubmitting}
                  onChange={(value) => {
                    if (
                      typeof value === 'number' &&
                      Number.isInteger(value) &&
                      value >= 1
                    ) {
                      updateDraft({ generation: value });
                    }
                  }}
                />
                <Select
                  aria-label="在籍状態"
                  value={draft.status}
                  data={statusOptions}
                  size="xs"
                  style={{ flex: 1, minWidth: 0 }}
                  allowDeselect={false}
                  disabled={isSubmitting}
                  onChange={(value) =>
                    value && updateDraft({ status: value as MemberStatus })
                  }
                />
              </Group>
              {primaryImage && (
                <Select
                  aria-label="写真の衣装タグ"
                  size="xs"
                  data={availablePhotoTypes.map((type) => ({
                    value: type.id,
                    label: type.name,
                  }))}
                  value={selectedPhotoType}
                  allowDeselect={false}
                  disabled={isSubmitting}
                  onChange={(value) =>
                    value &&
                    updateDraft({
                      photoTypeIDs: {
                        ...draft.photoTypeIDs,
                        [primaryImage.id]: value,
                      },
                    })
                  }
                />
              )}
            </Stack>
          }
          footer={
            <Group gap="xs" wrap="nowrap" w="100%">
              <ActionIcon
                aria-label="左右の色を入れ替える"
                variant="filled"
                color="dark"
                size="lg"
                disabled={isSubmitting}
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
                <IconArrowsLeftRight size={18} />
              </ActionIcon>
              <Button
                style={{ flex: 1 }}
                color={submitted ? 'teal' : 'violet'}
                leftSection={<IconCheck size={18} />}
                loading={isSubmitting}
                disabled={changedItems === 0 || !isOnline || submitted}
                onClick={handleSubmit}
              >
                {submitted
                  ? '送信済み・承認待ち'
                  : currentSubmission?.error
                    ? '同じ回答を再送'
                    : 'この回答を提案'}
              </Button>
              <ActionIcon
                aria-label="公式の回答に戻す"
                variant="filled"
                color="dark"
                size="lg"
                disabled={changedItems === 0 || isSubmitting}
                onClick={() => onDraftChange(createDraft(member))}
              >
                <IconRestore size={18} />
              </ActionIcon>
            </Group>
          }
        />
      </Box>
      <Group justify="space-between" w="100%">
        <Text size="xs" c="dimmed">
          ← 次のメンバー / 前のメンバー →
        </Text>
        <Badge
          color={submitted ? 'teal' : changedItems > 0 ? 'violet' : 'gray'}
          variant="light"
        >
          {submitted
            ? '承認待ち'
            : changedItems > 0
              ? `${changedItems}項目を編集`
              : '公式の回答'}
        </Badge>
      </Group>
      {currentSubmission?.error && (
        <Alert color="red" icon={<IconAlertCircle size={16} />}>
          {currentSubmission.error}
        </Alert>
      )}
      {!isOnline && (
        <Alert color="yellow" icon={<IconAlertCircle size={16} />}>
          送信にはインターネット接続が必要です。
        </Alert>
      )}
      <DonutRingModal
        opened={donutOpened}
        onClose={() => setDonutOpened(false)}
        colors={availableColors}
        selectedLeftColor={colorMap.get(draft.penlight.left_color_id)}
        selectedRightColor={colorMap.get(draft.penlight.right_color_id)}
        initialHand={activeHand}
        disabled={isSubmitting}
        onColorSelect={(hand, color) =>
          updateDraft({
            penlight: {
              ...draft.penlight,
              [hand === 'left' ? 'left_color_id' : 'right_color_id']: color.id,
            },
          })
        }
        onAnswer={({ leftColorId, rightColorId }) =>
          updateDraft({
            penlight: {
              ...draft.penlight,
              left_color_id: leftColorId,
              right_color_id: rightColorId,
            },
          })
        }
      />
    </>
  );
}

export function MetadataProposalPage() {
  const [bootstrap, setBootstrap] = useState<BootstrapResponse | null>(null);
  const [selectedGroupId, setSelectedGroupId] = useState('');
  const [currentIndex, setCurrentIndex] = useState(0);
  const [drafts, setDrafts] = useState<Record<string, DraftState>>({});
  const [submissions, setSubmissions] = useState<
    Record<string, SubmissionState>
  >({});
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    fetchBootstrapData({ includeGraduated: true })
      .then((data) => {
        if (!mounted) return;
        const requestedGroup = new URLSearchParams(window.location.search).get(
          'group_id',
        );
        setSelectedGroupId(
          data.groups.find((group) => group.id === requestedGroup)?.id ||
            data.groups[0]?.id ||
            '',
        );
        setBootstrap(data);
      })
      .catch((error) => {
        if (mounted)
          setError(error instanceof Error ? error.message : String(error));
      });
    return () => {
      mounted = false;
    };
  }, []);

  const groupMembers = useMemo(
    () =>
      (bootstrap?.members || [])
        .filter((member) => member.group_id === selectedGroupId)
        .sort(
          (a, b) =>
            Number(a.status === 'graduated') - Number(b.status === 'graduated'),
        ),
    [bootstrap, selectedGroupId],
  );
  const currentMember = groupMembers[currentIndex];

  if (error) {
    return (
      <Container size="xs" py="xl">
        <Alert color="red" title="読み込みに失敗しました">
          {error}
        </Alert>
      </Container>
    );
  }
  if (!bootstrap) {
    return (
      <Container size="xs" py="xl">
        <Text size="sm" c="dimmed">
          回答を読み込んでいます…
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
        minHeight: 640,
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <Box
        px="md"
        py="xs"
        style={{
          borderBottom: '1px solid var(--mantine-color-default-border)',
        }}
      >
        <Group justify="space-between" mb="xs">
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
            正しい回答を提案
          </Title>
        </Group>
        <Select
          aria-label="対象グループ"
          data={bootstrap.groups.map((group) => ({
            value: group.id,
            label: group.name,
          }))}
          value={selectedGroupId}
          onChange={(value) => {
            if (!value) return;
            setSelectedGroupId(value);
            setCurrentIndex(0);
          }}
          size="xs"
          allowDeselect={false}
        />
      </Box>
      <Box px="md" pt="xs">
        <Group justify="space-between" mb={4}>
          <Text size="xs" c="dimmed">
            表示されている回答をタップして編集
          </Text>
          <Text size="xs" fw={700}>
            {groupMembers.length ? currentIndex + 1 : 0} / {groupMembers.length}
          </Text>
        </Group>
        <Progress
          value={
            groupMembers.length
              ? ((currentIndex + 1) / groupMembers.length) * 100
              : 0
          }
          size="xs"
          radius="xl"
        />
      </Box>
      <Stack gap="xs" px="md" pt="sm" pb="sm" style={{ flex: 1, minHeight: 0 }}>
        {currentMember ? (
          <MemberAnswerEditor
            key={currentMember.id}
            member={currentMember}
            colors={bootstrap.colors}
            photoTypes={bootstrap.photo_types || []}
            draft={drafts[currentMember.id] || createDraft(currentMember)}
            submission={submissions[currentMember.id]}
            onDraftChange={(draft) =>
              setDrafts((current) => ({
                ...current,
                [currentMember.id]: draft,
              }))
            }
            onSubmissionChange={(submission) =>
              setSubmissions((current) => ({
                ...current,
                [currentMember.id]: submission,
              }))
            }
            onNavigate={(direction) =>
              setCurrentIndex((index) =>
                Math.max(
                  0,
                  Math.min(groupMembers.length - 1, index + direction),
                ),
              )
            }
          />
        ) : (
          <Text c="dimmed">このグループには登録メンバーがいません。</Text>
        )}
        <Text size="10px" c="dimmed" ta="center">
          送信した回答は承認後に公開データへ反映されます。
        </Text>
      </Stack>
    </Container>
  );
}
