
[`YTOutput`](internal/youtube/metadata/ytdlp/model.go:4): `type YTOutput struct { ID, Title, FullTitle, Description string; Duration int; ViewCount, LikeCount, CommentCount uint64; Channel, ChannelID, Thumbnail string; Tags, Categories []string; UploadDate, WebpageURL, Uploader, UploaderID, UploaderURL, Availability, MediaType string; AgeLimit int; Extractor, DisplayID, Filename, Ext string; Width, Height int; Resolution string; FPS int; VCodec, ACodec string; AudioChannels int; Language string; Subtitles, AutomaticCaptions map[string][]SubtitleFormat; Chapters []map[string]interface{}; Thumbnails []ThumbnailEntry }`

[`SubtitleFormat`](internal/youtube/metadata/ytdlp/model.go:44): `type SubtitleFormat struct { Ext, URL, Name, Protocol, Client string }`
[`YTPlaylistOutput`](internal/youtube/metadata/ytdlp/model.go:52): `type YTPlaylistOutput struct { ID, Title, Description, Channel, ChannelID, Thumbnail string; Tags []string; WebpageURL, Uploader string; Entries []YTPlaylistEntry }`
[`YTPlaylistEntry`](internal/youtube/metadata/ytdlp/model.go:65): `type YTPlaylistEntry struct { ID, Title, Description string; Duration int; ViewCount uint64; ChannelID, Channel, ChannelIDAlt, WebpageURL, OriginalURL, Thumbnail string; Thumbnails []ThumbnailEntry }`
[`ThumbnailEntry`](internal/youtube/metadata/ytdlp/model.go:80): `type ThumbnailEntry struct { URL string; Width, Height, Preference int; ID, Resolution string }`

### Preflight Data Models & Statistics Dependencies
[`internal/youtube/preflight/model.go`](internal/youtube/preflight/model.go):
- `type Request struct { Inputs []*youtubeurl.YouTubeInput }`
- `type Result struct { Valid bool; Error error; UniqueVideoCount, TotalResourceCount, EstimatedWork int; Stats PreflightStats }`
- `type Service interface { Check(ctx context.Context, req Request) (Result, error) }`
- `type Estimator interface { Estimate(ctx context.Context, videoIDs []string, fieldPlanner *fields.Planner) (EstimationResult, error); EstimateResource(ctx context.Context, resources []youtubeurl.YouTubeInput, fieldPlanner *fields.Planner) (EstimationResult, error) }`
- `type EstimationResult struct { Plan *planner.ExecutionPlan; MetadataMissing, MetadataPresent, SubtitleFetch int }`

[`internal/youtube/preflight/stats.go`](internal/youtube/preflight/stats.go):
- `type PreflightStats struct { MetadataAttempts, MetadataSuccesses, MetadataFailures, MetadataFallbacks int }`
- `func (s *PreflightStats) Add(other PreflightStats) { s.MetadataAttempts += other.MetadataAttempts; s.MetadataSuccesses += other.MetadataSuccesses; s.MetadataFailures += other.MetadataFailures; s.MetadataFallbacks += other.MetadataFallbacks }`

[`internal/youtube/stats/model.go`](internal/youtube/stats/model.go):
- `type MetadataStats struct { Attempts, Successes, Failures, CacheHits, CacheMisses, UpstreamFetches, EstimatedUpstreamFetches, YTDLPRequests, APIRequests, Fallbacks int }` and `func (s *MetadataStats) Add(other MetadataStats)`
- `type SubtitleStats struct { Attempts, Successes, Failures, CacheHits, CacheMisses, UpstreamFetches, PotentialSubtitleFetches, Fallbacks, ManualRequests, AutomaticRequests int }` and `func (s *SubtitleStats) Add(other SubtitleStats)`
- `type PlaylistStats struct { Attempts, Successes, Failures, ItemsDiscovered, Fallbacks int; Metadata MetadataStats }` and `func (s *PlaylistStats) Add(other PlaylistStats)`
- `type PipelineStats struct { Metadata MetadataStats; Subtitle SubtitleStats }` and `func (s *PipelineStats) Add(other PipelineStats)`
- `type ResourceStats struct { VideosProcessed, VideoFailures int; Metadata MetadataStats; Subtitle SubtitleStats }` and `func (s *ResourceStats) Add(other ResourceStats)`
- `type ProcessStats struct { ResourcesRequested, ResourcesSucceeded, ResourceFailures, VideosProcessed int; Metadata MetadataStats; Subtitle SubtitleStats }` and `func (s *ProcessStats) Add(other ProcessStats)`

