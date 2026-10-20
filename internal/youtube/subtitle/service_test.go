package subtitle_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Arvind215271/yt2bulk/internal/cache"
	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
)

type mockFetcher struct {
	getSubtitleFunc func(ctx context.Context, videoID, language, subType, format string) ([]byte, error)
}

func (m *mockFetcher) GetSubtitle(ctx context.Context, videoID, language, subType, format string) ([]byte, error) {
	if m.getSubtitleFunc != nil {
		return m.getSubtitleFunc(ctx, videoID, language, subType, format)
	}
	return nil, nil
}

func TestSubtitleService_DownloadSubtitleWithStats_CacheHit(t *testing.T) {
	tmpDir := t.TempDir()
	log := logger.New("test")
	cacheMgr := cache.NewManager(cache.Config{CacheDir: tmpDir, TTLDays: 1, MaxFiles: 100}, log)

	// Pre-populate cache
	cacheKey := "subtitle.en.json3"
	if err := cacheMgr.Save("vid1", cacheKey, []byte("cached subtitle data")); err != nil {
		t.Fatalf("failed to seed cache: %v", err)
	}

	meta := subtitle.SubtitleMetadata{
		Manual: []subtitle.SubtitleTrack{
			{LanguageCode: "en", Formats: []string{"json3"}},
		},
	}

	svc := subtitle.NewSubtitleService(cacheMgr, log, &mockFetcher{})

	req := subtitle.DownloadRequest{
		VideoID:  "vid1",
		Language: "en",
		Type:     "manual",
		Format:   "json3",
	}

	res, st, err := svc.DownloadSubtitleWithStats(context.Background(), req, meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(res.Content) != "cached subtitle data" {
		t.Errorf("expected content 'cached subtitle data', got %s", string(res.Content))
	}
	if st.Attempts != 1 || st.Successes != 1 || st.CacheHits != 1 || st.CacheMisses != 0 || st.UpstreamFetches != 0 || st.ManualRequests != 1 {
		t.Errorf("unexpected stats: %+v", st)
	}
}

func TestSubtitleService_DownloadSubtitleWithStats_CacheMiss(t *testing.T) {
	tmpDir := t.TempDir()
	log := logger.New("test")
	cacheMgr := cache.NewManager(cache.Config{CacheDir: tmpDir, TTLDays: 1, MaxFiles: 100}, log)

	fetcher := &mockFetcher{
		getSubtitleFunc: func(ctx context.Context, videoID, language, subType, format string) ([]byte, error) {
			return []byte("fetched subtitle data"), nil
		},
	}

	meta := subtitle.SubtitleMetadata{
		Automatic: []subtitle.SubtitleTrack{
			{LanguageCode: "es", Formats: []string{"json3"}},
		},
	}

	svc := subtitle.NewSubtitleService(cacheMgr, log, fetcher)

	req := subtitle.DownloadRequest{
		VideoID:  "vid2",
		Language: "es",
		Type:     "automatic",
		Format:   "json3",
	}

	res, st, err := svc.DownloadSubtitleWithStats(context.Background(), req, meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(res.Content) != "fetched subtitle data" {
		t.Errorf("expected content 'fetched subtitle data', got %s", string(res.Content))
	}
	if st.Attempts != 1 || st.Successes != 1 || st.CacheHits != 0 || st.CacheMisses != 1 || st.UpstreamFetches != 1 || st.AutomaticRequests != 1 {
		t.Errorf("unexpected stats: %+v", st)
	}
}

func TestSubtitleService_DownloadSubtitleWithStats_ValidationFailure(t *testing.T) {
	tmpDir := t.TempDir()
	log := logger.New("test")
	cacheMgr := cache.NewManager(cache.Config{CacheDir: tmpDir, TTLDays: 1, MaxFiles: 100}, log)

	svc := subtitle.NewSubtitleService(cacheMgr, log, &mockFetcher{})

	req := subtitle.DownloadRequest{
		VideoID:  "", // invalid
		Language: "en",
		Type:     "manual",
	}

	_, st, err := svc.DownloadSubtitleWithStats(context.Background(), req, subtitle.SubtitleMetadata{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if st.Attempts != 1 || st.Failures != 1 || st.Successes != 0 {
		t.Errorf("unexpected stats: %+v", st)
	}
}

func TestSubtitleService_DownloadSubtitleWithStats_TrackValidationFailure(t *testing.T) {
	tmpDir := t.TempDir()
	log := logger.New("test")
	cacheMgr := cache.NewManager(cache.Config{CacheDir: tmpDir, TTLDays: 1, MaxFiles: 100}, log)

	svc := subtitle.NewSubtitleService(cacheMgr, log, &mockFetcher{})

	req := subtitle.DownloadRequest{
		VideoID:  "vid3",
		Language: "de", // not in metadata
		Type:     "manual",
		Format:   "json3",
	}

	_, st, err := svc.DownloadSubtitleWithStats(context.Background(), req, subtitle.SubtitleMetadata{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if st.Attempts != 1 || st.Failures != 1 || st.ManualRequests != 1 {
		t.Errorf("unexpected stats: %+v", st)
	}
}

func TestSubtitleService_DownloadSubtitleWithStats_DownloadFailure(t *testing.T) {
	tmpDir := t.TempDir()
	log := logger.New("test")
	cacheMgr := cache.NewManager(cache.Config{CacheDir: tmpDir, TTLDays: 1, MaxFiles: 100}, log)

	fetcher := &mockFetcher{
		getSubtitleFunc: func(ctx context.Context, videoID, language, subType, format string) ([]byte, error) {
			return nil, errors.New("download failed")
		},
	}

	meta := subtitle.SubtitleMetadata{
		Automatic: []subtitle.SubtitleTrack{
			{LanguageCode: "en", Formats: []string{"json3"}},
		},
	}

	svc := subtitle.NewSubtitleService(cacheMgr, log, fetcher)

	req := subtitle.DownloadRequest{
		VideoID:  "vid4",
		Language: "en",
		Type:     "automatic",
		Format:   "json3",
	}

	_, st, err := svc.DownloadSubtitleWithStats(context.Background(), req, meta)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if st.Attempts != 1 || st.Failures != 1 || st.CacheMisses != 1 || st.UpstreamFetches != 1 || st.AutomaticRequests != 1 {
		t.Errorf("unexpected stats: %+v", st)
	}
}
