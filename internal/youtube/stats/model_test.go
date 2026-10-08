package stats

import (
	"testing"
)

func TestMetadataStats_Add(t *testing.T) {
	var s MetadataStats // zero value test
	if s.Attempts != 0 || s.Successes != 0 || s.Failures != 0 {
		t.Errorf("expected zero values, got %+v", s)
	}

	s.Add(MetadataStats{
		Attempts:                 10,
		Successes:                8,
		Failures:                 2,
		CacheHits:                5,
		CacheMisses:              5,
		UpstreamFetches:          5,
		EstimatedUpstreamFetches: 2,
		YTDLPRequests:            4,
		APIRequests:              1,
		Fallbacks:                1,
	})

	s.Add(MetadataStats{
		Attempts:                 5,
		Successes:                5,
		Failures:                 0,
		CacheHits:                2,
		CacheMisses:              3,
		UpstreamFetches:          3,
		EstimatedUpstreamFetches: 1,
		YTDLPRequests:            2,
		APIRequests:              1,
		Fallbacks:                0,
	})

	if s.Attempts != 15 || s.Successes != 13 || s.Failures != 2 ||
		s.CacheHits != 7 || s.CacheMisses != 8 || s.UpstreamFetches != 8 ||
		s.EstimatedUpstreamFetches != 3 ||
		s.YTDLPRequests != 6 || s.APIRequests != 2 || s.Fallbacks != 1 {
		t.Errorf("unexpected aggregated MetadataStats: %+v", s)
	}
}

func TestSubtitleStats_Add(t *testing.T) {
	var s SubtitleStats
	if s.Attempts != 0 || s.Successes != 0 {
		t.Errorf("expected zero values, got %+v", s)
	}

	s.Add(SubtitleStats{
		Attempts:                 6,
		Successes:                5,
		Failures:                 1,
		CacheHits:                3,
		CacheMisses:              3,
		UpstreamFetches:          3,
		PotentialSubtitleFetches: 1,
		Fallbacks:                1,
		ManualRequests:           2,
		AutomaticRequests:        4,
	})

	s.Add(SubtitleStats{
		Attempts:                 4,
		Successes:                4,
		Failures:                 0,
		CacheHits:                1,
		CacheMisses:              3,
		UpstreamFetches:          3,
		PotentialSubtitleFetches: 2,
		Fallbacks:                0,
		ManualRequests:           1,
		AutomaticRequests:        3,
	})

	if s.Attempts != 10 || s.Successes != 9 || s.Failures != 1 ||
		s.CacheHits != 4 || s.CacheMisses != 6 || s.UpstreamFetches != 6 ||
		s.PotentialSubtitleFetches != 3 ||
		s.Fallbacks != 1 || s.ManualRequests != 3 || s.AutomaticRequests != 7 {
		t.Errorf("unexpected aggregated SubtitleStats: %+v", s)
	}
}

func TestPlaylistStats_Add(t *testing.T) {
	var s PlaylistStats
	s.Add(PlaylistStats{
		Attempts:        2,
		Successes:       2,
		Failures:        0,
		ItemsDiscovered: 25,
		Fallbacks:       1,
		Metadata: MetadataStats{
			Attempts:  2,
			Successes: 2,
		},
	})

	s.Add(PlaylistStats{
		Attempts:        1,
		Successes:       0,
		Failures:        1,
		ItemsDiscovered: 10,
		Fallbacks:       0,
		Metadata: MetadataStats{
			Attempts: 1,
			Failures: 1,
		},
	})

	if s.Attempts != 3 || s.Successes != 2 || s.Failures != 1 ||
		s.ItemsDiscovered != 35 || s.Fallbacks != 1 ||
		s.Metadata.Attempts != 3 || s.Metadata.Successes != 2 || s.Metadata.Failures != 1 {
		t.Errorf("unexpected aggregated PlaylistStats: %+v", s)
	}
}

func TestPipelineStats_Add(t *testing.T) {
	var s PipelineStats
	s.Metadata.Attempts = 5
	s.Subtitle.Attempts = 5

	s.Add(PipelineStats{
		Metadata: MetadataStats{Attempts: 3},
		Subtitle: SubtitleStats{Attempts: 2},
	})

	if s.Metadata.Attempts != 8 || s.Subtitle.Attempts != 7 {
		t.Errorf("unexpected aggregated PipelineStats: %+v", s)
	}
}

func TestResourceStats_Add(t *testing.T) {
	var s ResourceStats
	s.VideosProcessed = 10
	s.VideoFailures = 2
	s.Metadata.Attempts = 10
	s.Subtitle.Attempts = 8

	s.Add(ResourceStats{
		VideosProcessed: 5,
		VideoFailures:   1,
		Metadata:        MetadataStats{Attempts: 5},
		Subtitle:        SubtitleStats{Attempts: 5},
	})

	if s.VideosProcessed != 15 || s.VideoFailures != 3 ||
		s.Metadata.Attempts != 15 || s.Subtitle.Attempts != 13 {
		t.Errorf("unexpected aggregated ResourceStats: %+v", s)
	}
}

func TestProcessStats_Add(t *testing.T) {
	var s ProcessStats
	s.ResourcesRequested = 20
	s.ResourcesSucceeded = 18
	s.ResourceFailures = 2
	s.VideosProcessed = 18

	s.Add(ProcessStats{
		ResourcesRequested: 10,
		ResourcesSucceeded: 9,
		ResourceFailures:   1,
		VideosProcessed:    9,
	})

	if s.ResourcesRequested != 30 || s.ResourcesSucceeded != 27 ||
		s.ResourceFailures != 3 || s.VideosProcessed != 27 {
		t.Errorf("unexpected aggregated ProcessStats: %+v", s)
	}
}

func TestMultiChildMerging(t *testing.T) {
	// Simulate multi-worker accumulation in concurrent pipeline/resource processing
	workers := []ResourceStats{
		{
			VideosProcessed: 10,
			VideoFailures:   0,
			Metadata:        MetadataStats{Attempts: 10, Successes: 10},
			Subtitle:        SubtitleStats{Attempts: 10, Successes: 9, Failures: 1},
		},
		{
			VideosProcessed: 15,
			VideoFailures:   1,
			Metadata:        MetadataStats{Attempts: 15, Successes: 14, Failures: 1},
			Subtitle:        SubtitleStats{Attempts: 15, Successes: 15},
		},
		{
			VideosProcessed: 5,
			VideoFailures:   2,
			Metadata:        MetadataStats{Attempts: 5, Successes: 3, Failures: 2},
			Subtitle:        SubtitleStats{Attempts: 5, Successes: 3, Failures: 2},
		},
	}

	var total ResourceStats
	for _, w := range workers {
		total.Add(w)
	}

	if total.VideosProcessed != 30 || total.VideoFailures != 3 ||
		total.Metadata.Attempts != 30 || total.Metadata.Successes != 27 || total.Metadata.Failures != 3 ||
		total.Subtitle.Attempts != 30 || total.Subtitle.Successes != 27 || total.Subtitle.Failures != 3 {
		t.Errorf("unexpected multi-child merged ResourceStats: %+v", total)
	}
}
