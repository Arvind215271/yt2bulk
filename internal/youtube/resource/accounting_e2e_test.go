package resource_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Arvind215271/askito/internal/cache"
	"github.com/Arvind215271/askito/internal/logger"
	"github.com/Arvind215271/askito/internal/youtube"
	"github.com/Arvind215271/askito/internal/youtube/description"
	"github.com/Arvind215271/askito/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/askito/internal/youtube/input"
	"github.com/Arvind215271/askito/internal/youtube/metadata"
	"github.com/Arvind215271/askito/internal/youtube/pipeline"
	"github.com/Arvind215271/askito/internal/youtube/planner"
	"github.com/Arvind215271/askito/internal/youtube/resource"
	"github.com/Arvind215271/askito/internal/youtube/signal"
	"github.com/Arvind215271/askito/internal/youtube/stats"
	"github.com/Arvind215271/askito/internal/youtube/subtitle"
	"github.com/Arvind215271/askito/internal/youtube/transcript"
)

type e2eMockMetaProvider struct {
	getVideoFunc         func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error)
	getPlaylistItemsFunc func(ctx context.Context, playlistID string, st *stats.MetadataStats) ([]youtube.PlaylistItem, error)
	getPlaylistMetaFunc  func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error)
}

func (m *e2eMockMetaProvider) GetVideo(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
	if m.getVideoFunc != nil {
		return m.getVideoFunc(ctx, videoID, st)
	}
	return youtube.Video{ID: videoID, Title: "Video " + videoID}, nil
}

func (m *e2eMockMetaProvider) GetPlaylistItems(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
	if m.getPlaylistItemsFunc != nil {
		return m.getPlaylistItemsFunc(ctx, playlistID, st)
	}
	return nil, nil
}

func (m *e2eMockMetaProvider) GetPlaylistMetadata(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
	if m.getPlaylistMetaFunc != nil {
		return m.getPlaylistMetaFunc(ctx, playlistID, st)
	}
	return youtube.Playlist{ID: playlistID, Title: "Playlist " + playlistID}, nil
}

type e2eMockSubFetcher struct {
	fetchFunc func(ctx context.Context, videoID, language, subType, format string) ([]byte, error)
}

func (m *e2eMockSubFetcher) GetSubtitle(ctx context.Context, videoID, language, subType, format string) ([]byte, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx, videoID, language, subType, format)
	}
	return []byte(`{"events":[{"tStartMs":0,"dDurationMs":1000,"segs":[{"utf8":"E2E Subtitle"}]}]}`), nil
}

func TestAccountingE2E_MultiResourceComplexTree(t *testing.T) {
	log := logger.New("test")
	tmpDir := t.TempDir()
	cacheMgr := cache.NewManager(cache.Config{CacheDir: tmpDir, TTLDays: 1, MaxFiles: 100}, log)

	ytdlpProv := &e2eMockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
				st.Successes++
			}
			return youtube.Video{
				ID:    videoID,
				Title: "Title " + videoID,
				SubtitleMetadata: subtitle.SubtitleMetadata{
					Manual: []subtitle.SubtitleTrack{{LanguageCode: "en", Formats: []string{"json3"}}},
				},
			}, nil
		},
		getPlaylistItemsFunc: func(ctx context.Context, playlistID string, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
			if st != nil {
				st.YTDLPRequests++
				st.Successes++
			}
			return []youtube.PlaylistItem{
				{VideoID: "pl_vid1", Title: "PL Vid 1", Position: 1},
				{VideoID: "pl_vid2", Title: "PL Vid 2", Position: 2},
			}, nil
		},
	}

	apiProv := &e2eMockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.APIRequests++
				st.Successes++
			}
			return youtube.Video{ID: videoID, Title: "API Title " + videoID}, nil
		},
	}

	metaSvc := metadata.NewService(apiProv, ytdlpProv)
	descSvc := description.NewService()
	subFetcher := &e2eMockSubFetcher{}
	subSvc := subtitle.NewSubtitleService(cacheMgr, log, subFetcher)
	transSvc := transcript.NewService()
	sigSvc := signal.NewSignalService()

	pipelineSvc := pipeline.NewService(metaSvc, descSvc, subSvc, transSvc, sigSvc, log, 4)
	resourceSvc := resource.NewService(metaSvc, pipelineSvc, log, 4)

	fp, err := fields.NewPlanner([]string{fields.FieldTitle, fields.FieldTranscriptText})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inputs := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypeVideo, ID: "vid_direct1"},
		{InputType: youtubeurl.InputTypePlaylist, ID: "playlist1"},
	}

	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
		Subtitle: &subtitle.DownloadRequest{
			Type:     "manual",
			Language: "en",
			Format:   "json3",
		},
	}

	// First run: Cache miss for subtitles
	_, processStats := resourceSvc.ProcessResourcesWithStats(context.Background(), inputs, req)

	// Verify counts:
	// Resources requested: 2 (1 video, 1 playlist)
	// Resources succeeded: 2
	// Videos processed: 3 (1 direct video + 2 items in playlist1)
	if processStats.ResourcesRequested != 2 {
		t.Errorf("expected 2 resources requested, got %d", processStats.ResourcesRequested)
	}
	if processStats.ResourcesSucceeded != 2 {
		t.Errorf("expected 2 resources succeeded, got %d", processStats.ResourcesSucceeded)
	}
	if processStats.VideosProcessed != 3 {
		t.Errorf("expected 3 videos processed, got %d", processStats.VideosProcessed)
	}
	if processStats.Metadata.Attempts != 4 { // 1 direct video + 1 playlist items fetch + 2 playlist videos = 4 metadata attempts
		t.Errorf("expected 4 metadata attempts, got %d", processStats.Metadata.Attempts)
	}
	if processStats.Subtitle.Attempts != 3 { // 3 videos requested subtitle download
		t.Errorf("expected 3 subtitle attempts, got %d", processStats.Subtitle.Attempts)
	}
	if processStats.Subtitle.CacheMisses != 3 {
		t.Errorf("expected 3 subtitle cache misses, got %d", processStats.Subtitle.CacheMisses)
	}
	if processStats.Subtitle.UpstreamFetches != 3 {
		t.Errorf("expected 3 subtitle upstream fetches, got %d", processStats.Subtitle.UpstreamFetches)
	}

	// Second run: With cache hits for subtitles
	_, processStats2 := resourceSvc.ProcessResourcesWithStats(context.Background(), inputs, req)
	if processStats2.Subtitle.CacheHits != 3 {
		t.Errorf("expected 3 subtitle cache hits on second run, got %d", processStats2.Subtitle.CacheHits)
	}
}

