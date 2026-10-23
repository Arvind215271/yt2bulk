package pipeline_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Arvind215271/yt2bulk/internal/cache"
	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/description"
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
	"github.com/Arvind215271/yt2bulk/internal/youtube/pipeline"
	"github.com/Arvind215271/yt2bulk/internal/youtube/planner"
	"github.com/Arvind215271/yt2bulk/internal/youtube/signal"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
	"github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
)

type mockMetaProvider struct {
	getVideoFunc         func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error)
	getPlaylistItemsFunc func(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error)
}

func (m *mockMetaProvider) GetVideo(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
	if m.getVideoFunc != nil {
		return m.getVideoFunc(ctx, videoID, st)
	}
	return youtube.Video{}, nil
}
func (m *mockMetaProvider) GetPlaylistItems(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
	if m.getPlaylistItemsFunc != nil {
		return m.getPlaylistItemsFunc(ctx, playlistID, limit, st)
	}
	return nil, nil
}
func (m *mockMetaProvider) GetPlaylistMetadata(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
	return youtube.Playlist{}, nil
}

type mockSubFetcher struct {
	getSubtitleFunc func(ctx context.Context, videoID, language, subType, format string) ([]byte, error)
}

func (m *mockSubFetcher) GetSubtitle(ctx context.Context, videoID, language, subType, format string) ([]byte, error) {
	if m.getSubtitleFunc != nil {
		return m.getSubtitleFunc(ctx, videoID, language, subType, format)
	}
	return []byte(`{"events":[{"tStartMs":0,"dDurationMs":1000,"segs":[{"utf8":"Hello world"}]}]}`), nil
}

func setupPipelineService(t *testing.T, metaProvider metadata.Provider, subFetcher subtitle.SubtitleFetcher) *pipeline.Service {
	log := logger.New("test")
	metaSvc := metadata.NewService(nil, metaProvider)
	descSvc := description.NewService()
	cacheMgr := cache.NewManager(cache.Config{CacheDir: t.TempDir(), TTLDays: 1, MaxFiles: 100}, log)
	subSvc := subtitle.NewSubtitleService(cacheMgr, log, subFetcher)
	transSvc := transcript.NewService(log)
	sigSvc := signal.NewSignalService()

	return pipeline.NewService(metaSvc, descSvc, subSvc, transSvc, sigSvc, log, 1)
}

func TestPipeline_MetadataOnlySuccess(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
			}
			return youtube.Video{ID: videoID, Title: "Test Video"}, nil
		},
	}

	svc := setupPipelineService(t, metaProv, &mockSubFetcher{})

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle})
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypeVideo, ID: "vid1"}}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
	}

	video, pipelineStats, err := svc.ProcessWithStats(context.Background(), "vid1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if video.Title != "Test Video" {
		t.Errorf("expected title 'Test Video', got %s", video.Title)
	}
	if pipelineStats.Metadata.Successes != 1 || pipelineStats.Metadata.YTDLPRequests != 1 {
		t.Errorf("unexpected metadata stats: %+v", pipelineStats.Metadata)
	}
	if pipelineStats.Subtitle.Attempts != 0 {
		t.Errorf("expected zero subtitle attempts, got %d", pipelineStats.Subtitle.Attempts)
	}
}

func TestPipeline_MetadataAndSubtitleSuccess(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
			}
			return youtube.Video{
				ID:    videoID,
				Title: "Test Video",
				SubtitleMetadata: subtitle.SubtitleMetadata{
					Manual: []subtitle.SubtitleTrack{
						{LanguageCode: "en", Formats: []string{"json3"}},
					},
				},
			}, nil
		},
	}

	subFetch := &mockSubFetcher{
		getSubtitleFunc: func(ctx context.Context, videoID, language, subType, format string) ([]byte, error) {
			return []byte(`{"events":[{"tStartMs":0,"dDurationMs":1000,"segs":[{"utf8":"Hello"}]}]}`), nil
		},
	}

	svc := setupPipelineService(t, metaProv, subFetch)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle, fields.FieldTranscriptText})
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypeVideo, ID: "vid1"}}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
		Subtitle: &subtitle.DownloadRequest{
			VideoID:  "vid1",
			Language: "en",
			Type:     "manual",
			Format:   "json3",
		},
	}

	video, pipelineStats, err := svc.ProcessWithStats(context.Background(), "vid1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if video.TranscriptText == "" {
		t.Errorf("expected transcript text to be populated")
	}
	if pipelineStats.Metadata.Successes != 1 {
		t.Errorf("expected metadata success 1, got %d", pipelineStats.Metadata.Successes)
	}
	if pipelineStats.Subtitle.Successes != 1 || pipelineStats.Subtitle.ManualRequests != 1 {
		t.Errorf("unexpected subtitle stats: %+v", pipelineStats.Subtitle)
	}
}

func TestPipeline_MetadataFailure(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
			}
			return youtube.Video{}, errors.New("metadata fetch failed")
		},
	}

	svc := setupPipelineService(t, metaProv, &mockSubFetcher{})

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle})
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypeVideo, ID: "vid1"}}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
	}

	_, pipelineStats, err := svc.ProcessWithStats(context.Background(), "vid1", req)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if pipelineStats.Metadata.Failures != 1 || pipelineStats.Metadata.Attempts != 1 {
		t.Errorf("unexpected metadata stats on failure: %+v", pipelineStats.Metadata)
	}
}

func TestPipeline_SubtitleFailure(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
			}
			return youtube.Video{
				ID:    videoID,
				Title: "Test Video",
				SubtitleMetadata: subtitle.SubtitleMetadata{
					Manual: []subtitle.SubtitleTrack{
						{LanguageCode: "en", Formats: []string{"json3"}},
					},
				},
			}, nil
		},
	}

	subFetch := &mockSubFetcher{
		getSubtitleFunc: func(ctx context.Context, videoID, language, subType, format string) ([]byte, error) {
			return nil, errors.New("network error")
		},
	}

	svc := setupPipelineService(t, metaProv, subFetch)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle, fields.FieldTranscriptText})
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypeVideo, ID: "vid1"}}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
		Subtitle: &subtitle.DownloadRequest{
			VideoID:  "vid1",
			Language: "en",
			Type:     "manual",
			Format:   "json3",
		},
	}

	_, pipelineStats, err := svc.ProcessWithStats(context.Background(), "vid1", req)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if pipelineStats.Metadata.Successes != 1 {
		t.Errorf("expected metadata success 1, got %d", pipelineStats.Metadata.Successes)
	}
	if pipelineStats.Subtitle.Failures != 1 || pipelineStats.Subtitle.Attempts != 1 {
		t.Errorf("unexpected subtitle stats on failure: %+v", pipelineStats.Subtitle)
	}
}

func TestPipeline_SubtitleNotRequested(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
			}
			return youtube.Video{ID: videoID, Title: "Test Video"}, nil
		},
	}

	svc := setupPipelineService(t, metaProv, &mockSubFetcher{})

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle})
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypeVideo, ID: "vid1"}}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
	}

	video, pipelineStats, err := svc.ProcessWithStats(context.Background(), "vid1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if video.Title != "Test Video" {
		t.Errorf("expected title 'Test Video'")
	}
	if pipelineStats.Subtitle.Attempts != 0 || pipelineStats.Subtitle.Successes != 0 {
		t.Errorf("expected zero subtitle stats when not requested, got %+v", pipelineStats.Subtitle)
	}
}
