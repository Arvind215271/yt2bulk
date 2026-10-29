const getApiUrl = () => {
  return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
};

export interface ChannelRequest {
  url: string;
  provider?: string;
  output?: 'id' | 'url' | 'both';
  limit_mode?: 'all' | 'first' | 'range';
  limit?: number;
  range_start?: number;
  range_end?: number;
}

export interface VideoEntry {
  id: string;
  url: string;
  position: number;
  title: string;
  duration: number;
}

export interface ChannelVideosResponse {
  channel_id: string;
  total: number;
  videos: VideoEntry[];
}

export async function fetchChannelVideos(request: ChannelRequest, signal?: AbortSignal): Promise<ChannelVideosResponse> {
  const apiUrl = getApiUrl();
  console.log('[ChannelService] fetchChannelVideos invoked', { apiUrl, request });

  try {
    const res = await fetch(`${apiUrl}/channel/videos`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
      signal
    });

    const text = await res.text();

    if (!res.ok) {
      throw new Error(text || 'Failed to fetch channel videos');
    }

    return JSON.parse(text);
  } catch (error: any) {
    if (error.name !== 'AbortError') {
      console.error('[ChannelService] Error fetching channel videos:', error);
    }
    throw error;
  }
}
