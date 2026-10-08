import type {
  BootstrapResponse,
  MetadataEditProposal,
  SubmitMetadataEditProposalRequest,
} from '@/types/generated';

export function getApiBaseUrl(): string {
  if (process.env.NEXT_PUBLIC_API_URL) {
    return process.env.NEXT_PUBLIC_API_URL;
  }
  if (typeof window !== 'undefined' && window.location.port === '3000') {
    return 'http://localhost:8080';
  }
  return '';
}

export async function fetchBootstrapData(options?: {
  includeGraduated?: boolean;
}): Promise<BootstrapResponse> {
  const baseUrl = getApiBaseUrl();
  const query = options?.includeGraduated ? '?include_graduated=true' : '';
  const res = await fetch(`${baseUrl}/api/v1/sync/bootstrap${query}`, {
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

export async function submitMetadataEditProposal(
  memberId: string,
  request: SubmitMetadataEditProposalRequest,
): Promise<MetadataEditProposal> {
  const baseUrl = getApiBaseUrl();
  const res = await fetch(
    `${baseUrl}/api/v1/members/${encodeURIComponent(memberId)}/metadata-edit-proposals`,
    {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    },
  );

  if (!res.ok) {
    let message = `提案の送信に失敗しました（${res.status} ${res.statusText}）`;
    try {
      const problem = (await res.json()) as { detail?: string; title?: string };
      message = problem.detail || problem.title || message;
    } catch {
      // Keep the HTTP fallback when the response is not Problem Details JSON.
    }
    throw new Error(message);
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
