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
		case youtubeurl.InputTypePlaylist, youtubeurl.InputTypeChannel:
			if s.logger != nil {
				s.logger.Debug("fetching collection metadata", "type", input.InputType, "id", input.ID)
			}

			// Helper to fetch collection metadata (playlist or channel)
			fetchFunc := func(provider metadata.ProviderType) (int, error) {
				var count int
				if input.InputType == youtubeurl.InputTypePlaylist {
					playlist, _, err := s.metadataService.GetPlaylistMetadataWithStats(ctx, input.ID, provider)
					if err != nil {
						return 0, err
					}
					count = playlist.ItemCount
				} else {
					channel, _, err := s.metadataService.GetChannelWithStats(ctx, input.ID, provider)
					if err != nil {
						return 0, err
					}
					count = channel.ItemCount
				}
				return count, nil
			}

			// Default to YTDLP only as requested.
			itemCount, err := fetchFunc(metadata.ProviderYTDLP)
			if err != nil {
				if s.logger != nil {
					s.logger.Error("ytdlp collection fetch failed", "type", input.InputType, "id", input.ID, "error", err)
				}
				return allVideoIDs, totalResources, playlistItemCount, resourceStats, fmt.Errorf("failed to fetch collection metadata via ytdlp: %w", err)
			}

			if s.logger != nil {
				s.logger.Debug("fetched collection metadata", "type", input.InputType, "id", input.ID, "itemCount", itemCount)
			}

			if totalResources+itemCount > s.maxLimit {
				if s.logger != nil {
					s.logger.Warn("!!! EXPORT LIMIT EXCEEDED (COLLECTION)", "totalResources", totalResources, "itemCount", itemCount, "maxLimit", s.maxLimit)
				}
				finalTotal := totalResources + itemCount
				return allVideoIDs, finalTotal, playlistItemCount, resourceStats, fmt.Errorf("EXPORT_LIMIT_EXCEEDED: reached %d, max allowed %d", finalTotal, s.maxLimit)
			}

			itemStats := PreflightStats{
				Type:             "video",
				ItemCount:        itemCount,
				MetadataCached:   0,
				MetadataUpstream: itemCount,
				SubtitleUpstream: itemCount,
			}

			if input.InputType == youtubeurl.InputTypePlaylist {
				playlistStats := PlaylistPreflightStats{
					Type:           "playlist",
					TotalPlaylists: 1,
					TotalVideos:    itemCount,
					Items:          itemStats,
				}
				resourceStats.Playlists.Add(playlistStats)
			} else {
				channelStats := ChannelPreflightStats{
					Type:          "channel",
					TotalChannels: 1,
					TotalVideos:   itemCount,
					Items:         itemStats,
				}
				resourceStats.Channels.Add(channelStats)
			}

			// Combined stats should include the video-level stats from the collection
			resourceStats.Combined.Add(itemStats)

			if itemCount == 0 && s.logger != nil {
				s.logger.Warn("collection fetched with 0 items", "type", input.InputType, "id", input.ID)
			}
			totalResources += itemCount
			playlistItemCount += itemCount
		default:
			return allVideoIDs, totalResources, playlistItemCount, resourceStats, fmt.Errorf("unsupported input type: %s", input.InputType)
		}
	}

	return allVideoIDs, totalResources, playlistItemCount, resourceStats, nil
}
