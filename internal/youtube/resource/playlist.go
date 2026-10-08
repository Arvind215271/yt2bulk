package resource

import (
	"context"
	"fmt"
	"sync"

	"github.com/Arvind215271/askito/internal/logger"
	"github.com/Arvind215271/askito/internal/youtube"
	"github.com/Arvind215271/askito/internal/youtube/metadata"
	"github.com/Arvind215271/askito/internal/youtube/pipeline"
	"github.com/Arvind215271/askito/internal/youtube/stats"
)

// ProcessPlaylist fetches playlist items, decides whether the pipeline is needed based on execution plan,
// maps playlist items to videos, processes them concurrently, and attaches them back to the Playlist model.
func ProcessPlaylist(
	ctx context.Context,
	metadataService *metadata.Service,
	pipelineService *pipeline.Service,
	playlistID string,
	req *pipeline.Request,
	limit int,
	concurrency int,
	logger *logger.Logger,
) *youtube.Playlist {
	playlist, _ := ProcessPlaylistWithStats(ctx, metadataService, pipelineService, playlistID, req, limit, concurrency, logger)
	return playlist
}

// ProcessPlaylistWithStats fetches playlist items and processes them, returning playlist and resource stats.
func ProcessPlaylistWithStats(
	ctx context.Context,
	metadataService *metadata.Service,
	pipelineService *pipeline.Service,
	playlistID string,
	req *pipeline.Request,
	limit int,
	concurrency int,
	logger *logger.Logger,
) (*youtube.Playlist, stats.ResourceStats) {
	playlist := &youtube.Playlist{
		ID: playlistID,
	}
	var resourceStats stats.ResourceStats

	if ctx.Err() != nil {
		if logger != nil {
			logger.Debug("playlist processing cancelled before start", "playlistID", playlistID, "error", ctx.Err())
		}
		return playlist, resourceStats
	}

	if metadataService == nil {
		playlist.Errors = append(playlist.Errors, youtube.Error{Message: "metadata service is nil"})
		return playlist, resourceStats
	}

	// 1. Fetch playlist items with stats
	items, metaStats, err := metadataService.GetPlaylistItemsWithStats(ctx, playlistID, limit, metadata.ProviderYTDLP)
	resourceStats.Metadata.Add(metaStats)
	if err != nil {
		if ctx.Err() != nil && logger != nil {
			logger.Warn("playlist items fetch failed during cancellation (ytdlp)", "playlistID", playlistID, "error", err)
		}
		items, metaStats, err = metadataService.GetPlaylistItemsWithStats(ctx, playlistID, limit, metadata.ProviderAPI)
		resourceStats.Metadata.Add(metaStats)
		if err != nil {
			if ctx.Err() != nil && logger != nil {
				logger.Warn("playlist items fetch failed during cancellation (api)", "playlistID", playlistID, "error", err)
			}
			playlist.Errors = append(playlist.Errors, youtube.Error{Message: fmt.Sprintf("failed to fetch playlist items: %v", err)})
			return playlist, resourceStats
		}
	} else if ctx.Err() != nil && logger != nil {
		logger.Debug("playlist items fetch finished successfully during cancellation", "playlistID", playlistID)
	}

	playlist.Items = items
	playlist.ItemCount = len(items)

	if len(items) == 0 {
		return playlist, resourceStats
	}

	// Determine if pipeline processing is needed
	needsPipeline := true
	if req != nil && req.ExecutionPlan != nil {
		plan := req.ExecutionPlan
		if !plan.NeedsMetadata() && !plan.NeedsDescription() && !plan.NeedsSubtitle() && !plan.NeedsTranscript() && !plan.NeedsSignal() {
			needsPipeline = false
		}
	}

	playlistVideos := make([]youtube.PlaylistVideo, len(items))

	if !needsPipeline {
		for i, item := range items {
			if ctx.Err() != nil {
				if logger != nil {
					logger.Debug("playlist item skipped (no pipeline) due to cancellation", "playlistID", playlistID, "index", i, "error", ctx.Err())
				}
				break
			}
			video := ItemToVideo(item)
			playlistVideos[i] = youtube.PlaylistVideo{
				Video:    video,
				Position: item.Position,
				AddedAt:  item.AddedAt,
			}
		}
		playlist.Videos = playlistVideos
		resourceStats.VideosProcessed += len(items)
		return playlist, resourceStats
	}

	if concurrency <= 0 {
		concurrency = 1
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, item := range items {
		if ctx.Err() != nil {
			if logger != nil {
				logger.Debug("playlist item unstarted skipped due to cancellation", "playlistID", playlistID, "index", i, "error", ctx.Err())
			}
			break
		}
		wg.Add(1)
		go func(index int, pi youtube.PlaylistItem) {
			defer wg.Done()

			if ctx.Err() != nil {
				if logger != nil {
					logger.Debug("playlist item skipped at goroutine start due to cancellation", "playlistID", playlistID, "error", ctx.Err())
				}
				return
			}

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				if logger != nil {
					logger.Debug("playlist item skipped waiting for semaphore due to cancellation", "playlistID", playlistID, "error", ctx.Err())
				}
				return
			}
			defer func() { <-sem }()

			if ctx.Err() != nil {
				if logger != nil {
					logger.Debug("playlist item skipped after semaphore due to cancellation", "playlistID", playlistID, "error", ctx.Err())
				}
				return
			}

			baseVideo := ItemToVideo(pi)
			var processedVideo *youtube.Video
			var videoResStats stats.ResourceStats
			if pipelineService != nil {
				rc := &pipeline.ResourceContext{
					Video:   &baseVideo,
					Request: req,
					Plan:    req.ExecutionPlan,
				}
				pVideo, pStats := pipelineService.ProcessResourceWithStats(ctx, rc, true)
				processedVideo = pVideo
				videoResStats.VideosProcessed++
				videoResStats.Metadata.Add(pStats.Metadata)
				videoResStats.Subtitle.Add(pStats.Subtitle)
				if pVideo != nil && len(pVideo.Errors) > 0 {
					videoResStats.VideoFailures++
				}
			} else {
				processedVideo = &baseVideo
				videoResStats.VideosProcessed++
			}

			if processedVideo == nil {
				processedVideo = &baseVideo
			}

			playlistVideos[index] = youtube.PlaylistVideo{
				Video:    *processedVideo,
				Position: pi.Position,
				AddedAt:  pi.AddedAt,
			}

			mu.Lock()
			resourceStats.Add(videoResStats)
			mu.Unlock()
		}(i, item)
	}

	wg.Wait()
	playlist.Videos = playlistVideos

	if ctx.Err() != nil && logger != nil {
		logger.Debug("playlist processing completed as cancelled", "playlistID", playlistID, "error", ctx.Err(), "videos_processed", resourceStats.VideosProcessed)
	}

	return playlist, resourceStats
}
