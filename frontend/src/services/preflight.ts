export interface PreflightRequest {
  inputs: string[];
  player_client?: string;
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

export async function checkPreflight(inputs: string[], player_client: string, signal?: AbortSignal): Promise<PreflightResponse> {
  const apiUrl = getApiUrl();
  console.log('[PreflightService] checkPreflight invoked', { apiUrl, inputs, player_client });
  
  try {
    const res = await fetch(`${apiUrl}/preflight`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ inputs, player_client }),
      signal
    });

    const text = await res.text();
    console.log('[PreflightService] fetch response received', { status: res.status, ok: res.ok, text });
    
    if (!res.ok) {
      let errorMessage = text || 'Failed to perform preflight check';
      let errorDetails: any = null;
      try {
        const errorJson = JSON.parse(text);
        
        console.log('[PreflightService] Parsing error JSON', { errorJson });

        // The backend `ErrorResponse` structure:
        // { success: false, message: string, data: FieldError[] | null, meta: { code: string } }
        // where `FieldError` is { field: string, message: string }
        
        if (errorJson && errorJson.data) {
          // If `data` exists, it's our structured field-level errors
          errorDetails = errorJson.data;
        } else {
          // Fallback: if no `data`, try parsing the message string for details
          if (errorJson && errorJson.message) {
            errorMessage = errorJson.message;
            const match = errorMessage.match(/reached (\d+), max allowed (\d+)/);
            if (match) {
              errorDetails = [
                { field: "reached", message: match[1] },
                { field: "max", message: match[2] }
              ];
            }
          }
          
          if (!errorDetails) {
            errorDetails = errorJson;
          }
        }
      } catch (e) {
        console.warn('[PreflightService] Failed to parse error response as JSON', e);
      }
      
      console.error('[PreflightService] Caught non-OK response', {
        status: res.status,
        message: errorMessage,
        details: errorDetails
      });
      
      const err = new Error(JSON.stringify({ message: errorMessage, data: errorDetails }));
      (err as any).details = errorDetails;
      throw err;
    }

    return JSON.parse(text);
  } catch (error: any) {
    if (error.name === 'AbortError') {
      console.log('[PreflightService] Request aborted');
    } else {
      console.error('[PreflightService] Caught error in checkPreflight fetch/processing:', error);
    }
    throw error;
  }
}
