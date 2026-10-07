'use client';

import {
  Badge,
  Button,
  Group,
  Modal,
  Progress,
  Stack,
  Text,
  ThemeIcon,
} from '@mantine/core';
import { IconCheck, IconCloudDownload, IconWifiOff } from '@tabler/icons-react';
import { useEffect, useState } from 'react';
import {
  getCachedImagesStats,
  prefetchImages,
} from '@/features/quiz/utils/imageCache';
import type { Member } from '@/types/generated';

interface OfflineModalProps {
  opened: boolean;
  onClose: () => void;
  members: Member[];
}

export function OfflineModal({ opened, onClose, members }: OfflineModalProps) {
  const [isDownloading, setIsDownloading] = useState(false);
  const [completed, setCompleted] = useState(0);
  const [total, setTotal] = useState(0);
  const [cachedCount, setCachedCount] = useState(0);
  const [isFinished, setIsFinished] = useState(false);

  useEffect(() => {
    if (opened && members.length > 0) {
      getCachedImagesStats(members).then((stats) => {
        setTotal(stats.total);
        setCachedCount(stats.cached);
      });
    }
  }, [opened, members]);

  const handleStartDownload = async () => {
    setIsDownloading(true);
    setIsFinished(false);
    setCompleted(0);

    const result = await prefetchImages(members, {
      // 拡張性: 将来グループ別に絞る場合は groupId を渡せる
      onProgress: (done, totalCount) => {
        setCompleted(done);
        setTotal(totalCount);
      },
    });

    setCachedCount(result.cached);
    setIsDownloading(false);
    setIsFinished(true);
  };

  const progressPercent = total > 0 ? Math.round((completed / total) * 100) : 0;
  const isFullyCached = total > 0 && cachedCount >= total;

  return (
    <Modal
      opened={opened}
      onClose={isDownloading ? () => {} : onClose}
      title={
        <Group gap={8}>
          <ThemeIcon color="teal" variant="light" size="md">
            <IconWifiOff size={18} />
          </ThemeIcon>
          <Text fw={700}>オフラインモード準備</Text>
        </Group>
      }
      centered
      radius="md"
    >
      <Stack gap="md">
        <Text size="sm" c="dimmed">
          ライブ会場などの電波圏外でも全メンバー・衣装の写真が表示されるよう、端末のキャッシュストレージに画像を保存します。
        </Text>

        <Group justify="space-between">
          <Text size="sm" fw={600}>
            端末保存状況:
          </Text>
          <Badge
            color={isFullyCached ? 'teal' : 'gray'}
            variant="light"
            size="md"
          >
            {cachedCount} / {total} 枚 保存済み
          </Badge>
        </Group>

        {isDownloading && (
          <Stack gap={6}>
            <Progress
              value={progressPercent}
              size="lg"
              radius="xl"
              animated
              color="teal"
            />
            <Group justify="space-between">
              <Text size="xs" c="dimmed">
                ダウンロード中... ({completed}/{total})
              </Text>
              <Text size="xs" fw={700} c="teal">
                {progressPercent}%
              </Text>
            </Group>
          </Stack>
        )}

        {isFinished && (
          <Group gap={8} c="teal">
            <IconCheck size={18} />
            <Text size="sm" fw={600}>
              全画像のローカル保存が完了しました！
            </Text>
          </Group>
        )}

        <Group justify="flex-end" mt="sm">
          <Button variant="default" onClick={onClose} disabled={isDownloading}>
            閉じる
          </Button>
          <Button
            color="teal"
            leftSection={<IconCloudDownload size={18} />}
            loading={isDownloading}
            onClick={handleStartDownload}
          >
            {isFullyCached ? '再ダウンロード' : '画像を全件ダウンロード'}
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
