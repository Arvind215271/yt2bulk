package playlist

import (
	"net/http"

	"github.com/Arvind215271/askito/internal/api"
	youtubeurl "github.com/Arvind215271/askito/internal/youtube/input"
	"github.com/Arvind215271/askito/internal/youtube/metadata"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	youtubeService *metadata.Service
}

func NewHandler(youtubeService *metadata.Service) *Handler {
	return &Handler{
		youtubeService: youtubeService,
	}
}

func (h *Handler) ExpandPlaylistVideos(c *echo.Context) error {
	var req PlaylistVideosRequest
	if err := (*c).Bind(&req); err != nil {
		return Err.BadRequest("Invalid request").Wrap(err)
	}

	if err := api.Validate(req); err != nil {
		return err
	}

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
	items, err := h.youtubeService.GetPlaylistItems(ctx, parsed.ID, 50, providerType)
	if err != nil {
		return Err.FetchFailed(err)
	}

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
