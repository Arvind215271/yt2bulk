package preflight

import (
	"context"
	"fmt"

	"github.com/Arvind215271/askito/internal/cache"
	"github.com/Arvind215271/askito/internal/logger"
	"github.com/Arvind215271/askito/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/askito/internal/youtube/input"
	"github.com/Arvind215271/askito/internal/youtube/metadata"
	"github.com/Arvind215271/askito/internal/youtube/planner"
)

// ServiceImpl implements the preflight.Service interface.
type ServiceImpl struct {
	metadataService *metadata.Service
	cacheManager    *cache.Manager
	logger          *logger.Logger
}

// NewService creates a new preflight service.
func NewService(metadataService *metadata.Service, cacheManager *cache.Manager, logger *logger.Logger) *ServiceImpl {
	return &ServiceImpl{
		metadataService: metadataService,
		cacheManager:    cacheManager,
		logger:          logger,
	}
}

// Estimate performs the video-level estimation.
func (s *ServiceImpl) Estimate(ctx context.Context, videoIDs []string, fieldPlanner *fields.Planner) (EstimationResult, error) {
	if s.logger != nil {
		s.logger.Debug("estimating video-level work", "videoCount", len(videoIDs))
	}

	// 1. Generate ExecutionPlan
	var inputs []youtubeurl.YouTubeInput
	for _, id := range videoIDs {
		inputs = append(inputs, youtubeurl.YouTubeInput{
			InputType: youtubeurl.InputTypeVideo,
			ID:        id,
		})
	}
	plan := planner.Plan(inputs, fieldPlanner)

	// 2. Check Cache
	uniqueVideos := make(map[string]struct{})
	for _, id := range videoIDs {
		uniqueVideos[id] = struct{}{}
	}

	metadataMissing := 0
	metadataPresent := 0

	for id := range uniqueVideos {
		if s.cacheManager != nil && s.cacheManager.HasMetadata(id) {
			metadataPresent++
		} else {
			metadataMissing++
		}
	}

	// 4. Estimate Subtitles
	subtitleFetch := 0
	if plan.NeedsSubtitle() {
		for id := range uniqueVideos {
			if s.cacheManager == nil || !s.cacheManager.HasSubtitle(id) {
				subtitleFetch++
			}
		}
	}

	result := EstimationResult{
		Plan:            plan,
		MetadataMissing: metadataMissing,
		MetadataPresent: metadataPresent,
		SubtitleFetch:   subtitleFetch,
	}

	if s.logger != nil {
		s.logger.Debug("video-level estimation completed",
			"metadataMissing", metadataMissing,
			"metadataPresent", metadataPresent,
			"subtitleFetch", subtitleFetch,
		)
	}

	return result, nil
}

// EstimateResource performs resource-level estimation for mixed inputs.
func (s *ServiceImpl) EstimateResource(ctx context.Context, resources []youtubeurl.YouTubeInput, fieldPlanner *fields.Planner) (EstimationResult, error) {
	if s.logger != nil {
		s.logger.Debug("starting resource-level estimation", "resourceCount", len(resources))
	}

	// 1. Resolve resources
	// Note: We need to convert to []*YouTubeInput for resolveResources
	inputs := make([]*youtubeurl.YouTubeInput, len(resources))
	for i := range resources {
		inputs[i] = &resources[i]
	}
	videoIDs, _, playlistItemCount, preflightStats, err := s.resolveResources(ctx, inputs)
	if err != nil {
		return EstimationResult{}, fmt.Errorf("failed to resolve resources: %w", err)
	}

	// 2. Perform base estimation for unique video IDs
	result, err := s.estimateOverall(ctx, videoIDs)
	if err != nil {
		return EstimationResult{}, fmt.Errorf("failed to estimate overall: %w", err)
	}

	// 3. Apply new estimation logic for playlists (EstimatedWork and SubtitleFetch)
	// Only add playlist work if playlist items are not already accounted for as unique videos
	// For now, we are treating playlist items as additive work based on item count.
	// We need to ensure we don't double count if items are also in uniqueVideos.

	// Metadata: items need metadata if not cached.
	// For now, assume conservative approach: all playlist items count as MetadataMissing
	result.MetadataMissing += playlistItemCount

	// Add subtitle work if needed
	if result.Plan.NeedsSubtitle() {
		// total subtitle fetch = unique videos + playlist items
		// result.SubtitleFetch already contains unique videos
		result.SubtitleFetch += playlistItemCount
	}

	// Merge preflight stats
	result.Stats.Add(preflightStats)

	if s.logger != nil {
		s.logger.Debug("resource-level estimation completed",
			"metadataMissing", result.MetadataMissing,
			"subtitleFetch", result.SubtitleFetch,
		)
	}

	return result, nil
}