### Function Mappings
- [`Check()`](internal/youtube/preflight/service.go:106): Input [`Request`](internal/youtube/preflight/model.go:12), Output [`Result`](internal/youtube/preflight/model.go:17), error. Validates inputs, resolves resources to IDs, calculates overall estimation.
- [`Estimate()`](internal/youtube/preflight/service.go:29): Input `[]string`, `*fields.Planner`, Output [`EstimationResult`](internal/youtube/preflight/model.go:38), error. Generates execution plan, checks cache for metadata, estimates subtitle needs.
- [`EstimateResource()`](internal/youtube/preflight/service.go:72): Input `[]youtubeurl.YouTubeInput`, `*fields.Planner`, Output [`EstimationResult`](internal/youtube/preflight/model.go:38), error. Resolves mixed input types, deduplicates IDs, delegates to `Estimate()`.
- [`resolveResources()`](internal/youtube/preflight/service.go:148): Input `[]*youtubeurl.YouTubeInput`, Output `[]string`, `int`, [`PreflightStats`](internal/youtube/preflight/stats.go:4), error. Iterates inputs, fetches playlist items, collects stats/video IDs.
- [`estimateOverall()`](internal/youtube/preflight/service.go:138): Input `[]string`, Output [`EstimationResult`](internal/youtube/preflight/model.go:38), error. Initializes default [`fields.Planner`](internal/youtube/fields/model.go), calls `Estimate()`.
- [`Service.GetPlaylistMetadata()`](internal/youtube/metadata/service.go:60): Input `playlistID`, `ProviderType`, Output `youtube.Playlist`, error. Wrapper around `GetPlaylistMetadataWithStats()`.
- [`Service.GetPlaylistItems()`](internal/youtube/metadata/service.go:104): Input `playlistID`, `ProviderType`, Output `[]youtube.PlaylistItem`, error. Wrapper around `GetPlaylistItemsWithStats()`.
- [`Service.GetPlaylistItemsWithStats()`](internal/youtube/metadata/service.go:69): Input `playlistID`, `ProviderType`, Output `[]youtube.PlaylistItem`, [`stats.MetadataStats`](internal/youtube/stats/model.go:3), error. Orchestrates provider selection, handles fallbacks, updates statistics.
- [`ytdlp.Provider.GetPlaylistMetadata()`](internal/youtube/metadata/ytdlp/provider.go:34): Input `playlistID`, `*stats.MetadataStats`, Output `youtube.Playlist`, error. Fetches playlist metadata via `ytdlp.Client`, maps to model.
- [`ytdlp.Provider.GetPlaylistItems()`](internal/youtube/metadata/ytdlp/provider.go:44): Input `playlistID`, `*stats.MetadataStats`, Output `[]youtube.PlaylistItem`, error. Fetches playlist via `ytdlp.Client`, maps entries.

Note: Relies on `metadata.Service` (`GetPlaylistItemsWithStats`) for playlists, `cache.Manager` for cache lookups, and `PreflightStats` mapped from upstream `MetadataStats`. Fallbacks occur at `Service` level if `YTDLP` fails and `API` is configured.











So what we want to do is change the preflight working.
Currently what it does is fetch the videoID for a playlist...

Whihc is soemthing that we do want. But for internal processing. Like..
What should happen is that. 

preflgiht have two logics.

1. It is used for estimation. aka how much proes or cost it will take to process seothig.
2. Second is when we add job model. here, we want to have certtain limit on metadata and playlist fetch per uesr... or per request type.
3. Thus, we cannot allow the pesron to fetch everythign at once. Aka a certain number of videos are only allowed to be fetched for ratelimit.

