package preflight

import (
	"context"
	"fmt"

	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/planner"
)

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
	result.MetadataMissing += playlistItemCount

	// Add subtitle work if needed
	if result.Plan.NeedsSubtitle() {
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

func (s *ServiceImpl) estimateOverall(ctx context.Context, videoIDs []string) (EstimationResult, error) {
	// Fetch default planner for estimation
	fieldPlanner, err := fields.NewPlanner([]string{})
	if err != nil {
		return EstimationResult{}, fmt.Errorf("failed to create planner: %w", err)
	}

	return s.Estimate(ctx, videoIDs, fieldPlanner)
}
