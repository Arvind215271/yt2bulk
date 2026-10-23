package export

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/Arvind215271/yt2bulk/internal/api"
	subtitleapi "github.com/Arvind215271/yt2bulk/internal/api/subtitle"
	"github.com/Arvind215271/yt2bulk/internal/config"
	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/export"
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/pipeline"
	"github.com/Arvind215271/yt2bulk/internal/youtube/planner"
	"github.com/Arvind215271/yt2bulk/internal/youtube/preflight"
	"github.com/Arvind215271/yt2bulk/internal/youtube/resource"
)

type Handler struct {
	resourceService  *resource.Service
	exportService    *export.Service
	preflightService preflight.Service
	cfg              config.Config
	logger           *logger.Logger
}

func NewHandler(
	resourceService *resource.Service,
	exportService *export.Service,
	preflightService preflight.Service,
	cfg config.Config,
	logger *logger.Logger,
) *Handler {
	return &Handler{
		resourceService:  resourceService,
		exportService:    exportService,
		preflightService: preflightService,
		cfg:              cfg,
		logger:           logger,
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
		h.logger.Error("failed to bind export request", "err", err)
		return api.Err.Common.
			BadRequest("invalid request body").
			Wrap(err)
	}

	h.logger.Info("received export request", "inputs_count", len(req.Inputs), "format", req.Format)

	if len(req.Inputs) == 0 {
		h.logger.Error("no inputs provided for export")
		return ErrInputRequired
	}

	// Run preflight check
	// Use ParseMany to handle multiple inputs robustly
	// Join inputs into a single string for ParseMany
	joinedInputs := strings.Join(req.Inputs, " ")
	parsedResults := youtubeurl.ParseMany(joinedInputs)

	parsedInputsForPreflight := make([]*youtubeurl.YouTubeInput, 0, len(parsedResults))
	for _, res := range parsedResults {
		if res.Input != nil {
			parsedInputsForPreflight = append(parsedInputsForPreflight, res.Input)
		} else {
			// Fallback if parsing fails for an item
			parsedInputsForPreflight = append(parsedInputsForPreflight, &youtubeurl.YouTubeInput{InputType: youtubeurl.InputTypeVideo, ID: res.Raw})
		}
	}
	preflightReq := preflight.Request{Inputs: parsedInputsForPreflight}
	ctx := c.Request().Context()
	res, err := h.preflightService.Check(ctx, preflightReq)
	if err != nil {
		h.logger.Error("preflight check failed in export", "err", err)
		return youtube.Err.Export.LimitExceeded(res.TotalResourceCount, h.cfg.MaxExportLimit).Wrap(err)
	}
	if !res.Valid {
		h.logger.Error("preflight check invalid in export", "err", res.Error)
		return youtube.Err.Export.LimitExceeded(res.TotalResourceCount, h.cfg.MaxExportLimit).Wrap(res.Error)
	}

	exportFormat, err := parseFormat(req.Format)
	if err != nil {
		return err
	}

	fieldPlanner, err := fields.NewPlanner(req.Fields, h.logger)
	if err != nil {
		return err
	}

	preferences, err := subtitleapi.BuildPreferences(req.Preferences)
	if err != nil {
		return api.Err.Common.BadRequest("invalid subtitle preferences").Wrap(err)
	}

	// Reuse parsed inputs for execution
	inputs := make([]youtubeurl.YouTubeInput, 0, len(parsedInputsForPreflight))
	for _, p := range parsedInputsForPreflight {
		if p != nil {
			inputs = append(inputs, *p)
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

	ctx = c.Request().Context()

	h.logger.Info("processing resources with stats for export", "count", len(inputs))
	resources, stats := h.resourceService.ProcessResourcesWithStats(ctx, inputs, pipelineReq)

	data, err := h.exportService.ExportBatchResource(resources, exportFormat, fieldPlanner, &stats)
	if err != nil {
		h.logger.Error("failed to export batch resource", "err", err)
		return err
	}

	h.logger.Info("successfully completed export request", "data_size", len(data), "format", exportFormat)

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
