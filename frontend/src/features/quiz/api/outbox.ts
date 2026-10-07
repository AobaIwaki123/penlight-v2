import { getApiBaseUrl } from '@/features/quiz/api/client';
import {
  getPendingAnswers,
  removePendingAnswers,
} from '@/features/quiz/utils/idb';
import type {
  BatchAnswerRequest,
  BatchAnswerResponse,
} from '@/types/generated';

let isSyncing = false;

/**
 * Flush all pending answers from IndexedDB outbox to backend /api/v1/quiz/answers/batch.
 */
export async function syncPendingAnswers(): Promise<BatchAnswerResponse | null> {
  if (isSyncing) return null;
  if (typeof navigator !== 'undefined' && !navigator.onLine) {
    return null;
  }

  const pending = await getPendingAnswers();
  if (!pending || pending.length === 0) {
    return null;
  }

  isSyncing = true;
  try {
    const baseUrl = getApiBaseUrl();
    const payload: BatchAnswerRequest = {
      answers: pending,
    };

    const res = await fetch(`${baseUrl}/api/v1/quiz/answers/batch`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
      },
      body: JSON.stringify(payload),
    });

    if (!res.ok) {
      console.warn('Failed to sync answers batch:', res.status, res.statusText);
      return null;
    }

    const data: BatchAnswerResponse = await res.json();
    if (data.synced_ids && data.synced_ids.length > 0) {
      await removePendingAnswers(data.synced_ids);
    }
    return data;
  } catch (err) {
    console.warn('Network error while syncing answer outbox:', err);
    return null;
  } finally {
    isSyncing = false;
  }
}
