package resource

import (
	"context"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/pipeline"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

// ProcessVideo handles processing a single video URL/ID using the pipeline.
func ProcessVideo(ctx context.Context, pipelineService *pipeline.Service, videoID string, req *pipeline.Request) *youtube.Video {
	video, _ := ProcessVideoWithStats(ctx, pipelineService, videoID, req)
	return video
}

// ProcessVideoWithStats processes a single video and returns stats.
func ProcessVideoWithStats(ctx context.Context, pipelineService *pipeline.Service, videoID string, req *pipeline.Request) (*youtube.Video, stats.ResourceStats) {
	var resourceStats stats.ResourceStats
	resourceStats.VideosProcessed++

	if pipelineService == nil {
		video := &youtube.Video{
			ID:     videoID,
			Errors: []youtube.Error{{Message: "pipeline service is nil"}},
		}
		resourceStats.VideoFailures++
		return video, resourceStats
	}

	video, pipelineStats := pipelineService.ProcessFaultTolerantWithStats(ctx, videoID, req)
	if video == nil {
		video = &youtube.Video{ID: videoID}
	}

	resourceStats.Metadata.Add(pipelineStats.Metadata)
	resourceStats.Subtitle.Add(pipelineStats.Subtitle)
	if len(video.Errors) > 0 {
		resourceStats.VideoFailures++
	}

	return video, resourceStats
}
