package resource

import (
	"context"
	"sync"

	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
	"github.com/Arvind215271/yt2bulk/internal/youtube/pipeline"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

type Service struct {
	metadataService *metadata.Service
	pipelineService *pipeline.Service
	logger          *logger.Logger
	concurrency     int
}

func NewService(
	metadataService *metadata.Service,
	pipelineService *pipeline.Service,
	logger *logger.Logger,
	concurrency int,
) *Service {
	if concurrency <= 0 {
		concurrency = 1
	}
	return &Service{
		metadataService: metadataService,
		pipelineService: pipelineService,
		logger:          logger,
		concurrency:     concurrency,
	}
}

// ProcessResources processes a slice of YouTube inputs (videos and playlists) while strictly
// preserving the input request order, dispatching to appropriate video or playlist handlers,
// and returning a uniform slice of youtube.Resource containers.
func (s *Service) ProcessResources(
	ctx context.Context,
	inputs []youtubeurl.YouTubeInput,
	req *pipeline.Request,
) []youtube.Resource {
	resources, _ := s.ProcessResourcesWithStats(ctx, inputs, req)
	return resources
}

// ProcessResourcesWithStats processes resources and returns resources and aggregated process stats.
func (s *Service) ProcessResourcesWithStats(
	ctx context.Context,
	inputs []youtubeurl.YouTubeInput,
	req *pipeline.Request,
) ([]youtube.Resource, stats.ProcessStats) {
	var processStats stats.ProcessStats
	resources := make([]youtube.Resource, len(inputs))
	if ctx.Err() != nil {
		if s.logger != nil {
			s.logger.Debug("resource batch processing cancelled before start", "error", ctx.Err())
		}
		return resources, processStats
	}

	if len(inputs) == 0 {
		return nil, processStats
	}

	concurrency := s.concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, inputItem := range inputs {
		if ctx.Err() != nil {
			if s.logger != nil {
				s.logger.Debug("resource batch item unstarted skipped due to cancellation", "index", i, "error", ctx.Err())
			}
			break
		}
		wg.Add(1)
		go func(index int, inp youtubeurl.YouTubeInput) {
			defer wg.Done()

			if ctx.Err() != nil {
				if s.logger != nil {
					s.logger.Debug("resource item skipped at goroutine start due to cancellation", "error", ctx.Err())
				}
				return
			}

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				if s.logger != nil {
					s.logger.Debug("resource item skipped waiting for semaphore due to cancellation", "error", ctx.Err())
				}
				return
			}
			defer func() { <-sem }()

			if ctx.Err() != nil {
				if s.logger != nil {
					s.logger.Debug("resource item skipped after semaphore due to cancellation", "error", ctx.Err())
				}
				return
			}

			var res youtube.Resource
			var localProcessStats stats.ProcessStats

			switch inp.InputType {
			case youtubeurl.InputTypePlaylist:
				playlist, playlistResStats := ProcessPlaylistWithStats(
					ctx,
					s.metadataService,
					s.pipelineService,
					inp.ID,
					req,
					req.Limit,
					concurrency,
					s.logger,
				)
				if playlist == nil {
					if ctx.Err() != nil && s.logger != nil {
						s.logger.Debug("playlist resource operation completed during cancellation", "id", inp.ID, "error", ctx.Err())
					}
					return
				}
				res = youtube.Resource{
					ID:       inp.ID,
					Type:     youtube.ResourceTypePlaylist,
					Playlist: playlist,
				}
				localProcessStats.VideosProcessed += playlistResStats.VideosProcessed
				localProcessStats.Metadata.Add(playlistResStats.Metadata)
				localProcessStats.Subtitle.Add(playlistResStats.Subtitle)
				if len(playlist.Errors) > 0 {
					localProcessStats.ResourceFailures++
				} else {
					localProcessStats.ResourcesSucceeded++
				}
			case youtubeurl.InputTypeVideo:
				fallthrough
			default:
				video, videoResStats := ProcessVideoWithStats(
					ctx,
					s.pipelineService,
					inp.ID,
					req,
				)
				if video == nil {
					if ctx.Err() != nil && s.logger != nil {
						s.logger.Debug("video resource operation completed during cancellation", "id", inp.ID, "error", ctx.Err())
					}
					return
				}
				res = youtube.Resource{
					ID:    inp.ID,
					Type:  youtube.ResourceTypeVideo,
					Video: video,
				}
				localProcessStats.VideosProcessed += videoResStats.VideosProcessed
				localProcessStats.Metadata.Add(videoResStats.Metadata)
				localProcessStats.Subtitle.Add(videoResStats.Subtitle)
				if len(video.Errors) > 0 {
					localProcessStats.ResourceFailures++
				} else {
					localProcessStats.ResourcesSucceeded++
				}
			}

			resources[index] = res

			mu.Lock()
			processStats.ResourcesRequested++
			processStats.ResourcesSucceeded += localProcessStats.ResourcesSucceeded
			processStats.ResourceFailures += localProcessStats.ResourceFailures
			processStats.VideosProcessed += localProcessStats.VideosProcessed
			processStats.Metadata.Add(localProcessStats.Metadata)
			processStats.Subtitle.Add(localProcessStats.Subtitle)
			mu.Unlock()
		}(i, inputItem)
	}

	wg.Wait()
	if s.logger != nil {
		if ctx.Err() != nil {
			s.logger.Debug("resource batch processing completed as cancelled",
				"requested", processStats.ResourcesRequested,
				"succeeded", processStats.ResourcesSucceeded,
				"failed", processStats.ResourceFailures,
				"videos_processed", processStats.VideosProcessed,
				"error", ctx.Err(),
			)
		} else {
			s.logger.Debug("resource batch processing completed",
				"requested", processStats.ResourcesRequested,
				"succeeded", processStats.ResourcesSucceeded,
				"failed", processStats.ResourceFailures,
				"videos_processed", processStats.VideosProcessed,
				"metadata_upstream", processStats.Metadata.UpstreamFetches,
				"subtitle_upstream", processStats.Subtitle.UpstreamFetches,
			)
		}
	}
	return resources, processStats
}
