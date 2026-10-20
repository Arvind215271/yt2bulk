package pipeline

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/description"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
	"github.com/Arvind215271/yt2bulk/internal/youtube/signal"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
	"github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
)

type Service struct {
	metadataService    *metadata.Service
	descriptionService *description.Service
	subtitleService    *subtitle.SubtitleService
	transcriptService  *transcript.Service
	signalService      *signal.SignalService

	logger      *logger.Logger
	concurrency int
}

func NewService(
	metadataService *metadata.Service,
	descriptionService *description.Service,
	subtitleService *subtitle.SubtitleService,
	transcriptService *transcript.Service,
	signalService *signal.SignalService,
	logger *logger.Logger,
	concurrency int,
) *Service {
	if concurrency <= 0 {
		concurrency = 1
	}

	return &Service{
		metadataService:    metadataService,
		descriptionService: descriptionService,
		subtitleService:    subtitleService,
		transcriptService:  transcriptService,
		signalService:      signalService,
		logger:             logger,
		concurrency:        concurrency,
	}
}

func cleanErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	s := err.Error()
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "File \"") ||
			strings.HasPrefix(trimmed, "Traceback") ||
			strings.HasPrefix(trimmed, "During handling") ||
			strings.HasPrefix(trimmed, "The above exception") ||
			(strings.HasPrefix(line, "    ") && !strings.Contains(trimmed, "Error")) {
			continue
		}
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) > 0 {
		return strings.Join(result, " — ")
	}
	return s
}

