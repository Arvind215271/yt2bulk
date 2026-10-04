export interface PreflightRequest {
  inputs: string[];
}

export interface PreflightResponse {
  valid: boolean;
  error?: string;
  unique_video_count: number;
  total_resource_count: number;
  estimated_work: number;
}

const getApiUrl = () => {
  return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
};

export async function checkPreflight(inputs: string[], signal?: AbortSignal): Promise<PreflightResponse> {
  const apiUrl = getApiUrl();
  const res = await fetch(`${apiUrl}/preflight`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ inputs }),
    signal
  });

  const text = await res.text();
  if (!res.ok) {
    throw new Error(text || 'Failed to perform preflight check');
  }

  return JSON.parse(text);
}
