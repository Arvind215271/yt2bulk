package transcript

import (
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/api"
	subtitleapi "github.com/Arvind215271/yt2bulk/internal/api/subtitle"
	"github.com/Arvind215271/yt2bulk/internal/youtube/fields"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/pipeline"
	"github.com/Arvind215271/yt2bulk/internal/youtube/planner"
	"github.com/Arvind215271/yt2bulk/internal/youtube/resource"
	// "github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	resourceService *resource.Service
}

func NewHandler(resourceService *resource.Service) *Handler {
	return &Handler{
		resourceService: resourceService,
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

	fieldPlanner, err := fields.NewPlanner([]string{"id", "transcript"})
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
		Subtitle: nil,
		Preferences: preferences,
		Format:      "json3",
		Transcript:  nil,
	}

	ctx := c.Request().Context()
	resources := h.resourceService.ProcessResources(ctx, inputs, pipelineReq)

	return c.JSON(http.StatusOK, resources)
}