4. Usually what we want to do.. Is like make the service prone to user request. And thus enable request till it does upstream fethces of certani number.
5. But playlist might be very large like 5000 video. Thus it is like not needed to fetch.
6. Thus, if a playlist doesn't exist. What it should do is get the playlist metadata only (without fetching the actual videoID).
7. It would contain the numebr of vide that iexst in it...
8. Thus we can do it quicker, will not have to prcoess the playlist and be able to tell how much cost it would be (aussming that the playlist video do not exist in cache.. which they usally won't.)
9. If a playlist videoID exist... then it is usaully two case. 1 the video ID exist.. if that is the cse. We shoud inlcude its fetch. (only if the metdata for the videoID exist.. it mgiht be that ony videoId wre fetched but no metadat for those id. )
10. if it doesn't exist. that uauly means it's video won't be preesnt  (in most of the case viode ID won't be cathced. So intead of processing everything, we can sipmy fetch the amount of video that exist in the playlist without downlaoding the whole playlist...)
11. Thus a 5k video playlist won't hold the requst for like 1  minute. And we are just estimating it.

alright... Let us keep it simple.



I don't think this have to be that complicated.



1. We are couting cost for  metadata only.
2. The reason is that.. There are two cases: 1. metadata is missing. Then whole viode will be missing and we can say that subtile woulnd't exist as well. So +1, +1 to both.
3. Metadata is presnt, Then check if the prefererd subtitle is present... If yes, then we process that and check if that exist. Depeding on the case we do +0, +1 or +0, +0. If the preferred subtitle exist.
4. For playlist.. What is playlist acutally? Nothing more... a colleciton of video.
5. But there are two wys we can parse. If the whole playlist does not exist.. Then we can asume that both metadat, subtitle won't exist. SO +1, +1 to both...&#x20;
6. If it exist.. Then we would have to parse that noralmly.. aka chekc for each videoId that might exsit. and so on...
7. Thus... The wrokng is. If playlist exist. Then use that. Else we fetch the playlist metadata to know the number of vidoes inside it. So we can do direclt +amountOfVideo to both video and subtitle.



Now the important thing... The processing should be of 3 lvel.

1. DIRECT VIDOE ID from the user input. They will all be combined inot a single hieray.. video_preflight or soemthing like that. They store two details.. or maybe more. Total video, metadatacahed, imeetaata usptreat, subtitle cahced, cubsitle upstreatm.. This will have a type: direct video ID or osethig... aka these are from the video ID given by the user.  it wouldget an arya of videdoID. Wiht item count aka otal viode ID given.&#x20;
2. Playlist Level. All thee playlist will be will be it. So we have two decision. If the playlist does not exist. We have to fetch the plallist metadata. If it does then it can be procesed using direct viode ID. its type would be playlist videos id... Thus.. Agawin the proesing is defined above. if metadata is feethced.. Simplt add that to all the fields, metataupstream, and subtitle upstream... the item count is total playlist ID given (not total video inside it we found).
3. resoruce lvel.. What it does is onting.. just combine the two of them.. and cmpute for total video only..  This contain the item couunt.. we have to keep it sepeate.. or maube merge? but ig showing direvt Vioed, playlsti item, and total vidoe count can be used? Kinda..  Then it would contain the reducant data above aka both video, playlist item. but a third data as well. that is again stats but its tupe is combiend: so it contian the sum of both video and playlist item and noting more...

That is the goal...



So ig.. do that?

nope this is also wrong..

ig.. Tihs is how it should look.

type PreflightStats struct {
    type: video
    ItemCount        int
    MetadataCached   int
    MetadataUpstream int
    SubtitleCached   int
    SubtitleUpstream int
}

plalist

type: plalist
total_plallist:
total_video:
video_metadataChaced:
Videos_metadat upsterm...

And so on. 

type ResourcePreflight struct {
    DirectVideos PreflightStats
    Playlists    PreflightStats
    
Then a coabinon of all the videoMetadta that exist..
}

So acutal there are 3 struut.

One is for videos...
Then plyastlist one is notign but.. type: playlist, totalPlaulsit, vioe prefliht sturuct.
And then for rerouce we have vioe pre, palylist pre, combined (that is same as viode pre but having both added).

Thus.. what we do is ismply have two unciont. One for viode add. that simply add to vidoe.. And then for playlist.. we can incrase the count manually ig? Or smething liek tht? insside it? kinda true...

That is the goal ig? kinda...







Yep. Let's lock the **architecture/workstream plan first**, without writing implementation files yet.

I’d keep it to **4 small workstreams**:

