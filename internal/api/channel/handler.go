package channel

import (
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/api"
	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube"
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

func (h *Handler) GetChannelMetadata(c *echo.Context) error {
	var req ChannelRequest
	if err := (*c).Bind(&req); err != nil {
		h.logger.Error("failed to bind channel request", "err", err)
		return Err.BadRequest("Invalid request").Wrap(err)
	}

	h.logger.Info("received channel metadata request", "url", req.URL)

	if err := api.Validate(req); err != nil {
		h.logger.Error("invalid channel request", "err", err)
		return err
	}

	parsed, err := youtubeurl.Parse(req.URL)
	if err != nil {
		return Err.InvalidURL().Wrap(err)
	}

	if parsed.InputType != youtubeurl.InputTypeChannel {
		return Err.NotAChannel()
	}

	providerType := metadata.ProviderType(req.Provider)
	if providerType == "" {
		providerType = metadata.ProviderYTDLP
	}

	ctx := (*c).Request().Context()
	playlist, err := h.youtubeService.GetChannel(ctx, parsed.NormalizedURL, providerType)
	if err != nil {
		h.logger.Error("failed to fetch channel metadata", "err", err)
		return Err.FetchFailed(err)
	}

	return (*c).JSON(http.StatusOK, ChannelMetadataResponse{
		ChannelID:   playlist.ID,
		Title:       playlist.Title,
		Description: playlist.Description,
		URL:         parsed.NormalizedURL,
	})
}

func (h *Handler) ListChannelVideos(c *echo.Context) error {
	var req ChannelRequest
	if err := (*c).Bind(&req); err != nil {
		h.logger.Error("failed to bind channel request", "err", err)
		return Err.BadRequest("Invalid request").Wrap(err)
	}

	h.logger.Info("received channel list request", "url", req.URL)

	if err := api.Validate(req); err != nil {
		h.logger.Error("invalid channel request", "err", err)
		return err
	}

	parsed, err := youtubeurl.Parse(req.URL)
	if err != nil {
		return Err.InvalidURL().Wrap(err)
	}

	if parsed.InputType != youtubeurl.InputTypeChannel {
		return Err.NotAChannel()
	}

	providerType := metadata.ProviderType(req.Provider)
	if providerType == "" {
		providerType = metadata.ProviderYTDLP
	}

	ctx := (*c).Request().Context()

	limit := req.Limit
	if req.LimitMode == "all" {
		limit = 10000 // Large enough
	} else if limit <= 0 {
		limit = 50
	}

	// For channels, GetChannelItems on the normalized URL gets the channel videos
	items, err := h.youtubeService.GetChannelItems(ctx, parsed.NormalizedURL, limit, providerType)
	if err != nil {
		return Err.FetchFailed(err)
	}

	// Apply Range filtering if requested
	if req.LimitMode == "range" && req.RangeStart > 0 && req.RangeEnd >= req.RangeStart {
		var filtered []youtube.PlaylistItem
		for _, item := range items {
			if item.Position >= req.RangeStart && item.Position <= req.RangeEnd {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	h.logger.Info("successfully fetched channel items", "count", len(items))

	outputMode := req.Output
	if outputMode == "" {
		outputMode = OutputTypeBoth
	}

	videos := make([]VideoEntry, 0, len(items))
	for _, item := range items {
		entry := VideoEntry{
			Position: item.Position,
			Title:    item.Title,
			Duration: int(item.Duration.Seconds()),
		}

		if outputMode == OutputTypeID || outputMode == OutputTypeBoth {
			entry.ID = item.VideoID
		}
		if outputMode == OutputTypeURL || outputMode == OutputTypeBoth {
			entry.URL = "https://www.youtube.com/watch?v=" + item.VideoID
		}

		videos = append(videos, entry)
	}

	return (*c).JSON(http.StatusOK, ChannelVideosResponse{
		ChannelID: parsed.ID,
		Total:     len(videos),
		Videos:    videos,
	})
}
