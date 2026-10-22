package transcript

import (
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/api"
	subtitleapi "github.com/Arvind215271/yt2bulk/internal/api/subtitle"
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/pipeline"
	"github.com/Arvind215271/yt2bulk/internal/youtube/planner"
	"github.com/Arvind215271/yt2bulk/internal/youtube/preflight"
	"github.com/Arvind215271/yt2bulk/internal/youtube/resource"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	resourceService  *resource.Service
	preflightService preflight.Service
}

func NewHandler(resourceService *resource.Service, preflightService preflight.Service) *Handler {
	return &Handler{
		resourceService:  resourceService,
		preflightService: preflightService,
	}
}

func (h *Handler) GetTranscript(c *echo.Context) error {
	var req TranscriptRequest
	if err := c.Bind(&req); err != nil {
		return api.Err.Common.BadRequest("invalid request body").Wrap(err)
	}

	if len(req.Inputs) == 0 {
		return api.Err.Common.BadRequest("inputs required")
	}

	parsedInputsForPreflight := make([]*youtubeurl.YouTubeInput, len(req.Inputs))
	for i, in := range req.Inputs {
		parsed, err := youtubeurl.Parse(in)
		if err == nil && parsed != nil {
			parsedInputsForPreflight[i] = parsed
		} else {
			parsedInputsForPreflight[i] = &youtubeurl.YouTubeInput{InputType: youtubeurl.InputTypeVideo, ID: in}
		}
	}
	preflightReq := preflight.Request{Inputs: parsedInputsForPreflight}
	ctx := c.Request().Context()
	res, err := h.preflightService.Check(ctx, preflightReq)
	if err != nil {
		return api.Err.Common.BadRequest(err.Error()).Wrap(err)
	}
	if !res.Valid {
		return api.Err.Common.BadRequest(res.Error.Error()).Wrap(res.Error)
	}

	fieldPlanner, err := fields.NewPlanner([]string{"id", "transcript_text", "transcript_signal"})
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
		Subtitle:      nil,
		Preferences:   preferences,
		Format:        "json3",
		Transcript:    nil,
		Signal:        req.Signal,
	}

	resources := h.resourceService.ProcessResources(ctx, inputs, pipelineReq)

	return c.JSON(http.StatusOK, resources)
}
