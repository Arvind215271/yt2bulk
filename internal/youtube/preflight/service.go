package preflight

import (
	"context"
	"fmt"

	"github.com/Arvind215271/yt2bulk/internal/cache"
	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
)

// ServiceImpl implements the preflight.Service interface.
type ServiceImpl struct {
	metadataService *metadata.Service
	cacheManager    *cache.Manager
	logger          *logger.Logger
	maxLimit        int
}

// GetMaxLimit returns the maximum export limit configured.
func (s *ServiceImpl) GetMaxLimit() int {
	return s.maxLimit
}

// NewService creates a new preflight service.
func NewService(metadataService *metadata.Service, cacheManager *cache.Manager, logger *logger.Logger, maxLimit int) *ServiceImpl {
	return &ServiceImpl{
		metadataService: metadataService,
		cacheManager:    cacheManager,
		logger:          logger,
		maxLimit:        maxLimit,
	}
}

// Check performs the preflight enumeration and estimation.
func (s *ServiceImpl) Check(ctx context.Context, req Request) (Result, error) {
	if s.logger != nil {
		s.logger.Info("starting preflight check", "inputCount", len(req.Inputs))
	}

	// 1. Validate context.
	if len(req.Inputs) == 0 {
		err := fmt.Errorf("inputs are required")
		return Result{Valid: false, Error: err}, err
	}

	// 2. Resolve video and playlist resources.
	videoIDs, totalResources, playlistItemCount, preflightStats, err := s.resolveResources(ctx, req.Inputs)
	if err != nil {
		return Result{
			Valid:              false,
			Error:              err,
			TotalResourceCount: totalResources,
		}, err
	}

	// 3. Produce the overall estimate.
	estResult, err := s.estimateOverall(ctx, videoIDs)
	if err != nil {
		errWrapped := fmt.Errorf("failed to estimate: %w", err)
		return Result{Valid: false, Error: errWrapped}, errWrapped
	}

	// 4. Calculate final EstimatedWork and SubtitleFetch
	explicitMetadataMissing := estResult.MetadataMissing
	explicitSubtitleFetch := estResult.SubtitleFetch

	// Log these values to understand why EstimatedWork is 1
	if s.logger != nil {
		s.logger.Debug("estimating work",
			"explicitMetadataMissing", explicitMetadataMissing,
			"explicitSubtitleFetch", explicitSubtitleFetch,
			"playlistItemCount", playlistItemCount,
			"planNeedsSubtitle", estResult.Plan.NeedsSubtitle(),
		)
	}

	playlistMetadataMissing := playlistItemCount
	playlistSubtitleFetch := 0
	if estResult.Plan.NeedsSubtitle() {
		playlistSubtitleFetch = playlistItemCount
	}

	subtitleFetch := explicitSubtitleFetch + playlistSubtitleFetch

	estimatedWork := explicitMetadataMissing + explicitSubtitleFetch + playlistMetadataMissing + playlistSubtitleFetch

	// 5. Return the new preflight result.
	result := Result{
		Valid:              true,
		UniqueVideoCount:   len(videoIDs),
		TotalResourceCount: totalResources,
		EstimatedWork:      estimatedWork,
		SubtitleFetch:      subtitleFetch,
		Stats:              preflightStats,
	}

	if s.logger != nil {
		s.logger.Info("preflight check completed",
			"valid", result.Valid,
			"estimatedWork", result.EstimatedWork,
			"subtitleFetch", result.SubtitleFetch,
		)
	}

	return result, nil
}
