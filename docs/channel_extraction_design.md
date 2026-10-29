# Channel Metadata Extraction Design Fix

## Issue Analysis
The current `flattenEntries` function in `internal/youtube/metadata/ytdlp/provider.go` performs indiscriminate recursive flattening of all nested playlists and tabs returned by `yt-dlp`. This causes:
1.  **Tab Merging:** Multiple tabs (e.g., "Videos", "Shorts", "Live") are flattened into a single list, losing their logical separation.
2.  **Metadata Loss:** Since the recursion flattens containers (tabs/playlists) into the same list as videos, context for what a video belongs to is lost.
3.  **Parsing Gaps:** Duration and other metadata might be missing or inconsistent if the flattening processes sub-entries that are not fully expanded or misclassified due to the mixed-entry structure.

## Proposed Design
Instead of naive recursive flattening, adopt a structured approach:

1.  **Entry Type Handling:**
    *   `YTPlaylistEntry` with `Type == "playlist"` or `Type == "tab"` should be treated as a container.
    *   `YTPlaylistEntry` with `Type == "video"` (or `url`) should be treated as a terminal node and processed for `youtube.PlaylistItem` mapping.

2.  **Refined Flattening Strategy:**
    *   Introduce a depth-aware `flattenEntries` or explicitly filter which entries to flatten at the top-level (e.g., look for the "Videos" tab entry and flatten only that).
    *   If full flattening is required, ensure that container metadata is not mixed with video metadata, and if a video is found, ensure its duration/id are correctly parsed from the `yt_dlp` response.

3.  **Specific Requirements:**
    *   `duration`: Ensure `entry.Duration` is safely cast to `time.Duration`.
    *   `video vs shorts`: Distinguish between shorts and regular videos if needed, perhaps by checking `entry.Duration` or title keywords, or using the specific tab if available.

4.  **Implementation Steps (for future phase):**
    *   Modify `GetChannelItems` in `internal/youtube/metadata/ytdlp/provider.go` to accept a tab-selection mechanism.
    *   Update `flattenEntries` to handle recursion base cases (e.g., only recurse if container type is allowed).
    *   Improve error handling and logging during the extraction process to identify missed or malformed entries.
