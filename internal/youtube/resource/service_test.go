package resource_test

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
	"github.com/Arvind215271/yt2bulk/internal/youtube/resource"
	"github.com/Arvind215271/yt2bulk/internal/youtube/signal"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
	"github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
)

type mockMetaProvider struct {
	getVideoFunc         func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error)
	getPlaylistItemsFunc func(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error)
	getPlaylistMetaFunc  func(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error)
}

func (m *mockMetaProvider) GetVideo(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
	if m.getVideoFunc != nil {
		return m.getVideoFunc(ctx, videoID, st)
	}
	return youtube.Video{ID: videoID, Title: "Default Video"}, nil
}

func (m *mockMetaProvider) GetPlaylistItems(ctx context.Context, playlistID string, limit int, st *stats.MetadataStats) ([]youtube.PlaylistItem, error) {
	if m.getPlaylistItemsFunc != nil {
		return m.getPlaylistItemsFunc(ctx, playlistID, limit, st)
	}
	return []youtube.PlaylistItem{
		{VideoID: "vid1", Title: "Playlist Video 1", Position: 1},
		{VideoID: "vid2", Title: "Playlist Video 2", Position: 2},
	}, nil
}

func (m *mockMetaProvider) GetPlaylistMetadata(ctx context.Context, playlistID string, st *stats.MetadataStats) (youtube.Playlist, error) {
	if m.getPlaylistMetaFunc != nil {
		return m.getPlaylistMetaFunc(ctx, playlistID, st)
	}
	return youtube.Playlist{ID: playlistID, Title: "Test Playlist"}, nil
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

func setupResourceService(t *testing.T, metaProvider metadata.Provider, subFetcher subtitle.SubtitleFetcher, concurrency int) *resource.Service {
	log := logger.New("test")
	metaSvc := metadata.NewService(nil, metaProvider)
	descSvc := description.NewService()
	_ = cache.NewManager(cache.Config{CacheDir: t.TempDir(), TTLDays: 1, MaxFiles: 100}, log)
	subSvc := subtitle.NewSubtitleService(log, subFetcher)
	transSvc := transcript.NewService(log)
	sigSvc := signal.NewSignalService()

	pipelineSvc := pipeline.NewService(metaSvc, descSvc, subSvc, transSvc, sigSvc, log, concurrency)
	return resource.NewService(metaSvc, pipelineSvc, log, concurrency)
}

func TestProcessVideoWithStats(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
				st.Successes++
			}
			return youtube.Video{ID: videoID, Title: "Video " + videoID}, nil
		},
	}
	svc := setupResourceService(t, metaProv, &mockSubFetcher{}, 1)
	_ = svc

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle})
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypeVideo, ID: "vid1"}}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
	}

	metaSvc := metadata.NewService(nil, metaProv)
	pipelineSvc := pipeline.NewService(metaSvc, description.NewService(logger.New("test")), subtitle.NewSubtitleService(logger.New("test"), &mockSubFetcher{}), transcript.NewService(logger.New("test")), signal.NewSignalService(), logger.New("test"), 1)

	video, resStats := resource.ProcessVideoWithStats(context.Background(), pipelineSvc, "vid1", req)
	if video.Title != "Video vid1" {
		t.Errorf("expected Title 'Video vid1', got %s", video.Title)
	}
	if resStats.VideosProcessed != 1 || resStats.VideoFailures != 0 {
		t.Errorf("unexpected resource stats: %+v", resStats)
	}
	if resStats.Metadata.YTDLPRequests != 1 {
		t.Errorf("expected 1 YTDLP request, got %d", resStats.Metadata.YTDLPRequests)
	}
}

func TestProcessPlaylistWithStats(t *testing.T) {
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
			if st != nil {
				st.YTDLPRequests++
				st.Successes++
			}
			return youtube.Video{ID: videoID, Title: "Title " + videoID}, nil
		},
	}
	svc := setupResourceService(t, metaProv, &mockSubFetcher{}, 2)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle})
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypePlaylist, ID: "pl1"}}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
	}

	resources, processStats := svc.ProcessResourcesWithStats(context.Background(), inputs, req)
	if len(resources) != 1 || resources[0].Type != youtube.ResourceTypePlaylist {
		t.Fatalf("expected 1 playlist resource")
	}
	pl := resources[0].Playlist
	if pl.ItemCount != 2 || len(pl.Videos) != 2 {
		t.Errorf("expected 2 playlist videos, got %d", len(pl.Videos))
	}
	if processStats.ResourcesRequested != 1 || processStats.ResourcesSucceeded != 1 || processStats.VideosProcessed != 2 {
		t.Errorf("unexpected process stats: %+v", processStats)
	}
}

