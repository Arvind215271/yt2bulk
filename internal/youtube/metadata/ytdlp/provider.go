package ytdlp

import (
	"context"
	"time"

	"github.com/Arvind215271/askito/internal/logger"
	"github.com/Arvind215271/askito/internal/youtube"
	"github.com/Arvind215271/askito/internal/youtube/stats"
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

	result := make([]youtube.PlaylistItem, 0, len(meta.Entries))
	if limit > 0 && len(meta.Entries) > limit {
		meta.Entries = meta.Entries[:limit]
	}
	for i, entry := range meta.Entries {
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

		result = append(result, youtube.PlaylistItem{
			VideoID:      entry.ID,
			Position:     i,
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
	return result, nil
}
