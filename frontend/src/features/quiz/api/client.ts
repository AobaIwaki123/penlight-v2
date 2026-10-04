import type { BootstrapResponse } from '@/types/generated';

function getApiBaseUrl(): string {
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL;
  }
  if (typeof window !== 'undefined' && window.location.port === '3000') {
    return 'http://localhost:8080';
  }
  return '';
}

export async function fetchBootstrapData(): Promise<BootstrapResponse> {
  const baseUrl = getApiBaseUrl();
  const res = await fetch(`${baseUrl}/api/v1/sync/bootstrap`, {
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
  const baseUrl = getApiBaseUrl();
  return `${baseUrl}/images/${imageKey}`;
}
