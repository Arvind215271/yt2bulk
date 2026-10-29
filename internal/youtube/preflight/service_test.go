package preflight

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Arvind215271/yt2bulk/internal/cache"
	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

type mockProvider struct {
	getPlaylistMetadataFunc func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error)
	getChannelFunc          func(ctx context.Context, channelURL string, st *stats.MetadataStats) (youtube.Playlist, error)
	getChannelItemsFunc     func(ctx context.Context, channelURL string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error)
}

func (m *mockProvider) GetPlaylistMetadata(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
	if m.getPlaylistMetadataFunc != nil {
		return m.getPlaylistMetadataFunc(ctx, playlistID, st)
	}
	return youtube.Playlist{}, nil
}

func (m *mockProvider) GetChannel(ctx context.Context, channelURL string, st *stats.MetadataStats) (youtube.Playlist, error) {
	if m.getChannelFunc != nil {
		return m.getChannelFunc(ctx, channelURL, st)
	}
	return youtube.Playlist{}, nil
}

func (m *mockProvider) GetChannelItems(ctx context.Context, channelURL string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
	if m.getChannelItemsFunc != nil {
		return m.getChannelItemsFunc(ctx, channelURL, limit, st)
	}
	return nil, nil
}

func (m *mockProvider) GetPlaylistItems(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
	return nil, nil
}

func (m *mockProvider) GetVideo(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
	return youtube.Video{}, nil
}