func TestProcessResourcesWithStats_MultiAndConcurrent(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if st != nil {
				st.YTDLPRequests++
				st.Successes++
			}
			return youtube.Video{ID: videoID, Title: "Video " + videoID}, nil
		},
	}
	svc := setupResourceService(t, metaProv, &mockSubFetcher{}, 4)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle})
	inputs := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypeVideo, ID: "vid1"},
		{InputType: youtubeurl.InputTypeVideo, ID: "vid2"},
		{InputType: youtubeurl.InputTypeVideo, ID: "vid3"},
	}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
	}

	resources, processStats := svc.ProcessResourcesWithStats(context.Background(), inputs, req)
	if len(resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(resources))
	}
	if processStats.ResourcesRequested != 3 || processStats.ResourcesSucceeded != 3 || processStats.VideosProcessed != 3 {
		t.Errorf("unexpected process stats: %+v", processStats)
	}
	if processStats.Metadata.YTDLPRequests != 3 {
		t.Errorf("expected 3 YTDLP requests, got %d", processStats.Metadata.YTDLPRequests)
	}
}

func TestProcessResourcesWithStats_PartialFailure(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			if videoID == "fail_vid" {
				if st != nil {
					st.YTDLPRequests++
					st.Failures++
				}
				return youtube.Video{}, errors.New("metadata failed")
			}
			if st != nil {
				st.YTDLPRequests++
				st.Successes++
			}
			return youtube.Video{ID: videoID, Title: "Success Vid"}, nil
		},
	}
	svc := setupResourceService(t, metaProv, &mockSubFetcher{}, 2)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle})
	inputs := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypeVideo, ID: "good_vid"},
		{InputType: youtubeurl.InputTypeVideo, ID: "fail_vid"},
	}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
	}

	resources, processStats := svc.ProcessResourcesWithStats(context.Background(), inputs, req)
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources")
	}
	if processStats.ResourcesRequested != 2 {
		t.Errorf("expected 2 requested, got %d", processStats.ResourcesRequested)
	}
	if processStats.VideosProcessed != 2 {
		t.Errorf("expected 2 videos processed, got %d", processStats.VideosProcessed)
	}
	if processStats.ResourceFailures != 1 || processStats.ResourcesSucceeded != 1 {
		t.Errorf("expected 1 success and 1 failure, got succeeded=%d, failures=%d", processStats.ResourcesSucceeded, processStats.ResourceFailures)
	}
}

func TestProcessResourcesWithStats_Cancellation(t *testing.T) {
	metaProv := &mockMetaProvider{
		getVideoFunc: func(ctx context.Context, videoID string, st *stats.MetadataStats) (youtube.Video, error) {
			return youtube.Video{ID: videoID, Title: "Should Not Run"}, nil
		},
	}
	svc := setupResourceService(t, metaProv, &mockSubFetcher{}, 1)

	fp, _ := fields.NewPlanner([]string{fields.FieldTitle})
	inputs := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypeVideo, ID: "vid1"},
		{InputType: youtubeurl.InputTypeVideo, ID: "vid2"},
	}
	ep := planner.Build(inputs, fp)
	req := &pipeline.Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	resources, processStats := svc.ProcessResourcesWithStats(ctx, inputs, req)
	if len(resources) != 2 {
		t.Fatalf("expected 2 resource slots, got %d", len(resources))
	}
	for i, res := range resources {
		if res.ID != "" || res.Video != nil {
			t.Errorf("expected empty resource at index %d, got %+v", i, res)
		}
	}
	if processStats.ResourcesRequested != 0 || processStats.ResourceFailures != 0 || processStats.VideosProcessed != 0 {
		t.Errorf("expected zero stats on immediate cancellation, got %+v", processStats)
	}
}
