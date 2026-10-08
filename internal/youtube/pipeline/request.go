package pipeline

import (
	"github.com/Arvind215271/askito/internal/youtube/fields"
	"github.com/Arvind215271/askito/internal/youtube/planner"
	"github.com/Arvind215271/askito/internal/youtube/signal"
	"github.com/Arvind215271/askito/internal/youtube/subtitle"
	"github.com/Arvind215271/askito/internal/youtube/transcript"
)

// Request composes feature-specific requests from each feature package.
// The pipeline only coordinates; validation and processing logic belongs to feature packages.
type Request struct {
	// FieldPlanner specifies which fields to include via the fields planner.
	FieldPlanner *fields.Planner

	// ExecutionPlan represents the execution stages determined by the planner.
	ExecutionPlan *planner.ExecutionPlan

	// Subtitle request for downloading subtitles.
	// Owned by the subtitle package; validation happens there.
	Subtitle *subtitle.DownloadRequest

	// Preferences for subtitle resolution.
	Preferences []subtitle.SubtitlePreference

	// Format for the subtitle download (e.g., json3, vtt).
	Format string

	// Transcript request for processing an already-parsed transcript.
	// Owned by the transcript package; validation happens there.
	Transcript *transcript.ProcessingRequest

	// Signal request for signal analysis on a transcript.
	// Owned by the signal package; validation happens there.
	Signal *signal.SignalRequest

	// Limit for bounded resource enumeration.
	Limit int
}
