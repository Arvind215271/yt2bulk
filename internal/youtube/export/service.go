package export

import (
	"sync"

	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

// Service is the application layer that orchestrates
// building export data + choosing exporter.
type Service struct {
	exporters map[Format]Exporter
	mu        sync.RWMutex
	logger    *logger.Logger
}

// NewService creates the export service with registered exporters.
func NewService(logger *logger.Logger) *Service {
	s := &Service{
		exporters: make(map[Format]Exporter),
		logger:    logger,
	}
	s.RegisterExporter(FormatJSON, &JSONExporter{Pretty: true})
	s.RegisterExporter(FormatCSV, &CSVExporter{})
	s.RegisterExporter(FormatMarkdown, &MarkdownExporter{})
	s.RegisterExporter(FormatExcel, &ExcelExporter{})
	s.RegisterExporter(FormatYAML, &YAMLExporter{})
	s.RegisterExporter(FormatXML, &XMLExporter{})
	return s
}

// RegisterExporter allows plugging in new formats
func (s *Service) RegisterExporter(format Format, exporter Exporter) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.exporters[format] = exporter
}

// ExportPlaylist orchestrates playlist export.
func (s *Service) ExportPlaylist(
	playlist youtube.Playlist,
	req PlaylistExportRequest,
) ([]byte, error) {

	data, err := BuildPlaylist(playlist, req.VideoFields)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to build playlist for export", "error", err, "playlist_id", playlist.ID)
		}
		return nil, err
	}

	return s.exportData(req.Format, data)
}

// ExportVideo orchestrates video export.
func (s *Service) ExportVideo(
	video youtube.Video,
	req VideoExportRequest,
) ([]byte, error) {

	data, err := BuildVideo(video, req.Fields)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to build video for export", "error", err, "video_id", video.ID)
		}
		return nil, err
	}

	return s.exportData(req.Format, data)
}

// ExportBatchVideos orchestrates batch video export.
func (s *Service) ExportBatchVideos(
	videos []youtube.Video,
	req BatchVideoExportRequest,
) ([]byte, error) {

	data, err := BuildBatchVideo(videos, req.VideoFields)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to build batch video for export", "error", err, "videos_count", len(videos))
		}
		return nil, err
	}

	return s.exportData(req.Format, data)
}

// ExportResource orchestrates a single youtube.Resource export.
func (s *Service) ExportResource(
	resource youtube.Resource,
	format Format,
	planner *fields.Planner,
) ([]byte, error) {

	data, err := BuildResource(resource, planner)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to build resource for export", "error", err, "resource_id", resource.ID)
		}
		return nil, err
	}

	return s.exportData(format, data)
}

// ExportBatchResource orchestrates multiple youtube.Resource exports.
func (s *Service) ExportBatchResource(
	resources []youtube.Resource,
	format Format,
	planner *fields.Planner,
	stats *stats.ProcessStats,
) ([]byte, error) {

	data, err := BuildBatchResource(resources, planner)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to build batch resource for export", "error", err, "resources_count", len(resources))
		}
		return nil, err
	}

	// Incorporate statistics into ExportData
	if stats != nil {
		data["statistics"] = stats
	}

	return s.exportData(format, data)
}

// internal shared logic
func (s *Service) exportData(format Format, data ExportData) ([]byte, error) {
	s.mu.RLock()
	exporter, ok := s.exporters[format]
	s.mu.RUnlock()

	if !ok {
		// Fallback to JSON if format is unspecified or unrecognized
		s.mu.RLock()
		exporter, ok = s.exporters[FormatJSON]
		s.mu.RUnlock()
		if !ok {
			return nil, youtube.Err.Export.InvalidFormat()
		}
	}

	res, err := exporter.Export(data)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("failed to export data", "error", err, "format", format)
		}
		return nil, err
	}

	if s.logger != nil {
		s.logger.Info("data exported successfully", "format", format, "size", len(res))
	}
	return res, nil
}