func (s *Service) ProcessResourceWithStats(
	ctx context.Context,
	rc *ResourceContext,
	faultTolerant bool,
) (*youtube.Video, stats.PipelineStats) {
	var pipelineStats stats.PipelineStats
	videoID := ""
	if rc != nil && rc.Video != nil {
		videoID = rc.Video.ID
	}

	if ctx.Err() != nil {
		if s.logger != nil {
			s.logger.Debug("pipeline processing cancelled before start", "videoID", videoID, "error", ctx.Err())
		}
		if faultTolerant {
			if rc != nil && rc.Video != nil {
				return rc.Video, pipelineStats
			}
			return &youtube.Video{}, pipelineStats
		}
		return nil, pipelineStats
	}

	if rc.Video == nil {
		rc.Video = &youtube.Video{}
	}
	video := rc.Video
	videoID = video.ID

	if rc.Request == nil || rc.Request.ExecutionPlan == nil {
		err := fmt.Errorf("pipeline request or execution plan is nil")
		if faultTolerant {
			video.Errors = append(video.Errors, youtube.Error{Message: cleanErrorMessage(err)})
			return video, pipelineStats
		}
		return nil, pipelineStats
	}

	plan := rc.Request.ExecutionPlan

	// 1. Metadata Stage
	if ctx.Err() != nil {
		if s.logger != nil {
			s.logger.Debug("pipeline stage skipped due to cancellation", "videoID", videoID, "stage", "metadata", "error", ctx.Err())
		}
		if faultTolerant {
			return video, pipelineStats
		}
		return nil, pipelineStats
	}
	if plan.NeedsMetadata() {
		metaStats, err := ProcessMetadata(ctx, rc, s.metadataService)
		pipelineStats.Metadata.Add(metaStats)
		if err != nil {
			if ctx.Err() != nil {
				if s.logger != nil {
					s.logger.Warn("metadata operation failed during cancellation", "videoID", videoID, "error", err)
				}
			} else {
				s.logError("metadata processing failed", "videoID", videoID, "error", err)
			}
			if faultTolerant {
				video.Errors = append(video.Errors, youtube.Error{Message: fmt.Sprintf("metadata fetch failed: %s", cleanErrorMessage(err))})
				return video, pipelineStats
			}
			return nil, pipelineStats
		} else if ctx.Err() != nil && s.logger != nil {
			s.logger.Debug("metadata operation finished successfully during cancellation", "videoID", videoID)
		}
	}

	// 2. Description Stage
	if ctx.Err() != nil {
		if s.logger != nil {
			s.logger.Debug("pipeline stage skipped due to cancellation", "videoID", videoID, "stage", "description", "error", ctx.Err())
		}
		if faultTolerant {
			return video, pipelineStats
		}
		return nil, pipelineStats
	}
	if plan.NeedsDescription() {
		err := ProcessDescription(ctx, rc, s.descriptionService)
		if err != nil {
			if ctx.Err() != nil {
				if s.logger != nil {
					s.logger.Warn("description operation failed during cancellation", "videoID", videoID, "error", err)
				}
			} else {
				s.logError("description processing failed", "videoID", videoID, "error", err)
			}
			if faultTolerant {
				video.Errors = append(video.Errors, youtube.Error{Message: fmt.Sprintf("description processing failed: %s", cleanErrorMessage(err))})
			} else {
				return nil, pipelineStats
			}
		} else if ctx.Err() != nil && s.logger != nil {
			s.logger.Debug("description operation finished successfully during cancellation", "videoID", videoID)
		}
	}

	// 3. Subtitle + Transcript Stage
	if ctx.Err() != nil {
		if s.logger != nil {
			s.logger.Debug("pipeline stage skipped due to cancellation", "videoID", videoID, "stage", "subtitle_transcript", "error", ctx.Err())
		}
		if faultTolerant {
			return video, pipelineStats
		}
		return nil, pipelineStats
	}
	var trans *transcript.Transcript
	if plan.NeedsSubtitle() || plan.NeedsTranscript() || plan.NeedsSignal() {
		var subStats stats.SubtitleStats
		var err error
		trans, subStats, err = ProcessSubtitleAndTranscript(ctx, rc, s.subtitleService, s.transcriptService)
		pipelineStats.Subtitle.Add(subStats)
		if err != nil {
			if ctx.Err() != nil {
				if s.logger != nil {
					s.logger.Warn("subtitle/transcript operation failed during cancellation", "videoID", videoID, "error", err)
				}
			} else {
				s.logError("subtitle/transcript processing failed", "videoID", videoID, "error", err)
			}
			if faultTolerant {
				video.Errors = append(video.Errors, youtube.Error{Message: fmt.Sprintf("subtitle/transcript processing failed: %s", cleanErrorMessage(err))})
			} else {
				return nil, pipelineStats
			}
		} else if ctx.Err() != nil && s.logger != nil {
			s.logger.Debug("subtitle/transcript operation finished successfully during cancellation", "videoID", videoID)
		}
	}

	// 4. Signal Stage
	if ctx.Err() != nil {
		if s.logger != nil {
			s.logger.Debug("pipeline stage skipped due to cancellation", "videoID", videoID, "stage", "signal", "error", ctx.Err())
		}
		if faultTolerant {
			return video, pipelineStats
		}
		return nil, pipelineStats
	}
	if plan.NeedsSignal() && trans != nil {
		err := ProcessSignal(ctx, rc, s.signalService, trans)
		if err != nil {
			if ctx.Err() != nil {
				if s.logger != nil {
					s.logger.Warn("signal operation failed during cancellation", "videoID", videoID, "error", err)
				}
			} else {
				s.logError("signal processing failed", "videoID", videoID, "error", err)
			}
			if faultTolerant {
				video.Errors = append(video.Errors, youtube.Error{Message: fmt.Sprintf("signal processing failed: %s", cleanErrorMessage(err))})
			} else {
				return nil, pipelineStats
			}
		} else if ctx.Err() != nil && s.logger != nil {
			s.logger.Debug("signal operation finished successfully during cancellation", "videoID", videoID)
		}
	}

	if ctx.Err() != nil && s.logger != nil {
		s.logger.Debug("pipeline processing completed as cancelled", "videoID", videoID, "error", ctx.Err(), "metadata_fetches", pipelineStats.Metadata.UpstreamFetches, "subtitle_fetches", pipelineStats.Subtitle.UpstreamFetches)
	}

	return video, pipelineStats
}

