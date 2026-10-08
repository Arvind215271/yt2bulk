   \### 1. Stats (Surface Level)\
   Defined in [\`internal/youtube/stats/model.go\`]\(internal/youtube/stats/model.go:1), the stats components provide hierarchical metrics aggregation through structs like [\`MetadataStats\`]\(internal/youtube/stats/model.go:3), [\`SubtitleStats\`]\(internal/youtube/stats/model.go:27), [\`PlaylistStats\`]\(internal/youtube/stats/model.go:51), and [\`ProcessStats\`]\(internal/youtube/stats/model.go:93). These track operational attempts, successes, failures, cache hits/misses, upstream fetches, and fallback frequencies across the pipeline.\
   \
   \### 2. Planner Package (Full)\
   Managed via [\`internal/youtube/planner/planner.go\`]\(internal/youtube/planner/planner.go:1) and [\`internal/youtube/planner/builder.go\`]\(internal/youtube/planner/builder.go:1), the planner analyzes input resources and field selections to construct an [\`ExecutionPlan\`]\(internal/youtube/planner/builder.go:1). It determines which pipeline stages are required (e.g., metadata, descriptions, subtitles, transcripts, or signals) to optimize and customize the execution flow.\
   \
   \### 3. Cache Manager and Its Working\
   Implemented in [\`internal/cache/manager.go\`]\(internal/cache/manager.go:1), the [\`Manager\`]\(internal/cache/manager.go:18) handles persistent caching of metadata, playlists, and subtitles on disk. \
   \- [\`Get()\`]\(internal/cache/manager.go:31) and [\`Save()\`]\(internal/cache/manager.go:43) read and write files using structured keys (e.g., [\`VideoKey()\`]\(internal/cache/manager.go:51), [\`SubtitleKey()\`]\(internal/cache/manager.go:59)).\
   \- [\`Cleanup()\`]\(internal/cache/manager.go:75) enforces TTL-based expiration and capacity-based file eviction (keeping 90% of \`MaxFiles\`).\
   \
   \### 4. Fetching Playlist Content\
   Handled primarily in [\`internal/youtube/resource/playlist.go\`]\(internal/youtube/resource/playlist.go:1) via [\`ProcessPlaylistWithStats\`]\(internal/youtube/resource/playlist.go:31), content fetching invokes [\`metadataService.GetPlaylistItemsWithStats\`]\(internal/youtube/metadata/service.go:1), trying [\`metadata.ProviderYTDLP\`]\(internal/youtube/metadata/provider_type.go:1) first and falling back to [\`metadata.ProviderAPI\`]\(internal/youtube/metadata/provider_type.go:1) upon failure. It maps items to videos, checks execution plans, and dispatches concurrent worker tasks if pipeline processing is needed.\
   \
   \### 5. Resource Component (Goal)\
   Implemented in [\`internal/youtube/resource/service.go\`]\(internal/youtube/resource/service.go:1), the resource service coordinates batch resource processing through [\`ProcessResourcesWithStats\`]\(internal/youtube/resource/service.go:52). It processes a slice of [\`youtubeurl.YouTubeInput\`]\(internal/youtube/input/model.go:1) items (videos and playlists) concurrently using bounded concurrency (semaphores), strictly preserves request order, handles context cancellations, and returns uniform [\`youtube.Resource\`]\(internal/youtube/export/model.go:1) containers along with aggregated [\`stats.ProcessStats\`]\(internal/youtube/stats/model.go:93).

   This is the current working of stast.

   Now we want to build preflight.

   The goal is to use stats model (mostly).Or crate a similar moel based on it.
   Then what it shoudl do is for video. get the palnner based on the feild erquisr. Then for those. it should chek if its metadta and subtitle eixst or not.. based on the pallning.If paln need sutitle. it should only count that then. Else not.. usually all plan need metadta ataslte..

   For subtile we need prefernce thing ig.. Also subtitle and metdata mgiht have achahce magner builndg there. Or in he cahcne manger itelf ig..

   For playlist.. Currently it should fetch the playlist. That is asked.
   return the playlist id links that are needed.
   And thus, then we cacn chekc if the neded chahee manger have those vidoe, subtitle or not.. That is the goal ig? Kinda...

   But ig this is knida bad. We cannot really acount for subtitle wihtout metadata..

   So it should only estiamte eor metadata only.. for the request video ID and should show many exist..

   uuslaly.. we can tho divide this inot many video and shit.. but we should noot.. ig.? it should colectively do it for video only... and not make another json{} for playlist ig? or maybe. it process normal video in one metadat and for playlist create nested one?

   Maybe we want to funcoitn. one should proces only video. Aka it doesn't care where that comes from playlist or video. Thus that is the goal.

   Second is a nested way. it shod create that for each item type: video, playlist. Akd it should create stats for it.
   if it is a video. it should show wheter it is prenet or not.. and all the toher detials that are needed here ornot.

   And for like... Uhh playlist. what it shoidl do.. is crate that same thing. but for the viodeId inie. aka for a single vido. is that is prenset or not...

   That is the goal... The goal for preflgiht is to check for the estiamted count. for upstream fetches..

   Wait.. so the goal is. If the subtitle is planned to be fund. it should exismate that for all the subtiel a  fetch might be reuqist.. thus. It should also tell the whole subtitle fetch usage if it exist there.. elsse not..

   Thus should be anther field.. aka oucnt totla video id where inside or outside. Then wht it shoudl do is do if.. if the subtiel field is acive. Thene ti should show additoanl cost of: video_Count for sutitle might be needed...  (if efethicng subtiele)...

   or somethign lie that...

   That is the goal? ig..

   Bit this becomes weird> ig..

   I think. we shod have one version that do for all video ID. Then another veriosn should not be for each requiest type and shit.. Aka if it is a resrouce item (eithr a playlist or video). What it shodl do.. is get the one versio.. (aka for all the video ID). if a video item  isrequest. It should simpy call for [array of 1 video] to the video ID funciotn. Then for playlist. it shoudl send an aryra of viode ID to the funcoitn.. and get a single stats only.. So it i kinda compresed and siht... For all indvualy video inside the plaulist? maybe we can have all the funcoint as well?

   Aka one is surface levl. Second is surface level but cout ffor all the reousrce time. And third is video lel.. it shoud accunt for video Id and not rersource.. aka it should crate a nested sruruce ig&#x20;
   Then... when it is done. it shoud work with chache maner to check if those exist or not..

   for subtitle.. we do not really have a good way to esimate ig.. The problem is that it require the type of automcia or manul and theen the lanuge type deepnign.. and ths it is harder to do it wiht metadata only.. this it shuld not add the cost for sutie. but should methoion hta tif subtile are fethced the esitamed cost is around all the video ID (explain the raeon that it cnanot know what subtitle are preetn in the metadata withhout atcualt ethc...).. Maybe.. it shod do.. Like? if a metadaata exist.. it shoudl proces it? Chek if the reuqet subtitle type eist or not? but i don't think so that is valid here.. as stil the person would need to tel  the tpe. and tat is a weird sysntex for it.. and to tell. So

   ig it should esitmate only metdata only..And for ssubtitel it should tell an esismtea cost euqla to the numer of videoID exist in the whole processing..

   Mabye that is the goal...

   So the stats should be.. like.. the procesing aka for all the ie. it shoudl cout for total videio id, plalist id and allthe infomratoin at the top leel.

   Then it shoudl have the data for the funciont and ssit. dpeneind  on whihc funciont call was request.. and that is the case...



# Preflight Architecture

Preflight is an **estimate-only planning layer** built around the existing `stats`, `planner`, and `cache` systems. It must never fetch metadata, subtitles, transcripts, or process videos. The only upstream operation allowed is playlist enumeration when required to resolve its video IDs.

## Core Design

The architecture has three levels:

1. **Video-level estimation** — the core reusable operation. It accepts a collection of video IDs and requested fields, independent of where those IDs came from. It builds the existing `ExecutionPlan`, determines required stages, checks the metadata cache, and aggregates metadata availability and estimated upstream fetches. If subtitles are planned, it reports **potential subtitle fetches equal to the relevant video-ID count**, because exact subtitle availability cannot be determined without actually resolving subtitle tracks and their language/type/fallback rules.

2. **Resource-level estimation** — accepts parsed video/playlist resources. A direct video contributes one video ID. A playlist is enumerated to obtain its video IDs. All IDs are globally deduplicated, then passed collectively to the video-level estimator. A playlist therefore receives one aggregated video-level estimate rather than independently processing every video.

3. **Overall preflight result** — combines request-level resource counts with the aggregated video-level statistics. It can additionally retain resource-level/nested information where useful, but must not duplicate estimation logic.

## Statistics

Reuse `internal/youtube/stats` wherever possible, especially `MetadataStats` and `SubtitleStats`. Add only fields genuinely required to represent estimated/potential work; do not create a parallel runtime-statistics hierarchy.

Metadata estimates are cache-aware: cached videos are hits, missing/expired entries contribute to estimated upstream fetches. Subtitle estimates are explicitly marked as potential rather than definite.

## Flow

`Check → parse → resolve resources → global video-ID deduplication → EstimateVideos → Result`

The frontend and API only consume this result. No estimation logic belongs in either layer.


Here is the complete workstream list, kept to the actual implementation scope without splitting things unnecessarily:

1. **000 - Statistics Model Alignment**  
   Align existing `stats` models with preflight estimation. Add only required planned/estimated fields; do not create a parallel statistics hierarchy.

2. **001 - Metadata Cache Inspection**  
   Provide the cache-only metadata presence/validity check required by preflight, using existing cache keys and expiration semantics.

3. **002 - Video-Level Estimation**  
   Implement the core estimator accepting `[]videoID` and requested fields. Reuse the existing planner, inspect metadata cache, calculate metadata availability/missing fetches, and report potential subtitle fetches.

4. **003 - Resource-Level Estimation**  
   Handle individual resources. Direct videos pass one ID; playlists pass their complete enumerated video-ID array to the same video estimator and receive one aggregate estimate.

5. **004 - Global Preflight Aggregation**  
   Resolve all resources, globally deduplicate video IDs, call the video estimator for the complete unique set, and construct overall resource/video counts and aggregate statistics.

6. **005 - Preflight Service Integration**  
   Replace the existing coarse `Check` implementation with the new parse → resolve → deduplicate → estimate → result flow.

7. **006 - API and Frontend Migration**  
   Update API models/handlers and frontend consumers to display the new preflight result without implementing estimation logic outside the backend.

8. **007 - Final Verification and Cleanup**  
   Run unit/integration tests, verify deduplication and estimates, confirm no actual video processing occurs, and remove obsolete preflight structures.

   