func TestService_Check(t *testing.T) {
	t.Run("video input", func(t *testing.T) {
		svc := NewService(nil, nil, logger.New("test"), 100)
		req := Request{
			Inputs: []*youtubeurl.YouTubeInput{
				{
					InputType: youtubeurl.InputTypeVideo,
					ID:        "vid1",
				},
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
		// Since default planner includes subtitles, vid1 needs metadata (1) + subtitle fetch (1) = 2
		if res.EstimatedWork != 2 {
			t.Errorf("expected 2 estimated work, got %d", res.EstimatedWork)
		}
	})

	t.Run("playlist input metadata item count", func(t *testing.T) {
		mockProv := &mockProvider{
			getPlaylistMetadataFunc: func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
				return youtube.Playlist{
					ID:        playlistID,
					ItemCount: 3,
				}, nil
			},
		}

		metaSvc := metadata.NewService(mockProv, nil, logger.New("test"))
		svc := NewService(metaSvc, nil, logger.New("test"), 100)

		req := Request{
			Inputs: []*youtubeurl.YouTubeInput{
				{
					InputType: youtubeurl.InputTypePlaylist,
					ID:        "pl1",
				},
			},
		}

		res, err := svc.Check(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !res.Valid {
			t.Errorf("expected valid result")
		}
		// Playlist metadata count estimation does not enumerate items to get unique video IDs directly
		if res.TotalResourceCount != 3 {
			t.Errorf("expected 3 total resources, got %d", res.TotalResourceCount)
		}
		// Playlist metadata missing (3) + playlist subtitle fetch (3) = 6
		if res.EstimatedWork != 6 {
			t.Errorf("expected 6 estimated work, got %d", res.EstimatedWork)
		}
		if res.Stats.Playlists.Items.MetadataCached != 0 {
			t.Errorf("expected 0 metadata cached, got %d", res.Stats.Playlists.Items.MetadataCached)
		}
		if res.Stats.Playlists.Items.MetadataUpstream != 3 {
			t.Errorf("expected 3 metadata upstream, got %d", res.Stats.Playlists.Items.MetadataUpstream)
		}
	})

	t.Run("Estimate video-level metadata cache hits and misses", func(t *testing.T) {
		// Setup temp cache
		cacheDir, err := os.MkdirTemp("", "cache-test-estimate")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(cacheDir)

		cfg := cache.Config{
			CacheDir: cacheDir,
			TTLDays:  1,
			MaxFiles: 10,
		}
		manager := cache.NewManager(cfg, nil)

		svc := NewService(nil, manager, logger.New("test"), 100)

		videoIDs := []string{"vid1", "vid2"}
		// vid1 in cache, vid2 not in cache
		manager.Save("vid1", manager.VideoKey(), []byte("{}"))

		planner, _ := fields.NewPlanner([]string{}, logger.New("test"))

		res, err := svc.Estimate(context.Background(), videoIDs, planner)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if res.MetadataPresent != 1 {
			t.Errorf("expected 1 metadata present, got %d", res.MetadataPresent)
		}
		if res.MetadataMissing != 1 {
			t.Errorf("expected 1 metadata missing, got %d", res.MetadataMissing)
		}
	})

	t.Run("Estimate subtitle fetch scenarios", func(t *testing.T) {
		svc := NewService(nil, nil, logger.New("test"), 100)

		// Planner with transcript_text (which enforces subtitle stage)
		pWithSub, _ := fields.NewPlanner([]string{"transcript_text"}, logger.New("test"))
		// Planner without subtitles/transcripts (e.g. only id/title)
		pNoSub, _ := fields.NewPlanner([]string{"id", "title"}, logger.New("test"))

		videoIDs := []string{"vid1", "vid2", "vid3"}

		// Scenario: Subtitle stage included in plan
		resWithSub, err := svc.Estimate(context.Background(), videoIDs, pWithSub)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resWithSub.SubtitleFetch != 3 {
			t.Errorf("expected 3 subtitle fetches, got %d", resWithSub.SubtitleFetch)
		}

		// Scenario: Subtitle stage NOT included in plan
		resNoSub, err := svc.Estimate(context.Background(), videoIDs, pNoSub)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resNoSub.SubtitleFetch != 0 {
			t.Errorf("expected 0 subtitle fetches, got %d", resNoSub.SubtitleFetch)
		}
	})

	t.Run("EstimateResource with playlist", func(t *testing.T) {
		mockProv := &mockProvider{
			getPlaylistMetadataFunc: func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
				return youtube.Playlist{
					ID:        playlistID,
					ItemCount: 2,
				}, nil
			},
		}

		metaSvc := metadata.NewService(mockProv, nil, logger.New("test"))
		svc := NewService(metaSvc, nil, logger.New("test"), 100)

		resources := []youtubeurl.YouTubeInput{
			{InputType: youtubeurl.InputTypePlaylist, ID: "pl1"},
		}

		planner, _ := fields.NewPlanner([]string{"transcript_text"}, logger.New("test"))

		res, err := svc.EstimateResource(context.Background(), resources, planner)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Playlist metadata estimation uses item count (2 items)
		// If subtitles are enabled in planner, it should be 2, otherwise 0.
		if res.SubtitleFetch != 2 {
			t.Errorf("expected 2 subtitle fetch, got %d", res.SubtitleFetch)
		}
	})

	t.Run("mixed input stats aggregation", func(t *testing.T) {
		mockProv := &mockProvider{
			getPlaylistMetadataFunc: func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
				return youtube.Playlist{
					ID:        playlistID,
					ItemCount: 2,
				}, nil
			},
		}

		metaSvc := metadata.NewService(mockProv, nil, logger.New("test"))
		svc := NewService(metaSvc, nil, logger.New("test"), 100)

		req := Request{
			Inputs: []*youtubeurl.YouTubeInput{
				{InputType: "video", ID: "vid3"},
				{InputType: "playlist", ID: "pl1"},
			},
		}

		res, err := svc.Check(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if res.UniqueVideoCount != 1 {
			t.Errorf("expected 1 unique video from explicit video inputs, got %d", res.UniqueVideoCount)
		}
		if res.TotalResourceCount != 3 { // 1 video + 2 playlist items from ItemCount
			t.Errorf("expected 3 total resources, got %d", res.TotalResourceCount)
		}
		// 1 video metadata + 1 video subtitle + 2 playlist metadata + 2 playlist subtitles = 6
		if res.EstimatedWork != 6 {
			t.Errorf("expected 6 estimated work, got %d", res.EstimatedWork)
		}
		if res.Stats.Playlists.Items.MetadataCached != 0 {
			t.Errorf("expected 0 metadata cached, got %d", res.Stats.Playlists.Items.MetadataCached)
		}
		if res.Stats.Playlists.Items.MetadataUpstream != 2 {
			t.Errorf("expected 2 metadata upstream, got %d", res.Stats.Playlists.Items.MetadataUpstream)
		}
	})

	t.Run("large playlist metadata count estimation", func(t *testing.T) {
		mockProv := &mockProvider{
			getPlaylistMetadataFunc: func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
				return youtube.Playlist{
					ID:        playlistID,
					ItemCount: 1500,
				}, nil
			},
		}

		metaSvc := metadata.NewService(mockProv, nil, logger.New("test"))
		svc := NewService(metaSvc, nil, logger.New("test"), 100)

		req := Request{
			Inputs: []*youtubeurl.YouTubeInput{
				{InputType: youtubeurl.InputTypePlaylist, ID: "large_pl"},
			},
		}

		// The service was initialized with maxLimit=100.
		// "large_pl" is expected to have 1500 items, which exceeds the limit.
		_, err := svc.Check(context.Background(), req)
		if err == nil {
			t.Fatalf("expected error from Check, got nil")
		}
		if !strings.Contains(err.Error(), "EXPORT_LIMIT_EXCEEDED") {
			t.Errorf("expected error containing EXPORT_LIMIT_EXCEEDED, got %v", err)
		}
	})

	t.Run("Reproduction: EstimatedWork calculation", func(t *testing.T) {
		svc := NewService(nil, nil, logger.New("test"), 100)
		req := Request{
			Inputs: []*youtubeurl.YouTubeInput{
				{
					InputType: youtubeurl.InputTypeVideo,
					ID:        "vid1",
				},
			},
		}

		res, err := svc.Check(context.Background(), req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		// Based on service.go: EstimatedWork = MetadataMissing(1) + SubtitleFetch(1) = 2
		if res.EstimatedWork != 2 {
			t.Errorf("DEBUG: Expected 2 estimated work, got %d", res.EstimatedWork)
		}
	})
}
