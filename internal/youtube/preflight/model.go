 package preflight

import (
	"context"

	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/planner"
)

// Request defines the input for a preflight check.
type Request struct {
	Inputs []*youtubeurl.YouTubeInput
}

// Result defines the output of a preflight check.
type Result struct {
	Valid              bool
	Error              error
	UniqueVideoCount   int
	TotalResourceCount int
	EstimatedWork      int
	SubtitleFetch      int
	Stats              ResourcePreflight
}

// Service defines the contract for preflight operations.
type Service interface {
	Check(ctx context.Context, req Request) (Result, error)
	GetMaxLimit() int
}

// Estimator interface for estimating resource requirements.
type Estimator interface {
	Estimate(ctx context.Context, videoIDs []string, fieldPlanner *fields.Planner) (EstimationResult, error)
	EstimateResource(ctx context.Context, resources []youtubeurl.YouTubeInput, fieldPlanner *fields.Planner) (EstimationResult, error)
}

// EstimationResult defines the output of an estimation.
type EstimationResult struct {
	Plan            *planner.ExecutionPlan
	MetadataMissing int
	MetadataPresent int
	SubtitleFetch   int // Potential
	Stats           ResourcePreflight
}
