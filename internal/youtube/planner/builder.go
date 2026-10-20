package planner

import (
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
)

// Build constructs an ExecutionPlan based on the requested resource types and field selections.
//
// Decision Tree & Optimization Flow:
//  1. Resource Classification: Determines if the inputs contain videos, playlists, or mixed resources.
//  2. Field Dependency Resolution: Checks the requested fields against field groups (Metadata, Description,
//     Transcript, Signal) and enforces prerequisites (e.g., Signal requires Transcript, Transcript requires
//     Subtitle, Subtitle and Description require Metadata).
//  3. Resource-Aware Optimization:
//     - For direct video inputs, any requested metadata fields trigger the Metadata stage.
//     - For playlist inputs, basic playlist item fields (like title, duration, id, etc.) can be satisfied
//     directly by playlist items without fetching individual video metadata. However, advanced fields
//     (like view count, like count, description parsing, or transcripts) trigger the Metadata stage.
//  4. Ordered Stage Compilation: Compiles all active capabilities into a strict, sequential list of
//     stages defined by OrderedStages.
func Build(resources []youtubeurl.YouTubeInput, fieldPlanner *fields.Planner) *ExecutionPlan {
	videosRequested := false
	playlistsRequested := false

	// Step 1: Scan and classify resource types
	for _, res := range resources {
		if res.InputType == youtubeurl.InputTypeVideo {
			videosRequested = true
		} else if res.InputType == youtubeurl.InputTypePlaylist {
			playlistsRequested = true
		}
	}

	mixedResources := videosRequested && playlistsRequested

	resourceInfo := ResourceInfo{
		VideosRequested:    videosRequested,
		PlaylistsRequested: playlistsRequested,
		MixedResources:     mixedResources,
	}

	// Initialize capability flags
	needsPlaylistItems := playlistsRequested
	needsMetadata := false
	needsDescription := false
	needsSubtitle := false
	needsTranscript := false
	needsSignal := false

	// Step 2: Evaluate field requirements from fieldPlanner
	if fieldPlanner == nil {
		// Default to all capabilities enabled if no field planner is provided
		needsMetadata = true
		needsDescription = true
		needsSubtitle = true
		needsTranscript = true
		needsSignal = true
	} else {
		needsDescription = fieldPlanner.HasAny(fields.DescriptionFields)
		needsTranscript = fieldPlanner.HasAny(fields.TranscriptFields) || fieldPlanner.ExportsEverything()
		needsSignal = fieldPlanner.HasAny(fields.SignalFields) || fieldPlanner.ExportsEverything()
		needsSubtitle = needsTranscript || needsSignal

		if fieldPlanner.ExportsEverything() {
			needsDescription = true
			needsMetadata = true
		} else {
			needsMetadata = fieldPlanner.HasAny(fields.MetadataFields) || needsDescription
		}
	}

	// Step 3: Enforce dependency resolution rules bottom-up/top-down
	if needsSignal {
		needsTranscript = true
	}
	if needsTranscript {
		needsSubtitle = true
	}
	if needsSubtitle {
		needsMetadata = true
	}
	if needsDescription {
		needsMetadata = true
	}

	// Step 4: Apply resource-specific optimization rules
	if videosRequested {
		if fieldPlanner == nil || fieldPlanner.HasAny(fields.MetadataFields) || fieldPlanner.ExportsEverything() {
			needsMetadata = true
		}
	} else if playlistsRequested {
		// Define fields that can be satisfied entirely from playlist items without fetching individual video metadata
		basicPlaylistFields := map[string]bool{
			fields.FieldID:                true,
			fields.FieldTitle:             true,
			fields.FieldChannelID:         true,
			fields.FieldChannelTitle:      true,
			fields.FieldThumbnails:        true,
			fields.FieldPublishedAt:       true,
			fields.FieldDuration:          true,
			fields.FieldDurationSeconds:   true,
			fields.FieldDurationMinutes:   true,
			fields.FieldDurationTimestamp: true,
			fields.FieldErrors:            true,
		}

		if fieldPlanner == nil || fieldPlanner.ExportsEverything() {
			needsMetadata = true
		} else {
			requiresMetadataStage := false
			for _, f := range fieldPlanner.ExportFields() {
				if !basicPlaylistFields[f] {
					requiresMetadataStage = true
					break
				}
			}
			if requiresMetadataStage || needsDescription || needsSubtitle {
				needsMetadata = true
			} else {
				// Optimization: If only basic playlist fields are requested, skip the Metadata stage entirely
				needsMetadata = false
			}
		}
	}

	// Double-check prerequisite cascading
	if needsDescription {
		needsMetadata = true
	}
	if needsSubtitle {
		needsMetadata = true
	}

	// Step 5: Compile ordered execution stages
	var stages []Stage
	for _, s := range OrderedStages {
		include := false
		switch s {
		case StagePlaylistItems:
			include = needsPlaylistItems
		case StageMetadata:
			include = needsMetadata
		case StageDescription:
			include = needsDescription
		case StageSubtitle:
			include = needsSubtitle
		case StageTranscript:
			include = needsTranscript
		case StageSignal:
			include = needsSignal
		}
		if include {
			stages = append(stages, s)
		}
	}

	return &ExecutionPlan{
		Stages:    stages,
		Resources: resourceInfo,
	}
}
