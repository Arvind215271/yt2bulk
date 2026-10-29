package ytdlp

import (
	"context"
	"time"

	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

type Provider struct {
	client *Client
	logger *logger.Logger
}

func NewProvider(client *Client, logger *logger.Logger) *Provider {
	return &Provider{
		client: client,
		logger: logger,
	}
}

func (p *Provider) GetVideo(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
	p.logger.Debug("getting video from ytdlp provider", "videoID", videoID)
	meta, err := p.client.GetVideo(ctx, videoID, st)
	if err != nil {
		p.logger.Error("failed to get video from ytdlp provider", "error", err, "videoID", videoID)
		return youtube.Video{}, err
	}
	return MapVideo(meta), nil
}

func (p *Provider) GetPlaylistMetadata(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
	p.logger.Debug("getting playlist metadata from ytdlp provider", "playlistID", playlistID)
	meta, err := p.client.GetPlaylistMetadata(ctx, playlistID, st)
	if err != nil {
		p.logger.Error("failed to get playlist metadata from ytdlp provider", "error", err, "playlistID", playlistID)
		return youtube.Playlist{}, err
	}
	return MapPlaylist(meta), nil
}

func (p *Provider) GetChannel(ctx context.Context, channelURL string, st *stats.MetadataStats) (youtube.Playlist, error) {
	p.logger.Debug("getting channel from ytdlp provider", "channelURL", channelURL)
	meta, err := p.client.GetChannel(ctx, channelURL, st)
	if err != nil {
		p.logger.Error("failed to get channel from ytdlp provider", "error", err, "channelURL", channelURL)
		return youtube.Playlist{}, err
	}
	return MapPlaylist(meta), nil
}

func (p *Provider) GetPlaylistSurfaceMetadata(ctx context.Context, playlistID string, st *stats.MetadataStats) (PlaylistSurfaceMetadata, error) {
	p.logger.Debug("getting playlist surface metadata from ytdlp provider", "playlistID", playlistID)
	meta, err := p.client.GetPlaylistSurfaceMetadata(ctx, playlistID, st)
	if err != nil {
		p.logger.Error("failed to get playlist surface metadata from ytdlp provider", "error", err, "playlistID", playlistID)
		return PlaylistSurfaceMetadata{}, err
	}
	return meta, nil
}

func (p *Provider) GetPlaylistItems(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
	meta, err := p.client.GetPlaylist(ctx, playlistID, st)
	if err != nil {
		return nil, err
	}

	var flattenEntries func([]YTPlaylistEntry) []youtube.PlaylistItem
	flattenEntries = func(entries []YTPlaylistEntry) []youtube.PlaylistItem {
		var items []youtube.PlaylistItem
		for _, entry := range entries {
			if entry.Type == "playlist" {
				items = append(items, flattenEntries(entry.Entries)...)
				continue
			}

			var thumbnails []youtube.Thumbnail
			for _, th := range entry.Thumbnails {
				thumbnails = append(thumbnails, youtube.Thumbnail{
					URL:    th.URL,
					Width:  th.Width,
					Height: th.Height,
				})
			}
			if len(thumbnails) == 0 && entry.Thumbnail != "" {
				thumbnails = append(thumbnails, youtube.Thumbnail{
					URL: entry.Thumbnail,
				})
			}

			url := entry.WebpageURL
			if url == "" {
				url = entry.OriginalURL
			}
			if url == "" && entry.ID != "" {
				url = "https://www.youtube.com/watch?v=" + entry.ID
			}

			channelID := entry.ChannelID
			if channelID == "" {
				channelID = entry.ChannelIDAlt
			}

			items = append(items, youtube.PlaylistItem{
				VideoID:      entry.ID,
				AddedAt:      time.Time{},
				Title:        entry.Title,
				Description:  entry.Description,
				Duration:     time.Duration(entry.Duration) * time.Second,
				ViewCount:    entry.ViewCount,
				ChannelID:    channelID,
				ChannelTitle: entry.Channel,
				URL:          url,
				Thumbnails:   thumbnails,
			})
		}
		return items
	}

	result := flattenEntries(meta.Entries)
	for i := range result {
		result[i].Position = i
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (p *Provider) GetChannelItems(ctx context.Context, channelURL string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
	p.logger.Debug("getting channel items from ytdlp provider", "channelURL", channelURL)
	meta, err := p.client.GetChannel(ctx, channelURL, st)
	if err != nil {
		p.logger.Error("failed to get channel items from ytdlp provider", "error", err, "channelURL", channelURL)
		return nil, err
	}

	rawCount := len(meta.Entries)
	p.logger.Debug("channel extraction: total entries raw", "count", rawCount)
	for idx, entry := range meta.Entries {
		p.logger.Debug("channel extraction: entry sample", "index", idx, "id", entry.ID, "title", entry.Title, "type", entry.Type, "entries_count", len(entry.Entries))
	}

	// Define a map for allowed top-level tabs/containers to flatten
	allowedTabs := map[string]bool{
		"Videos": true,
		"Shorts": true,
	}

	var flattenEntries func([]YTPlaylistEntry, bool) []youtube.PlaylistItem
	flattenEntries = func(entries []YTPlaylistEntry, isTopLevel bool) []youtube.PlaylistItem {
		var items []youtube.PlaylistItem
		for _, entry := range entries {
			// If it's a top-level entry (like a tab/container), check if it's allowed
			if isTopLevel && entry.Type == "playlist" {
				if _, ok := allowedTabs[entry.Title]; !ok {
					p.logger.Debug("channel extraction: skipping top-level container", "title", entry.Title)
					continue
				}
				p.logger.Debug("channel extraction: flattening allowed top-level container", "title", entry.Title)
				items = append(items, flattenEntries(entry.Entries, false)...)
				continue
			}

			// If it's a nested playlist/container, flatten it
			if entry.Type == "playlist" || len(entry.Entries) > 0 {
				items = append(items, flattenEntries(entry.Entries, false)...)
				continue
			}

			// Otherwise, it's a terminal video node
			var thumbnails []youtube.Thumbnail
			for _, th := range entry.Thumbnails {
				thumbnails = append(thumbnails, youtube.Thumbnail{
					URL:    th.URL,
					Width:  th.Width,
					Height: th.Height,
				})
			}
			if len(thumbnails) == 0 && entry.Thumbnail != "" {
				thumbnails = append(thumbnails, youtube.Thumbnail{
					URL: entry.Thumbnail,
				})
			}

			url := entry.WebpageURL
			if url == "" {
				url = entry.OriginalURL
			}
			if url == "" && entry.ID != "" {
				url = "https://www.youtube.com/watch?v=" + entry.ID
			}

			channelID := entry.ChannelID
			if channelID == "" {
				channelID = entry.ChannelIDAlt
			}

			items = append(items, youtube.PlaylistItem{
				VideoID:      entry.ID,
				AddedAt:      time.Time{},
				Title:        entry.Title,
				Description:  entry.Description,
				Duration:     time.Duration(entry.Duration) * time.Second,
				ViewCount:    entry.ViewCount,
				ChannelID:    channelID,
				ChannelTitle: entry.Channel,
				URL:          url,
				Thumbnails:   thumbnails,
			})
		}
		return items
	}

	result := flattenEntries(meta.Entries, true)
	p.logger.Debug("channel extraction: total entries after flattening", "raw_count", rawCount, "flattened_count", len(result))

	for i := range result {
		result[i].Position = i
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}
