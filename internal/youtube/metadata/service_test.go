package metadata_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Arvind215271/yt2bulk/internal/cache"
	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

type mockProvider struct {
	getVideoFunc         func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error)
	getPlaylistItemsFunc func(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error)
	getPlaylistMetaFunc  func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error)
}

func (m *mockProvider) GetVideo(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
	if m.getVideoFunc != nil {
		return m.getVideoFunc(ctx, videoID, st)
	}
	return youtube.Video{}, nil
}

func (m *mockProvider) GetPlaylistItems(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
	if m.getPlaylistItemsFunc != nil {
		return m.getPlaylistItemsFunc(ctx, playlistID, limit, st)
	}
	return nil, nil
}

func (m *mockProvider) GetPlaylistMetadata(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
	if m.getPlaylistMetaFunc != nil {
		return m.getPlaylistMetaFunc(ctx, playlistID, st)
	}
	return youtube.Playlist{}, nil
}

func TestMetadataService_GetVideoWithStats_Success(t *testing.T) {
	ytdlp := &mockProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
				st.CacheHits++
			}
			return youtube.Video{ID: videoID, Title: "Test Video"}, nil
		},
	}

	svc := metadata.NewService(nil, ytdlp)
	video, st, err := svc.GetVideoWithStats(context.Background(), "vid1", metadata.ProviderYTDLP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if video.ID != "vid1" {
		t.Errorf("expected video ID vid1, got %s", video.ID)
	}
	if st.Attempts != 1 || st.Successes != 1 || st.CacheHits != 1 || st.YTDLPRequests != 1 {
		t.Errorf("unexpected stats: %+v", st)
	}
}

func TestMetadataService_GetVideoWithStats_Fallback(t *testing.T) {
	ytdlp := &mockProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
				st.CacheMisses++
				st.UpstreamFetches++
			}
			return youtube.Video{}, errors.New("ytdlp failed")
		},
	}
	api := &mockProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.APIRequests++
				st.CacheMisses++
				st.UpstreamFetches++
			}
			return youtube.Video{ID: videoID, Title: "API Video"}, nil
		},
	}

	svc := metadata.NewService(api, ytdlp)
	video, st, err := svc.GetVideoWithStats(context.Background(), "vid1", metadata.ProviderYTDLP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if video.Title != "API Video" {
		t.Errorf("expected API Video title, got %s", video.Title)
	}
	if st.Attempts != 1 || st.Successes != 1 || st.Fallbacks != 1 || st.YTDLPRequests != 1 || st.APIRequests != 1 {
		t.Errorf("unexpected stats: %+v", st)
	}
}

func TestMetadataService_GetVideoWithStats_Failure(t *testing.T) {
	ytdlp := &mockProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
			}
			return youtube.Video{}, errors.New("total failure")
		},
	}

	svc := metadata.NewService(nil, ytdlp)
	_, st, err := svc.GetVideoWithStats(context.Background(), "vid1", metadata.ProviderYTDLP)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if st.Attempts != 1 || st.Failures != 1 || st.Successes != 0 {
		t.Errorf("unexpected stats: %+v", st)
	}
}

func TestMetadataService_GetPlaylistMetadataWithStats_Success(t *testing.T) {
	ytdlp := &mockProvider{
		getPlaylistMetaFunc: func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
			if st != nil {
				st.YTDLPRequests++
			}
			return youtube.Playlist{ID: playlistID, Title: "Surface Metadata Only"}, nil
		},
	}

	svc := metadata.NewService(nil, ytdlp)
	playlist, st, err := svc.GetPlaylistMetadataWithStats(context.Background(), "pl1", metadata.ProviderYTDLP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if playlist.ID != "pl1" || playlist.Title != "Surface Metadata Only" {
		t.Errorf("unexpected playlist: %+v", playlist)
	}
	if st.Attempts != 1 || st.Successes != 1 || st.YTDLPRequests != 1 {
		t.Errorf("unexpected stats: %+v", st)
	}
}

func TestMetadataService_GetPlaylistMetadataWithStats_Fallback(t *testing.T) {
	ytdlp := &mockProvider{
		getPlaylistMetaFunc: func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
			if st != nil {
				st.YTDLPRequests++
			}
			return youtube.Playlist{}, errors.New("ytdlp failed")
		},
	}
	api := &mockProvider{
		getPlaylistMetaFunc: func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
			if st != nil {
				st.APIRequests++
			}
			return youtube.Playlist{ID: playlistID, Title: "API Playlist"}, nil
		},
	}

	svc := metadata.NewService(api, ytdlp)
	playlist, st, err := svc.GetPlaylistMetadataWithStats(context.Background(), "pl1", metadata.ProviderYTDLP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if playlist.Title != "API Playlist" {
		t.Errorf("expected API Playlist title, got %s", playlist.Title)
	}
	if st.Attempts != 1 || st.Successes != 1 || st.Fallbacks != 1 || st.YTDLPRequests != 1 || st.APIRequests != 1 {
		t.Errorf("unexpected stats: %+v", st)
	}
}

func TestMetadataService_GetPlaylistMetadata_Cache(t *testing.T) {
	cacheDir := t.TempDir()
	cacheManager := cache.NewManager(cache.Config{
		CacheDir: cacheDir,
		TTLDays:  1,
	}, nil)

	playlistID := "pl1"
	playlist := youtube.Playlist{ID: playlistID, Title: "Test Playlist"}
	callCount := 0

	ytdlp := &mockProvider{
		getPlaylistMetaFunc: func(ctx context.Context, id string, st *stats.MetadataStats) (youtube.Playlist, error) {
			callCount++
			key := cacheManager.PlaylistMetadataKey()
			if cacheManager.Exists(id, key) {
				st.CacheHits++
				data, _ := cacheManager.Get(id, key)
				var p youtube.Playlist
				json.Unmarshal(data, &p)
				return p, nil
			}
			st.CacheMisses++
			st.UpstreamFetches++
			data, _ := json.Marshal(playlist)
			cacheManager.Save(id, key, data)
			return playlist, nil
		},
	}

	svc := metadata.NewService(nil, ytdlp)

	// First call: Miss
	_, st1, err := svc.GetPlaylistMetadataWithStats(context.Background(), playlistID, metadata.ProviderYTDLP)
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call to provider, got %d", callCount)
	}
	if st1.CacheMisses != 1 || st1.CacheHits != 0 {
		t.Errorf("unexpected stats on first call: %+v", st1)
	}

	// Second call: Hit
	var st2 stats.MetadataStats
	_, st2, err = svc.GetPlaylistMetadataWithStats(context.Background(), playlistID, metadata.ProviderYTDLP)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to provider, got %d", callCount)
	}
	if st2.CacheHits != 1 || st2.CacheMisses != 0 {
		t.Errorf("unexpected stats on second call: %+v", st2)
	}
}

func TestMetadataService_GetPlaylistMetadataWithStats_ItemCount(t *testing.T) {
	ytdlp := &mockProvider{
		getPlaylistMetaFunc: func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
			if st != nil {
				st.YTDLPRequests++
			}
			return youtube.Playlist{ID: playlistID, ItemCount: 42}, nil
		},
		getPlaylistItemsFunc: func(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
			t.Errorf("GetPlaylistItems should not have been called")
			return nil, nil
		},
	}

	svc := metadata.NewService(nil, ytdlp)
	playlist, _, err := svc.GetPlaylistMetadataWithStats(context.Background(), "pl1", metadata.ProviderYTDLP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if playlist.ID != "pl1" || playlist.ItemCount != 42 {
		t.Errorf("unexpected playlist: %+v", playlist)
	}
}
