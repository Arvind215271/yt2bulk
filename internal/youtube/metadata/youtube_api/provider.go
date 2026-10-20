// internal/youtube/providers/youtube_api/provider.go

package youtubeapi

import (
	"context"
	"fmt"

	"github.com/Arvind215271/yt2bulk/internal/logger"
	youtube "github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

type Provider struct {
	client *Client
	logger *logger.Logger
}

func NewProvider(
	client *Client,
	logger *logger.Logger,
) *Provider {
	return &Provider{
		client: client,
		logger: logger,
	}
}

func (p *Provider) GetPlaylistMetadata(
	ctx context.Context,
	playlistID string,
	st *stats.MetadataStats,
) (youtube.Playlist, error) {
	if st != nil {
		st.APIRequests++
		st.CacheMisses++
		st.UpstreamFetches++
	}

	// contain playlist metadata only not any videoID in it.
	if p.client == nil {
		return youtube.Playlist{}, youtube.Err.Playlist.FetchFailed().Wrap(fmt.Errorf("youtube client not initialized"))
	}
	playlist, err := p.client.GetPlaylist(
		ctx,
		playlistID,
	)
	if err != nil {
		return youtube.Playlist{}, err
	}

	thumbnailURL := p.getPlaylistThumbnail(playlist)
	var thumbnails []youtube.Thumbnail
	if thumbnailURL != "" {
		thumbnails = append(thumbnails, youtube.Thumbnail{
			URL: thumbnailURL,
		})
	}

	return youtube.Playlist{
		ID:           playlist.Id,
		Title:        playlist.Snippet.Title,
		Description:  playlist.Snippet.Description,
		ChannelID:    playlist.Snippet.ChannelId,
		ChannelTitle: playlist.Snippet.ChannelTitle,

		Thumbnails: thumbnails,

		ItemCount: int(playlist.ContentDetails.ItemCount),

		PrivacyStatus: playlist.Status.PrivacyStatus,

		PublishedAt: p.parseTime(
			playlist.Snippet.PublishedAt,
		),
	}, nil
}

// for a single video metadata fetching
func (p *Provider) GetVideo(
	ctx context.Context,
	videoID string,
	st *stats.MetadataStats,
) (youtube.Video, error) {
	if st != nil {
		st.APIRequests++
		st.CacheMisses++
		st.UpstreamFetches++
	}

	if p.client == nil {
		return youtube.Video{}, youtube.Err.Video.FetchFailed().Wrap(fmt.Errorf("youtube client not initialized"))
	}
	videoList, err := p.client.GetVideos(
		ctx,
		[]string{videoID},
	)
	if err != nil {
		return youtube.Video{}, err
	}

	if len(videoList) == 0 {
		return youtube.Video{}, youtube.Err.Video.NotFound()
	}

	video := videoList[0]

	return p.mapVideo(video), nil
}

func (p *Provider) GetPlaylistItems(
	ctx context.Context,
	playlistID string,
	limit int,
	st *stats.MetadataStats,
) ([]youtube.PlaylistItem, error) {
	if st != nil {
		st.APIRequests++
		st.CacheMisses++
		st.UpstreamFetches++
	}
	if p.client == nil {
		return nil, youtube.Err.Playlist.FetchFailed().Wrap(fmt.Errorf("youtube client not initialized"))
	}
	items, err := p.client.GetPlaylistItems(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	result := make([]youtube.PlaylistItem, 0, len(items))
	for _, item := range items {
		if item == nil || item.Snippet == nil || item.ContentDetails == nil {
			continue
		}

		var thumbnails []youtube.Thumbnail
		if item.Snippet.Thumbnails != nil {
			if item.Snippet.Thumbnails.Default != nil {
				thumbnails = append(thumbnails, youtube.Thumbnail{
					URL:    item.Snippet.Thumbnails.Default.Url,
					Width:  int(item.Snippet.Thumbnails.Default.Width),
					Height: int(item.Snippet.Thumbnails.Default.Height),
				})
			}
			if item.Snippet.Thumbnails.Medium != nil {
				thumbnails = append(thumbnails, youtube.Thumbnail{
					URL:    item.Snippet.Thumbnails.Medium.Url,
					Width:  int(item.Snippet.Thumbnails.Medium.Width),
					Height: int(item.Snippet.Thumbnails.Medium.Height),
				})
			}
			if item.Snippet.Thumbnails.High != nil {
				thumbnails = append(thumbnails, youtube.Thumbnail{
					URL:    item.Snippet.Thumbnails.High.Url,
					Width:  int(item.Snippet.Thumbnails.High.Width),
					Height: int(item.Snippet.Thumbnails.High.Height),
				})
			}
		}

		var privacyStatus string
		if item.Status != nil {
			privacyStatus = item.Status.PrivacyStatus
		}

		result = append(result, youtube.PlaylistItem{
			VideoID:       item.ContentDetails.VideoId,
			Position:      int(item.Snippet.Position),
			AddedAt:       p.parseTime(item.Snippet.PublishedAt),
			Title:         item.Snippet.Title,
			Description:   item.Snippet.Description,
			ChannelID:     item.Snippet.ChannelId,
			ChannelTitle:  item.Snippet.ChannelTitle,
			Thumbnails:    thumbnails,
			PrivacyStatus: privacyStatus,
		})
	}
	return result, nil
}
