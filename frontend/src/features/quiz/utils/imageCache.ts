import { getImageUrl } from '@/features/quiz/api/client';
import type { Member } from '@/types/generated';

const IMAGE_CACHE_NAME = 'penlight-images-v1';

/**
 * Register lightweight image caching Service Worker.
 */
export function registerImageServiceWorker(): void {
  if (typeof window !== 'undefined' && 'serviceWorker' in navigator) {
    window.addEventListener('load', () => {
      navigator.serviceWorker.register('/sw.js').catch((err) => {
        console.warn('Service Worker registration skipped:', err);
      });
    });
  }
}

interface ImagePrefetchOptions {
  /**
   * Optional group ID for scoping image downloads (extensibility design).
   * If omitted, prefetches all provided members.
   */
  groupId?: string;
  /**
   * Progress notification callback.
   */
  onProgress?: (completed: number, total: number) => void;
}

interface PrefetchResult {
  total: number;
  cached: number;
  newlyFetched: number;
}

/**
 * Prefetch member images into CacheStorage with extensibility for group filtering.
 */
export async function prefetchImages(
  members: Member[],
  options?: ImagePrefetchOptions,
): Promise<PrefetchResult> {
  if (typeof window === 'undefined' || !('caches' in window)) {
    return { total: 0, cached: 0, newlyFetched: 0 };
  }

  const targetMembers = options?.groupId
    ? members.filter((m) => m.group_id === options.groupId)
    : members;

  const imageUrls = Array.from(
    new Set(
      targetMembers
        .flatMap((m) =>
          (m.images || []).map((img) => getImageUrl(img.image_key)),
        )
        .filter(Boolean),
    ),
  );

  const total = imageUrls.length;
  if (total === 0) {
    options?.onProgress?.(0, 0);
    return { total: 0, cached: 0, newlyFetched: 0 };
  }

  const cache = await caches.open(IMAGE_CACHE_NAME);
  let cached = 0;
  let newlyFetched = 0;

  for (let i = 0; i < total; i++) {
    const url = imageUrls[i];
    try {
      const existing = await cache.match(url);
      if (existing) {
        cached++;
      } else {
        const response = await fetch(url);
        if (response.ok) {
          await cache.put(url, response.clone());
          newlyFetched++;
          cached++;
        }
      }
    } catch (err) {
      console.warn(`Failed to prefetch image ${url}:`, err);
    }
    options?.onProgress?.(i + 1, total);
  }

  return { total, cached, newlyFetched };
}

/**
 * Get count of already cached images in CacheStorage.
 */
export async function getCachedImagesStats(
  members: Member[],
  groupId?: string,
): Promise<{ total: number; cached: number }> {
  if (typeof window === 'undefined' || !('caches' in window)) {
    return { total: 0, cached: 0 };
  }

  const targetMembers = groupId
    ? members.filter((m) => m.group_id === groupId)
    : members;

  const imageUrls = Array.from(
    new Set(
      targetMembers
        .flatMap((m) =>
          (m.images || []).map((img) => getImageUrl(img.image_key)),
        )
        .filter(Boolean),
    ),
  );

  const total = imageUrls.length;
  if (total === 0) return { total: 0, cached: 0 };

  try {
    const cache = await caches.open(IMAGE_CACHE_NAME);
    let cached = 0;
    for (const url of imageUrls) {
      const match = await cache.match(url);
      if (match) {
        cached++;
      }
    }
    return { total, cached };
  } catch {
    return { total, cached: 0 };
  }
}
