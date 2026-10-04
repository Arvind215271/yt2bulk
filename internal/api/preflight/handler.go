package preflight

import (
	"net/http"

	youtubeurl "github.com/Arvind215271/askito/internal/youtube/input"
	"github.com/Arvind215271/askito/internal/youtube/preflight"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	preflightService preflight.Service
}

func NewHandler(preflightService preflight.Service) *Handler {
	return &Handler{
		preflightService: preflightService,
	}
}

func (h *Handler) Check(c *echo.Context) error {
	var req PreflightRequest
	if err := (*c).Bind(&req); err != nil {
		return Err.BadRequest("invalid request body").Wrap(err)
	}

	if len(req.Inputs) == 0 {
		return Err.BadRequest("inputs required")
	}

	inputStr := req.Inputs[0]
	parsed, err := youtubeurl.Parse(inputStr)
	if err != nil {
		return Err.BadRequest("invalid URL").Wrap(err)
	}

	preflightReq := preflight.Request{
		Input: parsed,
	}

	ctx := (*c).Request().Context()
	result, err := h.preflightService.Check(ctx, preflightReq)
	if err != nil {
		return Err.InternalError(err)
	}

	return (*c).JSON(http.StatusOK, PreflightResponse{
		Valid:              result.Valid,
		UniqueVideoCount:   result.UniqueVideoCount,
		TotalResourceCount: result.TotalResourceCount,
		EstimatedWork:      result.EstimatedWork,
	})
}