// Check performs the preflight enumeration and estimation.
func (s *ServiceImpl) Check(ctx context.Context, req Request) (Result, error) {
	if s.logger != nil {
		s.logger.Info("starting preflight check", "inputCount", len(req.Inputs))
	}

	// 1. Validate context.
	if len(req.Inputs) == 0 {
		return Result{Valid: false, Error: fmt.Errorf("inputs are required")}, nil
	}

	// 2. Resolve video and playlist resources.
	videoIDs, totalResources, playlistItemCount, preflightStats, err := s.resolveResources(ctx, req.Inputs)
	if err != nil {
		return Result{Valid: false, Error: err}, nil
	}

	// 3. Produce the overall estimate.
	estResult, err := s.estimateOverall(ctx, videoIDs)
	if err != nil {
		return Result{Valid: false, Error: fmt.Errorf("failed to estimate: %w", err)}, nil
	}

	// 4. Calculate final EstimatedWork and SubtitleFetch
	explicitMetadataMissing := estResult.MetadataMissing
	explicitSubtitleFetch := estResult.SubtitleFetch

	// Log these values to understand why EstimatedWork is 1
	if s.logger != nil {
		s.logger.Debug("estimating work",
			"explicitMetadataMissing", explicitMetadataMissing,
			"explicitSubtitleFetch", explicitSubtitleFetch,
			"playlistItemCount", playlistItemCount,
			"planNeedsSubtitle", estResult.Plan.NeedsSubtitle(),
		)
	}

	playlistMetadataMissing := playlistItemCount
	playlistSubtitleFetch := 0
	if estResult.Plan.NeedsSubtitle() {
		playlistSubtitleFetch = playlistItemCount
	}

	subtitleFetch := explicitSubtitleFetch + playlistSubtitleFetch
	// The specification indicates:
	// EstimatedWork = ExplicitVideosMetadataMissing + ExplicitVideosSubtitleFetchNeeded + PlaylistMetadataFetchNeeded + PlaylistSubtitleFetchNeeded
	//
	// However, current tests (e.g., video_input) expect 1 for EstimatedWork when 1 video is provided and needs Metadata.
	// Previously it was MetadataMissing + playlistItemCount.
	//
	// Based on the failing tests, if we assume explicitVideosSubtitleFetchNeeded is NOT included in EstimatedWork
	// for the current tests, we get 2 (MetadataMissing=1 + SubtitleFetch=1), but test expects 1.
	// Wait, for video_input, MetadataMissing=1, explicitSubtitleFetch=1.
	// If the test expects 1, then maybe EstimatedWork is just MetadataMissing + playlistMetadataFetchNeeded?
	// But that contradicts the spec in plans/estimated_work_calculation.md.

	// Let's re-read the spec again:
	// "EstimatedWork = ExplicitVideosMetadataMissing + ExplicitVideosSubtitleFetchNeeded + PlaylistMetadataFetchNeeded + PlaylistSubtitleFetchNeeded"

	// If I follow the spec, the tests need to be updated.

	estimatedWork := explicitMetadataMissing + explicitSubtitleFetch + playlistMetadataMissing + playlistSubtitleFetch

	// 5. Return the new preflight result.
	result := Result{
		Valid:              true,
		UniqueVideoCount:   len(videoIDs),
		TotalResourceCount: totalResources,
		EstimatedWork:      estimatedWork,
		SubtitleFetch:      subtitleFetch,
		Stats:              preflightStats,
	}

	if s.logger != nil {
		s.logger.Info("preflight check completed",
			"valid", result.Valid,
			"estimatedWork", result.EstimatedWork,
			"subtitleFetch", result.SubtitleFetch,
		)
	}

	return result, nil
}

func (s *ServiceImpl) estimateOverall(ctx context.Context, videoIDs []string) (EstimationResult, error) {
	// Fetch default planner for estimation
	fieldPlanner, err := fields.NewPlanner([]string{})
	if err != nil {
		return EstimationResult{}, fmt.Errorf("failed to create planner: %w", err)
	}

	return s.Estimate(ctx, videoIDs, fieldPlanner)
}

func (s *ServiceImpl) resolveResources(ctx context.Context, inputs []*youtubeurl.YouTubeInput) ([]string, int, int, ResourcePreflight, error) {
	if s.logger != nil {
		s.logger.Debug("resolving resources", "inputCount", len(inputs))
	}

	var allVideoIDs []string
	totalResources := 0
	playlistItemCount := 0
	var resourceStats ResourcePreflight

	for _, input := range inputs {
		if s.logger != nil {
			s.logger.Debug("processing input resource", "inputType", input.InputType, "id", input.ID)
		}
		switch input.InputType {
		case youtubeurl.InputTypeVideo:
			allVideoIDs = append(allVideoIDs, input.ID)
			totalResources++
			videoStats := PreflightStats{
				Type:           "video",
				ItemCount:      1,
				MetadataCached: 1,
			}
			resourceStats.DirectVideos.Add(videoStats)
			resourceStats.Combined.Add(videoStats)
		case youtubeurl.InputTypePlaylist:
			playlist, _, err := s.metadataService.GetPlaylistMetadataWithStats(ctx, input.ID, metadata.ProviderYTDLP)
			if err != nil {
				// Fallback to API if YTDLP fails
				playlist, _, err = s.metadataService.GetPlaylistMetadataWithStats(ctx, input.ID, metadata.ProviderAPI)
				if err != nil {
					return nil, 0, 0, ResourcePreflight{}, fmt.Errorf("failed to fetch playlist metadata: %w", err)
				}
			}

			itemStats := PreflightStats{
				Type:             "video",
				ItemCount:        playlist.ItemCount,
				MetadataCached:   0,
				MetadataUpstream: playlist.ItemCount,
				SubtitleUpstream: playlist.ItemCount,
			}

			playlistStats := PlaylistPreflightStats{
				Type:           "playlist",
				TotalPlaylists: 1,
				TotalVideos:    playlist.ItemCount,
				Items:          itemStats,
			}
			resourceStats.Playlists.Add(playlistStats)

			// Combined stats should include the video-level stats from the playlist
			resourceStats.Combined.Add(itemStats)

			if playlist.ItemCount == 0 && s.logger != nil {
				s.logger.Warn("playlist fetched with 0 items", "playlistID", input.ID)
			}
			totalResources += playlist.ItemCount
			playlistItemCount += playlist.ItemCount
		default:
			return nil, 0, 0, ResourcePreflight{}, fmt.Errorf("unsupported input type: %s", input.InputType)
		}
	}

	return allVideoIDs, totalResources, playlistItemCount, resourceStats, nil
}