1. **Preflight Model** — define the 3 structs and their exact relationships.
2. **Video Preflight** — implement the common per-video accounting logic.
3. **Playlist Preflight** — handle playlist existence, enumeration, and aggregate its videos using the video logic.
4. **Resource Preflight** — process direct videos + playlists and build the combined totals.

The important dependency is:

```text
01 Model
   ↓
02 Video Preflight
   ↓
03 Playlist Preflight
   ↓
04 Resource Preflight
```

And the core architecture is:

```text
                    ResourcePreflight
                   /        |        \
                  /         |         \
        DirectVideos     Playlists    Combined
             │               │            │
             ▼               ▼            ▼
      VideoPreflight  PlaylistPreflight  VideoPreflight
                              │
                              ▼
                       VideoPreflight
```

I'll keep each workstream narrowly scoped and make the implementation workers follow **exact files/functions/behavior**, with no architectural decisions left for them.
























'''
# WORKSTREAMS.MD FORMAT SPEC

## Purpose & Hard Rules
- Parser converts `workstreams.md` into a directory tree.
- **HARD LIMIT:** Every generated `.md` file MUST contain <= 250 words. The parser rejects anything above it.
- **No LLM Creativity:** Designed for 1.5B–3B models on 64k context. Zero critical thinking or independent planning allowed. All files, target paths, function signatures, and test cases must be fully preplanned.

## Format & Syntax
# 0xx - Workstream Title
(Text directly below becomes `workstream.md` describing WHAT it is, <= 250 words).

## objectives.md
(Numbered list of core steps).

<01-folder-name>
### objective.md
### 01-discovery.md
#### 00-architect.md
#### code.md
#### test.md
#### verification.md
</01-folder-name>

## commit.md
(Standard conventional commit message and summary).

### Heading Depth Rules
- Root file: `##`
- 1-level folder: `###`
- 2-level folder: `####`
- 3-level folder: `#####`

---

## Complete Canonical Example

# 0xx - Resource Enumeration
The system must enumerate and validate YouTube resources safely prior to execution while enforcing strict global deduplication and hard limits.

## objectives.md
1. Perform discovery audit on input and metadata services.
2. Define architecture and contract for resource enumeration.
3. Implement core enumeration logic with deduplication and limits.
4. Add comprehensive test suite and verification gates.

<01-resource-enumeration>

### objective.md
## Execution Checklist
1. [ ] **OBJ-001: Discovery Audit**
   - TARGET: Inspect `internal/youtube/input/` & metadata service.
   - BOUNDARY: Read-only. No code changes.
2. [ ] **OBJ-002: Architecture & Contract**
   - TARGET: Define `ResourceEnumeration` & global deduplication rules.
   - BOUNDARY: No processing pipelines, Redis, or HTTP handlers.
3. [ ] **OBJ-003: Core Implementation**
   - TARGET: Create `internal/preflight/enumerator.go` (Max 120 lines).
   - BOUNDARY: Use `seenVideos` map and abort if > `MAX_JOB_ITEMS`.
4. [ ] **OBJ-004: Verification**
   - TARGET: Add tests for duplicates, playlist overlaps, & cancellation.

## Hard Constraints
- **File Cap:** Max 150 lines per file.
- **Forbidden:** Never import `pipeline.Service` or `Job Manager`.
- **Deduplication:** Global across all inputs, never per-input concatenation.

#### 00-architect.md
## 1. Component Placement & Boundaries
- **Target Package:** `internal/preflight/` (New directory).
- **Allowed Dependencies:** `internal/youtube/input/` & `internal/youtube/metadata.Service`.
- **Forbidden:** No pipelines, no Redis, no HTTP handlers, no direct DB access.

## 2. Design Rationale
- Decouple ingestion from execution to ensure zero cost/usage side effects.
- Enforce global deduplication using a centralized map to maintain $O(N)$ unique video efficiency.

## 3. Action Required: Update `code.md`
- Append exact file target (`internal/preflight/enumerator.go`), `seenVideos` map, and `MAX_JOB_ITEMS` abort rule into `code.md` without wiping content.

#### 02-discovery.md
## Objectives
Inspect repository reality to identify existing file paths, ingestion patterns, and metadata services.
- **Input Parsing:** Under `internal/youtube/input/`.
- **Metadata Boundaries:** Under `internal/youtube/metadata/` (`GetPlaylistItemsWithStats`).
- **Action:** Append paths and signatures directly into `code.md`.

#### code.md
## Function_name: `Enumerate`
- **Inputs:** `ctx context.Context`, `inputs []YouTubeInput`, `limit int`
- **Outputs:** 
- **Dependencies:** `internal/youtube/input/` and `internal/youtube/metadata.Service`.

## Implementation Blueprint (Comment-to-Code Mapping)
1. `// Check context cancellation` -> If `ctx.Err() != nil`, return error.
2. `// Initialize global tracking maps` -> Create `seenVideos := map[string]struct{}{}`.
3. `// Process input list iteratively` -> Loop `inputs`. Add video ID or call metadata service.
4. `// Enforce global limits and deduplicate` -> Abort if unique count exceeds `limit`.

## Edge Cases & Expected IO
- **Empty Inputs:** Return empty `ResourceEnumeration` with zero error.
- **Duplicate Direct Videos:** Globally deduplicated to 1 unique item.
- **Zero Imagination:** Translate blueprint directly into syntax.

#### test.md
## Target Function: `Enumerate` (`internal/preflight/enumerator_test.go`)
## Implementation Blueprint (Comment-to-Code Mapping)
1. `// Test empty and nil input handling` -> Assert empty slice returns zero errors.
2. `// Test global deduplication for direct videos` -> Assert unique IDs only.
3. `// Test playlist item expansion and deduplication` -> Mock metadata service overlap.
4. `// Test max limit abort behavior` -> Assert immediate error return when over limit.
5. `// Test context cancellation propagation` -> Assert pre-cancelled context returns error.

#### verification.md
## Target & Execution Gates
1. **Test Suite:** Run `go test -v ./internal/preflight/...` with zero failures/races.
2. **Boundary Audit:** Confirm zero forbidden imports.
3. **Failure Protocol:** Abort pipeline immediately on any error or boundary violation.

</01-resource-enumeration>

## commit.md
feat add metadata file cache and pipeline planner
- Implement file-based caching for yt-dlp with TTL and capacity startup cleanup
- Introduce field planner and export pipeline for validated field selection and stage mapping
- Clean up error handling by removing appError fields and standardizing on standard errors
- Add default configuration and validation logic for word stats
'''














