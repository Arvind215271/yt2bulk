package channel

import (
	"github.com/labstack/echo/v5"
)

func RegisterChannelRoutes(g *echo.Group, h *Handler) {
	g.POST("/metadata", h.GetChannelMetadata)
	g.POST("/videos", h.ListChannelVideos)
}
