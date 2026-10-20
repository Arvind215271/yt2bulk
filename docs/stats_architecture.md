# Architecture: Processing Statistics and Accounting Layer

This document outlines the architecture for the internal processing statistics and accounting layer in Yt2bulk.

---

## 1. Current Execution Flow

The current call graph flows hierarchically from resource processing down to individual metadata/subtitle providers:

```text
resource.Service.ProcessResources()
    ├── ProcessVideo()
    │     └── pipeline.Service.ProcessFaultTolerant()
    │           └── ProcessResource()
    │                 ├── ProcessMetadata() -> metadata.Service.GetVideo()
    │                 ├── ProcessDescription() -> description.Service
    │                 └── ProcessSubtitleAndTranscript() -> subtitle.SubtitleService.DownloadSubtitle()
    │
    └── ProcessPlaylist()
          ├── metadata.Service.GetPlaylistItems() (yt-dlp -> fallback to API)
          └── [confluent/parallel] ProcessResource() for each playlist item (via pipeline.Service)
```

---

## 2. Recommended Stats Hierarchy

We introduce a dedicated shared package:
`internal/youtube/stats/`

Containing the following minimal types:

- **`MetadataStats`**: Tracks attempts, successes, failures, cache hits/misses, upstream fetches, yt-dlp requests, API requests, and fallbacks.
- **`SubtitleStats`**: Tracks attempts, successes, failures, cache hits/misses, upstream fetches, fallbacks, manual requests, and automatic requests.
- **`PlaylistStats`**: Tracks items discovered, attempts, successes, failures, fallbacks, and embedded metadata stats for playlist lookup.
- **`PipelineStats`**: Aggregates `MetadataStats` and `SubtitleStats`.
- **`ResourceStats`**: Aggregates videos processed, video failures, metadata stats, and subtitle stats (or collection of video/playlist pipeline stats).
- **`ProcessStats`**: Top-level aggregation of all resources processed.

---

## 3. Exact Counting Semantics

- **Attempt**: Incremented whenever an operation is invoked.
- **Success**: Incremented when an operation completes without error.
- **Failure**: Incremented when an operation returns an error.
- **CacheHit**: Incremented when data is successfully retrieved from cache without downstream/upstream call.
- **CacheMiss**: Incremented when cache lookup results in a miss or error.
- **UpstreamFetch**: Incremented when a real external call (yt-dlp or YouTube API) is performed.
- **YTDLPRequests / APIRequests**: Incremented per provider invocation.
- **Fallbacks**: Incremented when a primary provider fails and a fallback provider is attempted.

---

## 4. Cache Instrumentation Point

Cache hits and misses are checked and known within the respective provider/service implementations (e.g., inside `subtitle.SubtitleService` and metadata providers/clients). The client/provider checks the cache manager and increments `CacheHits` or `CacheMisses` explicitly.

---

## 5. Provider Instrumentation Point

Metadata provider implementations (`ytdlp` and `youtube_api`) or `metadata.Service` record provider requests (`YTDLPRequests`, `APIRequests`, `UpstreamFetches`) when executing real network/subprocess requests.

---

## 6. Subtitle Instrumentation Point

`subtitle.SubtitleService` checks cache (`CacheHits` / `CacheMisses`), attempts downstream download via Python worker pool (`UpstreamFetches`), validates/selects tracks, and handles fallback/requested vs actual types.

---

## 7. Propagation API

To avoid breaking existing callers, we introduce `WithStats` method variants alongside existing methods, where existing methods act as thin wrappers ignoring/discarding stats.
- `GetVideoWithStats`
- `GetPlaylistItemsWithStats`
- `DownloadSubtitleWithStats`
- `ProcessFaultTolerantWithStats`
- `ProcessResourcesWithStats`

---

## 8. Concurrency model

In concurrent processing loops (e.g., playlist item processing and resource processing), each worker goroutine accumulates stats locally and returns them via channels/return values. Parent layers aggregate worker stats using deterministic `.Add()` methods, avoiding shared mutable state and race conditions.

---

## 9. Package Placement

Stats models and helper functions (`Add` / `Merge`) live in `internal/youtube/stats/` to prevent circular dependencies between metadata, subtitle, pipeline, and resource packages.

---

## 10. Explicit Non-Goals

This architecture explicitly does **NOT** implement:
- Rate limiting or quotas
- Billing or cost calculation
- Job persistence or DB storage
- User-level usage accounting
- API response changes
- Telemetry/observability frameworks (e.g., OpenTelemetry)

---

## 11. Implementation Breakdown (7 Atomic Tasks)

1. **Task 1**: Define stats model in `internal/youtube/stats/` with unit tests.
2. **Task 2**: Instrument metadata layer with `WithStats` variants and thorough tests.
3. **Task 3**: Instrument subtitle layer with `DownloadSubtitleWithStats` and thorough tests.
4. **Task 4**: Propagate stats through pipeline via `ProcessFaultTolerantWithStats`.
5. **Task 5**: Propagate stats through resource layer (`ProcessResourcesWithStats`, `ProcessPlaylistWithStats`, etc.).
6. **Task 6**: End-to-end accounting tests covering multi-resource, playlists, cache hits, and fallbacks.
7. **Task 7**: Final architecture & reviewer verification pass.

---

## 12. Review Failure Rule

If any task review fails during verification, it must be reworked immediately. No failed task may be marked as complete.
