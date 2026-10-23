package description

import (
	"context"

	"github.com/Arvind215271/yt2bulk/internal/logger"
)

type Service struct {
	logger *logger.Logger
}

func NewService(logger *logger.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) GetDescription(ctx context.Context, description string) (Metadata, error) {
	s.logger.With("description_len", len(description)).Info("Processing description")
	return ProcessDescription(description), nil
}
