package metadata

import (
	"context"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

type Service struct {
	apiProvider   Provider
	ytdlpProvider Provider
}

func NewService(
	apiProvider Provider,
	ytdlpProvider Provider,
) *Service {
	return &Service{
		apiProvider:   apiProvider,
		ytdlpProvider: ytdlpProvider,
	}
}

func (s *Service) GetPlaylistMetadataWithStats(
	ctx context.Context,
	playlistID string,
	providerType ProviderType,
) (youtube.Playlist, stats.MetadataStats, error) {
	var st stats.MetadataStats
	st.Attempts++

	var playlist youtube.Playlist
	var err error

	useAPI := (providerType == ProviderAPI)
	if useAPI && s.apiProvider != nil {
		playlist, err = s.apiProvider.GetPlaylistMetadata(ctx, playlistID, &st)
	} else if s.ytdlpProvider != nil {
		playlist, err = s.ytdlpProvider.GetPlaylistMetadata(ctx, playlistID, &st)
		if err != nil && s.apiProvider != nil {
			st.Fallbacks++
			playlist, err = s.apiProvider.GetPlaylistMetadata(ctx, playlistID, &st)
		}
	} else if s.apiProvider != nil {
		playlist, err = s.apiProvider.GetPlaylistMetadata(ctx, playlistID, &st)
	} else {
		err = youtube.Err.Playlist.FetchFailed()
	}

	if err != nil {
		st.Failures++
		return youtube.Playlist{}, st, err
	}

	st.Successes++
	return playlist, st, nil
}

func (s *Service) GetPlaylistMetadata(
	ctx context.Context,
	playlistID string,
	providerType ProviderType,
) (youtube.Playlist, error) {
	playlist, _, err := s.GetPlaylistMetadataWithStats(ctx, playlistID, providerType)
	return playlist, err
}

func (s *Service) GetPlaylistItemsWithStats(
	ctx context.Context,
	playlistID string,
	limit int,
	providerType ProviderType,
) ([]youtube.PlaylistItem, stats.MetadataStats, error) {
	var st stats.MetadataStats
	st.Attempts++

	var items []youtube.PlaylistItem
	var err error

	useAPI := (providerType == ProviderAPI)
	if useAPI && s.apiProvider != nil {
		items, err = s.apiProvider.GetPlaylistItems(ctx, playlistID, limit, &st)
	} else if s.ytdlpProvider != nil {
		items, err = s.ytdlpProvider.GetPlaylistItems(ctx, playlistID, limit, &st)
		if err != nil && s.apiProvider != nil {
			st.Fallbacks++
			items, err = s.apiProvider.GetPlaylistItems(ctx, playlistID, limit, &st)
		}
	} else if s.apiProvider != nil {
		items, err = s.apiProvider.GetPlaylistItems(ctx, playlistID, limit, &st)
	} else {
		err = youtube.Err.Playlist.FetchFailed()
	}

	if err != nil {
		st.Failures++
		return nil, st, err
	}

	st.Successes++
	return items, st, nil
}

func (s *Service) GetPlaylistItems(
	ctx context.Context,
	playlistID string,
	limit int,
	providerType ProviderType,
) ([]youtube.PlaylistItem, error) {
	items, _, err := s.GetPlaylistItemsWithStats(ctx, playlistID, limit, providerType)
	return items, err
}

func (s *Service) GetVideoWithStats(
	ctx context.Context,
	videoID string,
	providerType ProviderType,
) (youtube.Video, stats.MetadataStats, error) {
	var st stats.MetadataStats
	st.Attempts++

	var video youtube.Video
	var err error

	useAPI := (providerType == ProviderAPI)
	if useAPI && s.apiProvider != nil {
		video, err = s.apiProvider.GetVideo(ctx, videoID, &st)
	} else if s.ytdlpProvider != nil {
		video, err = s.ytdlpProvider.GetVideo(ctx, videoID, &st)
		if err != nil && s.apiProvider != nil {
			st.Fallbacks++
			video, err = s.apiProvider.GetVideo(ctx, videoID, &st)
		}
	} else if s.apiProvider != nil {
		video, err = s.apiProvider.GetVideo(ctx, videoID, &st)
	} else {
		err = youtube.Err.Video.FetchFailed()
	}

	if err != nil {
		st.Failures++
		return youtube.Video{}, st, err
	}

	st.Successes++
	return video, st, nil
}

func (s *Service) GetVideo(
	ctx context.Context,
	videoID string,
	providerType ProviderType,
) (youtube.Video, error) {
	video, _, err := s.GetVideoWithStats(ctx, videoID, providerType)
	return video, err
}