func (s *Service) ProcessResource(
	ctx context.Context,
	rc *ResourceContext,
	faultTolerant bool,
) *youtube.Video {
	video, _ := s.ProcessResourceWithStats(ctx, rc, faultTolerant)
	return video
}

func (s *Service) ProcessWithStats(
	ctx context.Context,
	videoID string,
	req *Request,
) (*youtube.Video, stats.PipelineStats, error) {
	if req == nil || req.ExecutionPlan == nil {
		return nil, stats.PipelineStats{}, fmt.Errorf("pipeline request or execution plan is nil for video %s", videoID)
	}

	rc := &ResourceContext{
		Video:   &youtube.Video{ID: videoID},
		Request: req,
		Plan:    req.ExecutionPlan,
	}

	video, pipelineStats := s.ProcessResourceWithStats(ctx, rc, false)
	if video == nil || len(video.Errors) > 0 {
		return nil, pipelineStats, fmt.Errorf("failed to process video %s", videoID)
	}

	return video, pipelineStats, nil
}

func (s *Service) Process(
	ctx context.Context,
	videoID string,
	req *Request,
) (*youtube.Video, error) {
	video, _, err := s.ProcessWithStats(ctx, videoID, req)
	return video, err
}

func (s *Service) ProcessFaultTolerantWithStats(
	ctx context.Context,
	videoID string,
	req *Request,
) (*youtube.Video, stats.PipelineStats) {
	if req == nil || req.ExecutionPlan == nil {
		return &youtube.Video{
			ID:     videoID,
			Errors: []youtube.Error{{Message: "pipeline request or execution plan is nil"}},
		}, stats.PipelineStats{}
	}

	rc := &ResourceContext{
		Video:   &youtube.Video{ID: videoID},
		Request: req,
		Plan:    req.ExecutionPlan,
	}

	return s.ProcessResourceWithStats(ctx, rc, true)
}

func (s *Service) ProcessFaultTolerant(
	ctx context.Context,
	videoID string,
	req *Request,
) *youtube.Video {
	video, _ := s.ProcessFaultTolerantWithStats(ctx, videoID, req)
	return video
}

func (s *Service) ProcessVideos(
	ctx context.Context,
	videoIDs []string,
	req *Request,
) []*youtube.Video {
	if len(videoIDs) == 0 {
		return nil
	}

	results := make([]*youtube.Video, len(videoIDs))
	concurrency := s.concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, videoID := range videoIDs {
		if ctx.Err() != nil {
			if s.logger != nil {
				s.logger.Debug("videos batch item unstarted skipped due to cancellation", "index", i, "error", ctx.Err())
			}
			break
		}
		wg.Add(1)
		go func(index int, id string) {
			defer wg.Done()

			if ctx.Err() != nil {
				if s.logger != nil {
					s.logger.Debug("video item skipped at goroutine start due to cancellation", "videoID", id, "error", ctx.Err())
				}
				return
			}

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				if s.logger != nil {
					s.logger.Debug("video item skipped waiting for semaphore due to cancellation", "videoID", id, "error", ctx.Err())
				}
				return
			}
			defer func() { <-sem }()

			if ctx.Err() != nil {
				if s.logger != nil {
					s.logger.Debug("video item skipped after semaphore due to cancellation", "videoID", id, "error", ctx.Err())
				}
				return
			}

			results[index] = s.ProcessFaultTolerant(ctx, id, req)
		}(i, videoID)
	}

	wg.Wait()
	if s.logger != nil {
		if ctx.Err() != nil {
			s.logger.Debug("videos batch processing completed as cancelled", "error", ctx.Err())
		} else {
			s.logger.Debug("videos batch processing completed")
		}
	}
	return results
}

func (s *Service) logDebug(msg string, args ...any) {
	if s.logger == nil {
		return
	}
	s.logger.Debug(msg, args...)
}

func (s *Service) logWarn(msg string, args ...any) {
	if s.logger == nil {
		return
	}
	s.logger.Warn(msg, args...)
}

func (s *Service) logError(msg string, args ...any) {
	if s.logger == nil {
		return
	}
	s.logger.Error(msg, args...)
}
