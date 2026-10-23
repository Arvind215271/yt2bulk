package transcript

import (
	"fmt"

	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
)

type Service struct {
	logger *logger.Logger
}

func NewService(logger *logger.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) Parse(result *subtitle.SubtitleResult) (*Transcript, error) {
	if result == nil {
		if s.logger != nil {
			s.logger.Error("subtitle result is nil")
		}
		return nil, fmt.Errorf("subtitle result is nil")
	}

	if s.logger != nil {
		s.logger.Debug("parsing transcript from subtitle result", "format", result.Format, "language", result.Language)
	}

	switch result.Format {

	case "json3":
		segments, err := ParseJSON3ToSegments(result.Content)
		if err != nil {
			if s.logger != nil {
				s.logger.Error("failed to parse json3 transcript", "error", err)
			}
			return nil, fmt.Errorf("failed to parse json3 transcript: %w", err)
		}

		if s.logger != nil {
			s.logger.Info("successfully parsed json3 transcript", "segments_count", len(segments), "language", result.Language)
		}

		return &Transcript{
			Language: result.Language,
			Segments: segments,
		}, nil

	default:
		if s.logger != nil {
			s.logger.Error("unsupported subtitle format for transcript parsing", "format", result.Format)
		}
		return nil, fmt.Errorf("unsupported subtitle format: %s", result.Format)
	}
}

func (s *Service) Process(t *Transcript, req *ProcessingRequest) (string, error) {
	if req == nil {
		if s.logger != nil {
			s.logger.Error("transcript processing request is nil")
		}
		return "", fmt.Errorf("transcript processing request is nil")
	}

	if s.logger != nil {
		s.logger.Debug("processing transcript", "window_size", req.WindowSize, "output_format", req.Output)
	}

	processed := t
	if req.WindowSize > 0 {
		processed = t.GroupByDuration(req.WindowSize)
	}

	switch req.Output {

	case "plain-text":
		res := processed.ToPlainText()
		if s.logger != nil {
			s.logger.Info("processed transcript to plain-text successfully", "length", len(res))
		}
		return res, nil
	case "timeline-text":
		res := processed.ToTimelineText()
		if s.logger != nil {
			s.logger.Info("processed transcript to timeline-text successfully", "length", len(res))
		}
		return res, nil
	default:
		if s.logger != nil {
			s.logger.Error("unsupported output format for transcript processing", "output", req.Output)
		}
		return "", fmt.Errorf("unsupported output format: %s", req.Output)
	}
}
