package metadata

import (
	"context"

	"github.com/Arvind215271/askito/internal/youtube"
	"github.com/Arvind215271/askito/internal/youtube/stats"
)

type Provider interface {
	GetPlaylistMetadata(
		ctx context.Context,
		playlistID string,
		st *stats.MetadataStats,
	) (youtube.Playlist, error)

	GetPlaylistItems(
		ctx context.Context,
		playlistID string,
		limit int,
		st *stats.MetadataStats,
	) ([]youtube.PlaylistItem, error)

	GetVideo(
		ctx context.Context,
		videoID string,
		st *stats.MetadataStats,
	) (youtube.Video, error)
}