I'd use **6 workstreams (000–005)** for the complete backend and frontend implementation. That keeps each workstream small enough for a 1.5B–3B model without splitting straightforward work into unnecessary pieces.

| Workstream | Name | Scope |
|---|---|---|
| `000` | Preflight Data Model | Define the three structures: direct videos, playlists, and combined resource statistics. |
| `001` | Video Preflight | Implement metadata/subtitle cache checks and per-video cost accounting. |
| `002` | Playlist Preflight | Use cached playlist metadata when available; otherwise fetch only playlist metadata and estimate video costs from its count. |
| `003` | Resource Preflight Integration | Combine direct videos and playlist statistics, integrate the new model into `Check`, `Estimate`, and `EstimateResource`, and preserve validation behavior. |
| `004` | Backend API Integration | Update the preflight response model, handler, serialization, and backend tests to expose the new accounting. |
| `005` | Frontend Preflight Integration | Update frontend response types, display direct-video counts, playlist counts, total videos, metadata cache/upstream counts, and subtitle cache/upstream counts. |

### Dependency order

```text
000 Preflight Data Model
          |
          v
001 Video Preflight
          |
          v
002 Playlist Preflight
          |
          v
003 Resource Preflight Integration
          |
          v
004 Backend API Integration
          |
          v
005 Frontend Preflight Integration
```

### Important boundaries

- **Workstreams 000–003:** Core preflight logic.
- **Workstream 004:** Backend API contract and response integration.
- **Workstream 005:** Frontend types, presentation, and integration tests.

Keep each workstream in the exact Markdown format you supplied, with every generated `.md` file under 250 words. Each should specify exact files, functions, implementation behavior, test cases, and verification commands.

**One important detail:** Workstream `002` must not enumerate thousands of playlist video IDs when playlist metadata is missing. It should retrieve the playlist's metadata and use its video count to estimate the required work. When cached playlist metadata is available, it should inspect the available video IDs and perform the normal per-video cache accounting.

This gives you **6 workstreams total**, without creating separate workstreams for trivial helpers or individual tests.

