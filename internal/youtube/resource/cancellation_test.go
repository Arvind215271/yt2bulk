package resource_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Arvind215271/yt2bulk/internal/cache"
	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/description"
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
	"github.com/Arvind215271/yt2bulk/internal/youtube/pipeline"
	"github.com/Arvind215271/yt2bulk/internal/youtube/planner"
	"github.com/Arvind215271/yt2bulk/internal/youtube/resource"
	"github.com/Arvind215271/yt2bulk/internal/youtube/signal"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
	"github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
)

func TestCancellation_BeforeWorkStarts(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			t.Fatalf("getVideo should not be called when context is pre-cancelled")
			return youtube.Video{}, nil
		},
	}
	svc := setupResourceService(t, metaProv, &mockSubFetcher{}, 1)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle}, logger.New("test"))
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypeVideo, ID: "vid1"}}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{FieldPlanner: fp, ExecutionPlan: ep}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resources, processStats := svc.ProcessResourcesWithStats(ctx, inputs, req)
	if len(resources) != 1 || resources[0].Video != nil {
		t.Errorf("expected empty resource, got %+v", resources)
	}
	if processStats.ResourcesRequested != 0 || processStats.VideosProcessed != 0 {
		t.Errorf("expected zero stats, got %+v", processStats)
	}
}

func TestCancellation_WaitingForConcurrency(t *testing.T) {
	started1 := make(chan struct{})
	release1 := make(chan struct{})

	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if ctx.Err() != nil {
				return youtube.Video{}, context.Canceled
			}
			if videoID == "vid1" {
				close(started1)
				<-release1
				if st != nil {
					st.YTDLPRequests++
					st.Successes++
				}
				return youtube.Video{ID: videoID, Title: "Vid 1"}, nil
			}
			return youtube.Video{}, context.Canceled
		},
	}

	svc := setupResourceService(t, metaProv, &mockSubFetcher{}, 1)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle}, logger.New("test"))
	inputs := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypeVideo, ID: "vid1"},
		{InputType: youtubeurl.InputTypeVideo, ID: "vid2"},
	}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{FieldPlanner: fp, ExecutionPlan: ep}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	var processStats stats.ProcessStats

	go func() {
		_, processStats = svc.ProcessResourcesWithStats(ctx, inputs, req)
		close(done)
	}()

	select {
	case <-started1:
		cancel()
		close(release1)
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for vid1 to start")
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for processing to complete")
	}

	if processStats.ResourcesRequested > 2 {
		t.Errorf("expected at most 2 resource requested, got %d", processStats.ResourcesRequested)
	}
}

func TestCancellation_BetweenPipelineStages(t *testing.T) {
	subCalled := make(chan struct{})

	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
				st.Successes++
			}
			return youtube.Video{
				ID:    videoID,
				Title: "Vid",
				SubtitleMetadata: subtitle.SubtitleMetadata{
					Manual: []subtitle.SubtitleTrack{{LanguageCode: "en", Formats: []string{"json3"}}},
				},
			}, nil
		},
	}

	subFetcher := &mockSubFetcher{
		getSubtitleFunc: func(ctx context.Context, videoID, language, subType, format string) ([]byte, error) {
			close(subCalled)
			return nil, errors.New("should not be called")
		},
	}

	log := logger.New("test")
	metaSvc := metadata.NewService(nil, metaProv, log)
	descSvc := description.NewService(log)
	c := cache.NewManager(cache.Config{CacheDir: t.TempDir(), TTLDays: 1, MaxFiles: 100}, log)
	subSvc := subtitle.NewSubtitleService(c, log, subFetcher)
	transSvc := transcript.NewService(log)
	sigSvc := signal.NewSignalService()

	pipelineSvc := pipeline.NewService(metaSvc, descSvc, subSvc, transSvc, sigSvc, log, 1)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle, fields.FieldTranscriptText}, log)
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypeVideo, ID: "vid1"}}
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metaProv.getVideoFunc = func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
		cancel()
		if st != nil {
			st.YTDLPRequests++
			st.Successes++
		}
		return youtube.Video{
			ID:    videoID,
			Title: "Vid",
			SubtitleMetadata: subtitle.SubtitleMetadata{
				Manual: []subtitle.SubtitleTrack{{LanguageCode: "en", Formats: []string{"json3"}}},
			},
		}, nil
	}

	video, _ := resource.ProcessVideoWithStats(ctx, pipelineSvc, "vid1", req)
	select {
	case <-subCalled:
		t.Fatalf("subtitle fetcher should not have been called due to cancellation between stages")
	default:
	}
	if video.Title != "Vid" {
		t.Errorf("expected metadata to complete with title 'Vid', got %s", video.Title)
	}
}

