package preflight

import (
	"context"
	"testing"

	"github.com/Arvind215271/askito/internal/youtube"
	youtubeurl "github.com/Arvind215271/askito/internal/youtube/input"
	"github.com/Arvind215271/askito/internal/youtube/metadata"
	"github.com/Arvind215271/askito/internal/youtube/stats"
)

type mockProvider struct {
	getPlaylistItemsFunc func(ctx context.Context, playlistID string, st *stats.MetadataStats) ([]youtube.PlaylistItem, error)
}

func (m *mockProvider) GetPlaylistMetadata(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
	return youtube.Playlist{}, nil
}

func (m *mockProvider) GetPlaylistItems(ctx context.Context, playlistID string, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
	return m.getPlaylistItemsFunc(ctx, playlistID, st)
}

func (m *mockProvider) GetVideo(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
	return youtube.Video{}, nil
}

func TestService_Check(t *testing.T) {
	t.Run("video input", func(t *testing.T) {
		svc := NewService(nil)
		req := Request{
			Input: &youtubeurl.YouTubeInput{
				InputType: youtubeurl.InputTypeVideo,
				ID:        "vid1",
			},
		}

		res, err := svc.Check(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !res.Valid {
			t.Errorf("expected valid result")
		}
		if res.UniqueVideoCount != 1 {
			t.Errorf("expected 1 unique video, got %d", res.UniqueVideoCount)
		}
		if res.TotalResourceCount != 1 {
			t.Errorf("expected 1 total resource, got %d", res.TotalResourceCount)
		}
		if res.EstimatedWork != 1 {
			t.Errorf("expected 1 estimated work, got %d", res.EstimatedWork)
		}
	})

	t.Run("playlist input deduplication", func(t *testing.T) {
		mockProv := &mockProvider{
			getPlaylistItemsFunc: func(ctx context.Context, playlistID string, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
				return []youtube.PlaylistItem{
					{VideoID: "vid1"},
					{VideoID: "vid2"},
					{VideoID: "vid1"}, // Duplicate
				}, nil
			},
		}

		metaSvc := metadata.NewService(mockProv, nil)
		svc := NewService(metaSvc)

		req := Request{
			Input: &youtubeurl.YouTubeInput{
				InputType: youtubeurl.InputTypePlaylist,
				ID:        "pl1",
			},
		}

		res, err := svc.Check(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !res.Valid {
			t.Errorf("expected valid result")
		}
		if res.UniqueVideoCount != 2 {
			t.Errorf("expected 2 unique videos, got %d", res.UniqueVideoCount)
		}
		if res.TotalResourceCount != 3 {
			t.Errorf("expected 3 total resources, got %d", res.TotalResourceCount)
		}
		if res.EstimatedWork != 2 {
			t.Errorf("expected 2 estimated work, got %d", res.EstimatedWork)
		}
	})
}
