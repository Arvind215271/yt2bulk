package preflight

import (
	"net/http"

	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/preflight"
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
		return Err.InternalError(err)
	}

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
