package pipeline

import (
	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/planner"
)

// ResourceContext holds the state for processing a resource through the pipeline.
type ResourceContext struct {
	Resource *youtube.Resource
	Video    *youtube.Video
	Playlist *youtube.Playlist
	Request  *Request
	Plan     *planner.ExecutionPlan
}
