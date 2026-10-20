package planner

import (
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
)

// Planner orchestrates the creation of ExecutionPlans.
type Planner struct{}

// New creates a new execution Planner instance.
func New() *Planner {
	return &Planner{}
}

// Build creates an ExecutionPlan given a list of input resources and a field planner.
func (p *Planner) Build(resources []youtubeurl.YouTubeInput, fieldPlanner *fields.Planner) *ExecutionPlan {
	return Build(resources, fieldPlanner)
}

// Plan is a shorthand static or method helper for building an execution plan directly.
func Plan(resources []youtubeurl.YouTubeInput, fieldPlanner *fields.Planner) *ExecutionPlan {
	return Build(resources, fieldPlanner)
}
