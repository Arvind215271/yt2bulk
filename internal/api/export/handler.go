package export

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/Arvind215271/askito/internal/api"
	subtitleapi "github.com/Arvind215271/askito/internal/api/subtitle"
	"github.com/Arvind215271/askito/internal/youtube/export"
	"github.com/Arvind215271/askito/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/askito/internal/youtube/input"
	"github.com/Arvind215271/askito/internal/youtube/pipeline"
	"github.com/Arvind215271/askito/internal/youtube/planner"
	"github.com/Arvind215271/askito/internal/youtube/resource"
)

type Handler struct {
	resourceService *resource.Service
	exportService   *export.Service
}

func NewHandler(
	resourceService *resource.Service,
	exportService *export.Service,
) *Handler {
	return &Handler{
		resourceService: resourceService,
		exportService:   exportService,
	}
}

func parseFormat(s string) (export.Format, error) {
	if s == "" {
		return export.FormatJSON, nil
	}

	format := export.Format(s)

	switch format {
	case export.FormatJSON, export.FormatCSV, export.FormatMarkdown, export.FormatExcel, export.FormatYAML, export.FormatXML:
		return format, nil

	default:
		return "", ErrInvalidFormat
	}
}

func (h *Handler) Export(c *echo.Context) error {
	var req ExportRequest

	if err := c.Bind(&req); err != nil {
		return api.Err.Common.
			BadRequest("invalid request body").
			Wrap(err)
	}

	if len(req.Inputs) == 0 {
		return ErrInputRequired
	}

	exportFormat, err := parseFormat(req.Format)
	if err != nil {
		return err
	}

	fieldPlanner, err := fields.NewPlanner(req.Fields)
	if err != nil {
		return err
	}

	preferences, err := subtitleapi.BuildPreferences(req.Preferences)
	if err != nil {
		return api.Err.Common.BadRequest("invalid subtitle preferences").Wrap(err)
	}

	inputs := make([]youtubeurl.YouTubeInput, len(req.Inputs))
	for i, in := range req.Inputs {
		parsed, err := youtubeurl.Parse(in)
		if err == nil && parsed != nil {
			inputs[i] = *parsed
		} else {
			inputs[i] = youtubeurl.YouTubeInput{InputType: youtubeurl.InputTypeVideo, ID: in}
		}
	}

	executionPlan := planner.Build(inputs, fieldPlanner)

	pipelineReq := &pipeline.Request{
		FieldPlanner:  fieldPlanner,
		ExecutionPlan: executionPlan,
		Subtitle:      req.Subtitle,
		Preferences:   preferences,
		Format:        "json3",
		Transcript:    req.Transcript,
		Signal:        req.Signal,
	}

	ctx := c.Request().Context()

	resources, stats := h.resourceService.ProcessResourcesWithStats(ctx, inputs, pipelineReq)

	data, err := h.exportService.ExportBatchResource(resources, exportFormat, fieldPlanner, &stats)
	if err != nil {
		return err
	}

	contentType := "application/json"
	switch exportFormat {
	case export.FormatCSV:
		contentType = "text/csv"
	case export.FormatMarkdown:
		contentType = "text/markdown"
	case export.FormatYAML:
		contentType = "application/yaml"
	case export.FormatXML:
		contentType = "application/xml"
	case export.FormatExcel:
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	return c.Blob(
		http.StatusOK,
		contentType,
		data,
	)
}
