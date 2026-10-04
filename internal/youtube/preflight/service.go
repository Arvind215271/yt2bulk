package preflight

import (
	"context"
	"fmt"

	"github.com/Arvind215271/askito/internal/youtube/metadata"
)

// ServiceImpl implements the preflight.Service interface.
type ServiceImpl struct {
	metadataService *metadata.Service
}

// NewService creates a new preflight service.
func NewService(metadataService *metadata.Service) *ServiceImpl {
	return &ServiceImpl{
		metadataService: metadataService,
	}
}

// Check performs the preflight enumeration and estimation.
func (s *ServiceImpl) Check(ctx context.Context, req Request) (Result, error) {
	if req.Input == nil {
		return Result{Valid: false, Error: fmt.Errorf("input is required")}, nil
	}

	uniqueVideos := make(map[string]struct{})
	totalResources := 0

	switch req.Input.InputType {
	case "video":
		uniqueVideos[req.Input.ID] = struct{}{}
		totalResources = 1
	case "playlist":
		items, err := s.metadataService.GetPlaylistItems(ctx, req.Input.ID, metadata.ProviderYTDLP)
		if err != nil {
			// Fallback to API if YTDLP fails
			items, err = s.metadataService.GetPlaylistItems(ctx, req.Input.ID, metadata.ProviderAPI)
			if err != nil {
				return Result{Valid: false, Error: fmt.Errorf("failed to fetch playlist items: %w", err)}, nil
			}
		}

		for _, item := range items {
			uniqueVideos[item.VideoID] = struct{}{}
		}
		totalResources = len(items)
	default:
		return Result{Valid: false, Error: fmt.Errorf("unsupported input type: %s", req.Input.InputType)}, nil
	}

	return Result{
		Valid:              true,
		UniqueVideoCount:   len(uniqueVideos),
		TotalResourceCount: totalResources,
		EstimatedWork:      len(uniqueVideos),
	}, nil
}
