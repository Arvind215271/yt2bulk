package export

import (
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

type Format string

const (
	FormatJSON     Format = "json"
	FormatCSV      Format = "csv"
	FormatMarkdown Format = "markdown"
	FormatExcel    Format = "excel"
	FormatYAML     Format = "yaml"
	FormatXML      Format = "xml"
)

type ExportData map[string]any

type PlaylistExportRequest struct {
	PlaylistID  string          `json:"playlist_id"`
	VideoFields *fields.Planner `json:"-"`
	Format      Format          `json:"format"`
}

type VideoExportRequest struct {
	VideoID string          `json:"video_id"`
	Fields  *fields.Planner `json:"-"`
	Format  Format          `json:"format"`
}

type BatchVideoExportRequest struct {
	VideoIDs    []string        `json:"video_ids"`
	VideoFields *fields.Planner `json:"-"`
	Format      Format          `json:"format"`
}

type ResourceExportRequest struct {
	ResourceIDs []string        `json:"resource_ids"`
	Fields      *fields.Planner `json:"-"`
	Format      Format          `json:"format"`
}

type ExportResponse struct {
	SchemaVersion string              `json:"schema_version"`
	Format        Format              `json:"format"`
	Data          ExportData          `json:"data"`
	Statistics    *stats.ProcessStats `json:"statistics,omitempty"`
}
