
package preflight

import (
	"context"
	"fmt"

	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
)

func (s *ServiceImpl) resolveResources(ctx context.Context, inputs []*youtubeurl.YouTubeInput) ([]string, int, int, ResourcePreflight, error) {
	if s.logger != nil {
		s.logger.Debug("!!! RESOLVING RESOURCES START", "inputCount", len(inputs), "maxLimit", s.maxLimit)
	}

	var allVideoIDs []string
	totalResources := 0
	playlistItemCount := 0
	var resourceStats ResourcePreflight

	for _, input := range inputs {
		if s.logger != nil {
			s.logger.Debug("processing input resource", "inputType", input.InputType, "id", input.ID, "totalResources", totalResources)
		}
		switch input.InputType {
		case youtubeurl.InputTypeVideo:
			totalResources++
			if s.logger != nil {
				s.logger.Debug("... processing video", "id", input.ID, "totalResources", totalResources)
			}
			if totalResources > s.maxLimit {
				if s.logger != nil {
					s.logger.Warn("!!! EXPORT LIMIT EXCEEDED", "totalResources", totalResources, "maxLimit", s.maxLimit)
				}
				return allVideoIDs, totalResources, playlistItemCount, resourceStats, fmt.Errorf("EXPORT_LIMIT_EXCEEDED: reached %d, max allowed %d", totalResources, s.maxLimit)
			}
			allVideoIDs = append(allVideoIDs, input.ID)
			videoStats := PreflightStats{
				Type:             "video",
				ItemCount:        1,
				MetadataCached:   1,
				MetadataUpstream: 0,
			}
			resourceStats.DirectVideos.Add(videoStats)
			resourceStats.Combined.Add(videoStats)
		case youtubeurl.InputTypePlaylist:
			if s.logger != nil {
				s.logger.Debug("fetching playlist metadata", "playlistID", input.ID)
			}
			playlist, _, err := s.metadataService.GetPlaylistMetadataWithStats(ctx, input.ID, metadata.ProviderYTDLP)
			if err != nil {
				if s.logger != nil {
					s.logger.Warn("ytdlp playlist fetch failed, trying API", "playlistID", input.ID, "error", err)
				}
				// Fallback to API if YTDLP fails
				playlist, _, err = s.metadataService.GetPlaylistMetadataWithStats(ctx, input.ID, metadata.ProviderAPI)
				if err != nil {
					return allVideoIDs, totalResources, playlistItemCount, resourceStats, fmt.Errorf("failed to fetch playlist metadata: %w", err)
				}
			}
			if s.logger != nil {
				s.logger.Debug("fetched playlist metadata", "playlistID", input.ID, "itemCount", playlist.ItemCount)
			}

			if totalResources+playlist.ItemCount > s.maxLimit {
				if s.logger != nil {
					s.logger.Warn("!!! EXPORT LIMIT EXCEEDED (PLAYLIST)", "totalResources", totalResources, "itemCount", playlist.ItemCount, "maxLimit", s.maxLimit)
				}
				// FIX: Update totalResources to reflect the reached amount when returning error
				finalTotal := totalResources + playlist.ItemCount
				return allVideoIDs, finalTotal, playlistItemCount, resourceStats, fmt.Errorf("EXPORT_LIMIT_EXCEEDED: reached %d, max allowed %d", finalTotal, s.maxLimit)
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
			return allVideoIDs, totalResources, playlistItemCount, resourceStats, fmt.Errorf("unsupported input type: %s", input.InputType)
		}
	}

	return allVideoIDs, totalResources, playlistItemCount, resourceStats, nil
}
