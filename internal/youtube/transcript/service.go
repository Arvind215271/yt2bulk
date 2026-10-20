package transcript

import (
	"fmt"

	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Parse(result *subtitle.SubtitleResult) (*Transcript, error) {
	if result == nil {
		return nil, fmt.Errorf("subtitle result is nil")
	}

	switch result.Format {

	case "json3":
		segments, err := ParseJSON3ToSegments(result.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to parse json3 transcript: %w", err)
		}

		return &Transcript{
			Language: result.Language,
			Segments: segments,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported subtitle format: %s", result.Format)
	}
}

func (s *Service) Process(t *Transcript, req *ProcessingRequest) (string, error) {
	if req == nil {
		return "", fmt.Errorf("transcript processing request is nil")
	}

	processed := t
	if req.WindowSize > 0 {
		processed = t.GroupByDuration(req.WindowSize)
	}

	switch req.Output {
	
	case "plain-text":
		return processed.ToPlainText(), nil
	case "timeline-text":
		return processed.ToTimelineText(), nil
	default:
		return "", fmt.Errorf("unsupported output format: %s", req.Output)
	}
}
