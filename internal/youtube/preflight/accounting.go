package preflight

import (
	"context"

	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
	"github.com/Arvind215271/yt2bulk/internal/youtube/planner"
)

// addVideoPreflight performs the fixed video accounting based on metadata and subtitle cache state.
func (s *ServiceImpl) addVideoPreflight(ctx context.Context, stats *PreflightStats, videoID string, plan *planner.ExecutionPlan) {
	stats.ItemCount++

	// 1. Check metadata cache
	if s.cacheManager != nil && s.cacheManager.HasMetadata(videoID) {
		stats.MetadataCached++
	} else {
		stats.MetadataUpstream++
		stats.SubtitleUpstream++
		return
	}

	// 2. If subtitle processing is not requested, return.
	if !plan.NeedsSubtitle() {
		return
	}

	// 3. Check subtitle cache
	if s.cacheManager != nil && s.cacheManager.HasSubtitle(videoID) {
		stats.SubtitleCached++
	} else {
		stats.SubtitleUpstream++
	}
}

// addPlaylistPreflight performs the fixed playlist accounting based on metadata state.
func (s *ServiceImpl) addPlaylistPreflight(ctx context.Context, stats *PlaylistPreflightStats, playlistID string, plan *planner.ExecutionPlan) {
	stats.Type = "playlist"
	stats.TotalPlaylists++

	// 3. Check playlist metadata cache
	videoIDs, exists := s.cacheManager.GetPlaylistVideoIDs(playlistID)

	if exists {
		stats.TotalVideos += len(videoIDs)
		// 6. Call addVideoPreflight for every video ID into stats.Items
		for _, id := range videoIDs {
			s.addVideoPreflight(ctx, &stats.Items, id, plan)
		}
	} else {
		// If playlist metadata does not exist, fetch metadata count
		playlist, _, err := s.metadataService.GetPlaylistMetadataWithStats(ctx, playlistID, metadata.ProviderYTDLP)
		if err == nil {
			stats.TotalVideos += playlist.ItemCount
			stats.Items.MetadataUpstream += playlist.ItemCount
			if plan.NeedsSubtitle() {
				stats.Items.SubtitleUpstream += playlist.ItemCount
			}
		}
	}
}
