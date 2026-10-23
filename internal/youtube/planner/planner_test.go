package planner

import (
	"testing"

	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
)

func TestPlanner_PlaylistTitle(t *testing.T) {
	resources := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypePlaylist, ID: "PL123"},
	}
	fp, err := fields.NewPlanner([]string{fields.FieldTitle}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan := New().Build(resources, fp)

	if !plan.NeedsPlaylistItems() {
		t.Errorf("expected NeedsPlaylistItems to be true")
	}
	if plan.NeedsMetadata() {
		t.Errorf("expected NeedsMetadata to be false for playlist title")
	}
	if plan.NeedsSubtitle() {
		t.Errorf("expected NeedsSubtitle to be false")
	}
	if plan.NeedsTranscript() {
		t.Errorf("expected NeedsTranscript to be false")
	}
	if plan.NeedsSignal() {
		t.Errorf("expected NeedsSignal to be false")
	}
}

func TestPlanner_PlaylistDuration(t *testing.T) {
	resources := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypePlaylist, ID: "PL123"},
	}
	fp, err := fields.NewPlanner([]string{fields.FieldDuration}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan := Plan(resources, fp)

	if !plan.NeedsPlaylistItems() {
		t.Errorf("expected NeedsPlaylistItems to be true")
	}
	if plan.NeedsMetadata() {
		t.Errorf("expected NeedsMetadata to be false for playlist duration")
	}
}

func TestPlanner_PlaylistLikeCount(t *testing.T) {
	resources := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypePlaylist, ID: "PL123"},
	}
	fp, err := fields.NewPlanner([]string{fields.FieldLikeCount}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan := Plan(resources, fp)

	if !plan.NeedsPlaylistItems() {
		t.Errorf("expected NeedsPlaylistItems to be true")
	}
	if !plan.NeedsMetadata() {
		t.Errorf("expected NeedsMetadata to be true for playlist like_count")
	}
	if plan.NeedsSubtitle() {
		t.Errorf("expected NeedsSubtitle to be false")
	}
}

func TestPlanner_PlaylistTranscript(t *testing.T) {
	resources := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypePlaylist, ID: "PL123"},
	}
	fp, err := fields.NewPlanner([]string{fields.FieldTranscriptText}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan := Plan(resources, fp)

	if !plan.NeedsPlaylistItems() {
		t.Errorf("expected NeedsPlaylistItems to be true")
	}
	if !plan.NeedsMetadata() {
		t.Errorf("expected NeedsMetadata to be true")
	}
	if !plan.NeedsSubtitle() {
		t.Errorf("expected NeedsSubtitle to be true")
	}
	if !plan.NeedsTranscript() {
		t.Errorf("expected NeedsTranscript to be true")
	}
}

func TestPlanner_VideoTitle(t *testing.T) {
	resources := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypeVideo, ID: "vid123"},
	}
	fp, err := fields.NewPlanner([]string{fields.FieldTitle})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan := Plan(resources, fp)

	if plan.NeedsPlaylistItems() {
		t.Errorf("expected NeedsPlaylistItems to be false for video")
	}
	if !plan.NeedsMetadata() {
		t.Errorf("expected NeedsMetadata to be true for video title")
	}
	if plan.NeedsSubtitle() {
		t.Errorf("expected NeedsSubtitle to be false")
	}
}

func TestPlanner_VideoSignal(t *testing.T) {
	resources := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypeVideo, ID: "vid123"},
	}
	fp, err := fields.NewPlanner([]string{fields.FieldTranscriptSignal})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan := Plan(resources, fp)

	if plan.NeedsPlaylistItems() {
		t.Errorf("expected NeedsPlaylistItems to be false")
	}
	if !plan.NeedsMetadata() {
		t.Errorf("expected NeedsMetadata to be true")
	}
	if !plan.NeedsSubtitle() {
		t.Errorf("expected NeedsSubtitle to be true")
	}
	if !plan.NeedsTranscript() {
		t.Errorf("expected NeedsTranscript to be true")
	}
	if !plan.NeedsSignal() {
		t.Errorf("expected NeedsSignal to be true")
	}
}

func TestPlanner_MixedResources(t *testing.T) {
	resources := []youtubeurl.YouTubeInput{
		{InputType: youtubeurl.InputTypeVideo, ID: "vid123"},
		{InputType: youtubeurl.InputTypePlaylist, ID: "PL123"},
	}
	fp, err := fields.NewPlanner([]string{fields.FieldTitle})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan := Plan(resources, fp)

	if !plan.Resources.MixedResources {
		t.Errorf("expected MixedResources to be true")
	}
	if !plan.NeedsPlaylistItems() {
		t.Errorf("expected NeedsPlaylistItems to be true")
	}
	if !plan.NeedsMetadata() {
		t.Errorf("expected NeedsMetadata to be true because of video resource")
	}
}
