package pipeline

import (
	"testing"

	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/planner"
	"github.com/Arvind215271/yt2bulk/internal/youtube/signal"
	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
	"github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
)

func TestRequest_Combinations(t *testing.T) {
	fp, err := fields.NewPlanner([]string{
		fields.FieldTranscriptText,
		fields.FieldTranscriptSignal,
		fields.FieldDescriptionChapters,
		fields.FieldTitle,
	})
	if err != nil {
		t.Fatalf("unexpected error creating planner: %v", err)
	}
	inputs := []youtubeurl.YouTubeInput{{InputType: youtubeurl.InputTypeVideo, ID: "dQw4w9WgXcQ"}}
	ep := planner.Build(inputs, fp)

	r1 := Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
		Transcript:    &transcript.ProcessingRequest{Output: "plain-text"},
	}

	if r1.FieldPlanner == nil {
		t.Errorf("expected FieldPlanner not to be nil")
	}
	if r1.ExecutionPlan == nil {
		t.Errorf("expected ExecutionPlan not to be nil")
	}

	r2 := Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
		Subtitle:      &subtitle.DownloadRequest{Type: "auto", Language: "en", Format: "vtt"},
	}
	if r2.FieldPlanner == nil {
		t.Errorf("expected FieldPlanner not to be nil")
	}

	r3 := Request{
		FieldPlanner:  fp,
		ExecutionPlan: ep,
		Signal:        &signal.SignalRequest{Analysis: "words", UseHeavy: true},
	}
	if r3.FieldPlanner == nil {
		t.Errorf("expected FieldPlanner not to be nil")
	}
}
