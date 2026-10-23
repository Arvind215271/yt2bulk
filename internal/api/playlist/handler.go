package playlist

import (
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/api"
	"github.com/Arvind215271/yt2bulk/internal/logger"
	youtubeurl "github.com/Arvind215271/yt2bulk/internal/youtube/input"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
	"github.com/Arvind215271/yt2bulk/internal/youtube/preflight"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	youtubeService   *metadata.Service
	preflightService preflight.Service
	logger           *logger.Logger
}

func NewHandler(
	youtubeService *metadata.Service,
	preflightService preflight.Service,
	logger *logger.Logger,
) *Handler {
	return &Handler{
		youtubeService:   youtubeService,
		preflightService: preflightService,
		logger:           logger,
	}
}

func (h *Handler) ExpandPlaylistVideos(c *echo.Context) error {
	var req PlaylistVideosRequest
	if err := (*c).Bind(&req); err != nil {
		h.logger.Error("failed to bind playlist request", "err", err)
		return Err.BadRequest("Invalid request").Wrap(err)
	}

	h.logger.Info("received playlist request", "url", req.URL)

	if err := api.Validate(req); err != nil {
		h.logger.Error("invalid playlist request", "err", err)
		return err
	}

	// Add Preflight Check
	parsed, err := youtubeurl.Parse(req.URL)
	if err != nil {
		return Err.InvalidURL().Wrap(err)
	}

	if parsed.InputType != youtubeurl.InputTypePlaylist {
		return Err.NotAPlaylist()
	}

	providerType := metadata.ProviderType(req.Provider)
	if providerType == "" {
		providerType = metadata.ProviderYTDLP
	}

	if providerType != metadata.ProviderAPI && providerType != metadata.ProviderYTDLP {
		return Err.InvalidProvider()
	}

	ctx := (*c).Request().Context()

	// Preflight check
	preflightReq := preflight.Request{Inputs: []*youtubeurl.YouTubeInput{parsed}}
	res, err := h.preflightService.Check(ctx, preflightReq)
	if err != nil {
		h.logger.Error("preflight check failed in playlist", "err", err)
		return err // Return preflight error directly
	}
	if !res.Valid {
		h.logger.Error("invalid preflight results in playlist", "err", res.Error)
		return res.Error // Return preflight validation error
	}

	items, err := h.youtubeService.GetPlaylistItems(ctx, parsed.ID, 50, providerType)
	if err != nil {
		h.logger.Error("failed to fetch playlist items", "err", err)
		return Err.FetchFailed(err)
	}

	h.logger.Info("successfully fetched playlist items", "count", len(items))

	outputMode := req.Output
	if outputMode == "" {
		outputMode = OutputTypeBoth
	}

	videos := make([]VideoEntry, 0, len(items))
	for _, item := range items {
		entry := VideoEntry{
			Position: item.Position,
		}
		// if !item.AddedAt.IsZero() {
		// 	entry.AddedAt = item.AddedAt.Format("2006-01-02T15:04:05Z07:00")
		// }

		if outputMode == OutputTypeID || outputMode == OutputTypeBoth {
			entry.ID = item.VideoID
		}
		if outputMode == OutputTypeURL || outputMode == OutputTypeBoth {
			entry.URL = "https://www.youtube.com/watch?v=" + item.VideoID
		}

		videos = append(videos, entry)
	}

	return (*c).JSON(http.StatusOK, PlaylistVideosResponse{
		PlaylistID: parsed.ID,
		Total:      len(videos),
		Videos:     videos,
	})
}
