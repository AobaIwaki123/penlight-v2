import type { BootstrapResponse } from '@/types/generated';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export async function fetchBootstrapData(): Promise<BootstrapResponse> {
  const res = await fetch(`${API_BASE_URL}/api/v1/sync/bootstrap`, {
    headers: {
      Accept: 'application/json',
    },
  });

  if (!res.ok) {
    throw new Error(
      `Failed to fetch master bootstrap data: ${res.status} ${res.statusText}`,
    );
  }

  return res.json();
}

export function getImageUrl(imageKey?: string): string {
  if (!imageKey) return '';
  if (imageKey.startsWith('http://') || imageKey.startsWith('https://')) {
    return imageKey;
  }
  return `${API_BASE_URL}/images/${imageKey}`;
}
