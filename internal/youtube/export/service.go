package export

import (
	"sync"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

// Service is the application layer that orchestrates
// building export data + choosing exporter.
type Service struct {
	exporters map[Format]Exporter
	mu        sync.RWMutex
}

// NewService creates the export service with registered exporters.
func NewService() *Service {
	s := &Service{
		exporters: make(map[Format]Exporter),
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

	return exporter.Export(data)
}
