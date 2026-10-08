# Preflight EstimatedWork and Subtitle Fetch Calculation Specification

## Overview
Refine the `EstimatedWork` and `SubtitleFetch` calculation to accurately reflect the work required for both direct video inputs and playlist items, while accounting for cache presence.

## Current Logic
- `EstimatedWork` = `TotalResourceCount` (sum of explicit videos + playlist `ItemCount`).
- `SubtitleFetch` = `len(uniqueVideos)` if subtitles are needed.
- Cache check is performed only for explicit `videoIDs`. Playlist items are not individually checked against the cache.

## Proposed Logic

### 1. `EstimatedWork` Calculation
`EstimatedWork` = `ExplicitVideosMetadataMissing` + `ExplicitVideosSubtitleFetchNeeded` + `PlaylistMetadataFetchNeeded` + `PlaylistSubtitleFetchNeeded`

#### Components:
- **`ExplicitVideosMetadataMissing`**: Count of unique explicit videos NOT in cache (`!s.cacheManager.HasMetadata(id)`).
- **`ExplicitVideosSubtitleFetchNeeded`**: Count of unique explicit videos needing subtitles AND not cached.
- **`PlaylistMetadataFetchNeeded`**: Total `ItemCount` for all playlists (assuming all items are absent as a conservative baseline).
- **`PlaylistSubtitleFetchNeeded`**: Total `ItemCount` for all playlists (if subtitles are needed).

### 2. `SubtitleFetch` Calculation
`SubtitleFetch` = `ExplicitVideosSubtitleFetchNeeded` + `PlaylistSubtitleFetchNeeded`

## Implementation Steps
1.  **Refactor `Estimate`**: Update `Estimate` to accept `planner` and `cacheManager` to accurately calculate `MetadataMissing` and `SubtitleFetch` for explicit videos.
2.  **Refactor `resolveResources`**: Keep it responsible for collecting playlist metadata (`ItemCount`).
3.  **Update `EstimateResource` / `Check`**: Combine these to calculate the final `EstimatedWork` and `SubtitleFetch` according to the proposed logic.
4.  **Update `EstimationResult`**: Ensure it holds enough data to compute the total `EstimatedWork`.
