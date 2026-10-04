package preflight

import (
	"context"

	youtubeurl "github.com/Arvind215271/askito/internal/youtube/input"
)

// Request defines the input for a preflight check.
type Request struct {
	Input *youtubeurl.YouTubeInput
}

// Result defines the output of a preflight check.
type Result struct {
	Valid              bool
	Error              error
	UniqueVideoCount   int
	TotalResourceCount int
	EstimatedWork      int
}

// Service defines the contract for preflight operations.
type Service interface {
	Check(ctx context.Context, req Request) (Result, error)
}
