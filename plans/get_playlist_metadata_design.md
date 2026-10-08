# Architecture Design: GetPlaylistMetadata

## Goal
Implement `GetPlaylistMetadata` to route playlist metadata requests through the existing `WorkerManager` and `SingleClient` without creating new pools.

## Design
1.  **Protocol**: Reuse `PlaylistRequest` in `internal/youtube/metadata/ytdlp/python/protocol.go` to send a `playlist_metadata` command to the python worker.
2.  **SingleClient (`internal/youtube/metadata/ytdlp/python/client.go`)**:
    *   Add `GetPlaylistMetadata(ctx context.Context, playlistID string) (map[string]any, error)`.
    *   This will construct a `PlaylistRequest` with `Cmd: "playlist_metadata"` and send it to the worker.
    *   Reuse the existing communication logic (send, receive, unmarshal, handle errors).
3.  **WorkerManager (`internal/youtube/metadata/ytdlp/python/manager.go`)**:
    *   Add `GetPlaylistMetadata(ctx context.Context, playlistID string) (map[string]any, error)`.
    *   This will create a `request` for the queue, utilizing the new `SingleClient.GetPlaylistMetadata` method.
    *   This reuses the existing worker loop, request queuing, and error handling (replacement) logic.
4.  **Python Worker (`internal/youtube/metadata/ytdlp/python/python_worker_single.py`)**:
    *   The `main` loop already contains a `playlist_metadata` command handler (seen in `python_worker_single.py` lines 158-200) that performs `extract_info` with `extract_flat=True` and `playlist_items='0'`. This matches the requirement.

## Verification
*   Implement unit tests in `manager_test.go` (or similar) to verify the flow through `WorkerManager` to `SingleClient` and ensure `playlist_metadata` command is sent to the worker.
