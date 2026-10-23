package preflight

import (
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/logger"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/preflight"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	preflightService preflight.Service
	logger           *logger.Logger
}

func NewHandler(preflightService preflight.Service, logger *logger.Logger) *Handler {
	return &Handler{
		preflightService: preflightService,
		logger:           logger,
	}
}

func (h *Handler) Check(c *echo.Context) error {
	var req PreflightRequest
	if err := (*c).Bind(&req); err != nil {
		h.logger.Error("failed to bind preflight request", "err", err)
		return Err.BadRequest("invalid request body").Wrap(err)
	}

	h.logger.Info("received preflight request", "inputs_count", len(req.Inputs))

	if len(req.Inputs) == 0 {
		h.logger.Error("no inputs provided for preflight")
		return Err.BadRequest("inputs required")
	}

	// NOTE: The preflight service now internally enforces the limit
	// during resource resolution. If the limit is exceeded,
	// it will return an error that we can map or handle appropriately.

	var parsedInputs []*youtubeurl.YouTubeInput
	for _, inputStr := range req.Inputs {
		parsed, err := youtubeurl.Parse(inputStr)
		if err != nil {
			return Err.BadRequest("invalid URL").Wrap(err)
		}
		parsedInputs = append(parsedInputs, parsed)
	}

	preflightReq := preflight.Request{
		Inputs: parsedInputs,
	}

	ctx := (*c).Request().Context()
	result, err := h.preflightService.Check(ctx, preflightReq)
	if err != nil {
		h.logger.Error("preflight check execution failed", "err", err)
		return Err.BadRequest(err.Error()).Wrap(err)
	}
	if !result.Valid {
		h.logger.Error("preflight check result invalid", "err", result.Error)
		return Err.BadRequest(result.Error.Error()).Wrap(result.Error)
	}

	h.logger.Info("preflight check completed successfully", "unique_videos", result.UniqueVideoCount, "total_resources", result.TotalResourceCount)

	return (*c).JSON(http.StatusOK, PreflightResponse{
		Valid:              result.Valid,
		UniqueVideoCount:   result.UniqueVideoCount,
		TotalResourceCount: result.TotalResourceCount,
		EstimatedWork:      result.EstimatedWork,
		DirectVideos: PreflightStatsResponse{
			Type:             result.Stats.DirectVideos.Type,
			ItemCount:        result.Stats.DirectVideos.ItemCount,
			MetadataCached:   result.Stats.DirectVideos.MetadataCached,
			MetadataUpstream: result.Stats.DirectVideos.MetadataUpstream,
			SubtitleCached:   result.Stats.DirectVideos.SubtitleCached,
			SubtitleUpstream: result.Stats.DirectVideos.SubtitleUpstream,
		},
		Playlists: PlaylistPreflightStatsResponse{
			Type:           result.Stats.Playlists.Type,
			TotalPlaylists: result.Stats.Playlists.TotalPlaylists,
			TotalVideos:    result.Stats.Playlists.TotalVideos,
			Items: PreflightStatsResponse{
				Type:             result.Stats.Playlists.Items.Type,
				ItemCount:        result.Stats.Playlists.Items.ItemCount,
				MetadataCached:   result.Stats.Playlists.Items.MetadataCached,
				MetadataUpstream: result.Stats.Playlists.Items.MetadataUpstream,
				SubtitleCached:   result.Stats.Playlists.Items.SubtitleCached,
				SubtitleUpstream: result.Stats.Playlists.Items.SubtitleUpstream,
			},
		},
		Combined: PreflightStatsResponse{
			Type:             result.Stats.Combined.Type,
			ItemCount:        result.Stats.Combined.ItemCount,
			MetadataCached:   result.Stats.Combined.MetadataCached,
			MetadataUpstream: result.Stats.Combined.MetadataUpstream,
			SubtitleCached:   result.Stats.Combined.SubtitleCached,
			SubtitleUpstream: result.Stats.Combined.SubtitleUpstream,
		},
	})
}
