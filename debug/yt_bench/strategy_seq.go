package yt_bench

import (
	"context"
	"time"

	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata/ytdlp"
)

func RunStrategyA(ctx context.Context, client *ytdlp.Client, videoIDs []string) (BenchmarkResult, error) {
	start := time.Now()

	// Implementation: Loop and call GetVideo
	for _, id := range videoIDs {
		_, _ = client.GetVideo(ctx, id, nil)
	}

	return BenchmarkResult{
		Strategy:           "Sequential",
		PlaylistSize:       len(videoIDs),
		TotalExecutionTime: time.Since(start),
	}, nil
}
