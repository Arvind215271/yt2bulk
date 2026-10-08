export interface PreflightRequest {
  inputs: string[];
}

export interface PreflightStats {
  type: string;
  item_count: number;
  metadata_cached: number;
  metadata_upstream: number;
  subtitle_cached: number;
  subtitle_upstream: number;
}

export interface PlaylistPreflightStats {
  type: string;
  total_playlists: number;
  total_videos: number;
  items: PreflightStats;
}

export interface PreflightResponse {
  valid: boolean;
  error?: string;
  unique_video_count: number;
  total_resource_count: number;
  estimated_work: number;
  direct_videos?: PreflightStats;
  playlists?: PlaylistPreflightStats;
  combined?: PreflightStats;
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