func TestCancellation_DuringSubtitleExecution(t *testing.T) {
	subStarted := make(chan struct{})
	releaseSub := make(chan struct{})

	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
				st.Successes++
			}
			return youtube.Video{
				ID:    videoID,
				Title: "Vid",
				SubtitleMetadata: subtitle.SubtitleMetadata{
					Manual: []subtitle.SubtitleTrack{{LanguageCode: "en", Formats: []string{"json3"}}},
				},
			}, nil
		},
	}

	subFetcher := &mockSubFetcher{
		getSubtitleFunc: func(ctx context.Context, videoID, language, subType, format string) ([]byte, error) {
			close(subStarted)
			<-releaseSub
			return []byte(`{"events":[{"tStartMs":0,"dDurationMs":1000,"segs":[{"utf8":"Sub"}]}]}`), nil
		},
	}

	log := logger.New("test")
	metaSvc := metadata.NewService(nil, metaProv, log)
	descSvc := description.NewService(log)
	c := cache.NewManager(cache.Config{CacheDir: t.TempDir(), TTLDays: 1, MaxFiles: 100}, log)
	subSvc := subtitle.NewSubtitleService(c, log, subFetcher)
	transSvc := transcript.NewService(log)
	sigSvc := signal.NewSignalService()

	pipelineSvc := pipeline.NewService(metaSvc, descSvc, subSvc, transSvc, sigSvc, log, 1)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle, fields.FieldTranscriptText}, log)
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypeVideo, ID: "vid1"}}
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	var resStats stats.ResourceStats

	go func() {
		_, resStats = resource.ProcessVideoWithStats(ctx, pipelineSvc, "vid1", req)
		close(done)
	}()

	select {
	case <-subStarted:
		cancel()
		close(releaseSub)
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for subtitle to start")
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for processing to complete")
	}

	if resStats.Subtitle.UpstreamFetches != 1 {
		t.Errorf("expected 1 subtitle upstream fetch recorded, got %d", resStats.Subtitle.UpstreamFetches)
	}
}

func TestCancellation_GenuineFailureBeforeCancellation(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
				st.Failures++
			}
			return youtube.Video{}, errors.New("genuine metadata failure")
		},
	}
	svc := setupResourceService(t, metaProv, &mockSubFetcher{}, 1)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle}, logger.New("test"))
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypeVideo, ID: "fail_vid"}}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{FieldPlanner: fp, ExecutionPlan: ep}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resources, processStats := svc.ProcessResourcesWithStats(ctx, inputs, req)
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource")
	}
	if processStats.ResourceFailures != 1 {
		t.Errorf("expected 1 resource failure for genuine failure, got %d", processStats.ResourceFailures)
	}
	if len(resources[0].Video.Errors) == 0 {
		t.Errorf("expected video errors to contain genuine failure")
	}
}

func TestCancellation_SuccessfulOperationFollowedByCancellation(t *testing.T) {
	started1 := make(chan struct{})
	release1 := make(chan struct{})

	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if ctx.Err() != nil {
				return youtube.Video{}, context.Canceled
			}
			if videoID == "vid1" {
				close(started1)
				<-release1
				if st != nil {
					st.YTDLPRequests++
					st.Successes++
				}
				return youtube.Video{ID: videoID, Title: "Vid 1"}, nil
			}
			return youtube.Video{}, context.Canceled
		},
	}
	svc := setupResourceService(t, metaProv, &mockSubFetcher{}, 1)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle}, logger.New("test"))
	inputs := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypeVideo, ID: "vid1"},
		{InputType: youtubeurl.InputTypeVideo, ID: "vid2"},
	}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{FieldPlanner: fp, ExecutionPlan: ep}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	var resources []youtube.Resource
	var processStats stats.ProcessStats

	go func() {
		resources, processStats = svc.ProcessResourcesWithStats(ctx, inputs, req)
		close(done)
	}()

	select {
	case <-started1:
		close(release1)
		time.Sleep(10 * time.Millisecond)
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for vid1 to start")
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for processing to complete")
	}

	if len(resources) != 2 {
		t.Fatalf("expected 2 resources")
	}
	if resources[0].Video == nil || resources[0].Video.Title != "Vid 1" {
		t.Errorf("expected vid1 to succeed")
	}
	if processStats.ResourcesSucceeded != 1 {
		t.Errorf("expected 1 resource succeeded, got %d", processStats.ResourcesSucceeded)
	}
}

func TestCancellation_PlaylistCancellation(t *testing.T) {
	metaProv := &mockMetaProvider{
		getPlaylistItemsFunc: func(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
			if st != nil {
				st.YTDLPRequests++
				st.Successes++
			}
			return []youtube.PlaylistItem{
				{VideoID: "v1", Title: "V1", Position: 1},
				{VideoID: "v2", Title: "V2", Position: 2},
			}, nil
		},
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			return youtube.Video{ID: videoID, Title: "V1"}, nil
		},
	}
	svc := setupResourceService(t, metaProv, &mockSubFetcher{}, 1)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle}, logger.New("test"))
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypePlaylist, ID: "pl1"}}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{FieldPlanner: fp, ExecutionPlan: ep}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metaProv.getVideoFunc = func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
		cancel()
		if st != nil {
			st.YTDLPRequests++
			st.Successes++
		}
		return youtube.Video{ID: videoID, Title: "V1"}, nil
	}

	resources, processStats := svc.ProcessResourcesWithStats(ctx, inputs, req)
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource")
	}
	if processStats.ResourceFailures != 0 {
		t.Errorf("expected no resource failures on cancelled playlist processing, got %d", processStats.ResourceFailures)
	}
}

func TestCancellation_NormalExecutionRegression(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
				st.Successes++
			}
			return youtube.Video{ID: videoID, Title: "Normal " + videoID}, nil
		},
	}
	svc := setupResourceService(t, metaProv, &mockSubFetcher{}, 2)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle}, logger.New("test"))
	inputs := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypeVideo, ID: "v1"},
		{InputType: youtubeurl.InputTypeVideo, ID: "v2"},
	}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{FieldPlanner: fp, ExecutionPlan: ep}

	resources, processStats := svc.ProcessResourcesWithStats(context.Background(), inputs, req)
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}
	if processStats.ResourcesRequested != 2 || processStats.ResourcesSucceeded != 2 || processStats.VideosProcessed != 2 {
		t.Errorf("unexpected process stats: %+v", processStats)
	}
}