func TestAccountingE2E_FallbacksAndPartialFailures(t *testing.T) {
	log := logger.New("test")
	tmpDir := t.TempDir()
	cacheMgr := cache.NewManager(cache.Config{CacheDir: tmpDir, TTLDays: 1, MaxFiles: 100}, log)

	ytdlpProv := &e2eMockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if videoID == "fallback_vid" || videoID == "fail_vid" {
				if st != nil {
					st.YTDLPRequests++
				}
				return youtube.Video{}, errors.New("ytdlp failed")
			}
			if st != nil {
				st.YTDLPRequests++
				st.Successes++
			}
			return youtube.Video{ID: videoID, Title: "OK Vid"}, nil
		},
	}

	apiProv := &e2eMockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if videoID == "fail_vid" {
				if st != nil {
					st.APIRequests++
				}
				return youtube.Video{}, errors.New("api failed too")
			}
			if st != nil {
				st.APIRequests++
				st.Successes++
			}
			return youtube.Video{ID: videoID, Title: "API Fallback Title"}, nil
		},
	}

	metaSvc := metadata.NewService(apiProv, ytdlpProv)
	descSvc := description.NewService()
	subFetcher := &e2eMockSubFetcher{}
	subSvc := subtitle.NewSubtitleService(cacheMgr, log, subFetcher)
	transSvc := transcript.NewService()
	sigSvc := signal.NewSignalService()

	pipelineSvc := pipeline.NewService(metaSvc, descSvc, subSvc, transSvc, sigSvc, log, 2)
	resourceSvc := resource.NewService(metaSvc, pipelineSvc, log, 2)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle})
	inputs := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypeVideo, ID: "fallback_vid"},
		{InputType: youtubeurl.InputTypeVideo, ID: "fail_vid"},
	}

	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
	}

	_, processStats := resourceSvc.ProcessResourcesWithStats(context.Background(), inputs, req)

	if processStats.ResourcesRequested != 2 {
		t.Errorf("expected 2 requested, got %d", processStats.ResourcesRequested)
	}
	// fallback_vid succeeds via API fallback, fail_vid fails completely.
	// So 1 resource succeeds, 1 resource fails.
	if processStats.ResourcesSucceeded != 1 || processStats.ResourceFailures != 1 {
		t.Errorf("expected 1 success and 1 failure, got succeeded=%d, failures=%d", processStats.ResourcesSucceeded, processStats.ResourceFailures)
	}
	if processStats.Metadata.Fallbacks != 2 {
		t.Errorf("expected 2 metadata fallbacks, got %d", processStats.Metadata.Fallbacks)
	}
	if processStats.Metadata.APIRequests != 2 {
		t.Errorf("expected 2 API requests (for fallbacks), got %d", processStats.Metadata.APIRequests)
	}
	if processStats.Metadata.Failures != 1 { // only fail_vid fails both ytdlp and api
		t.Errorf("expected 1 metadata failure, got %d", processStats.Metadata.Failures)
	}
}
