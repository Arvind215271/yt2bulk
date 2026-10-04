package preflight

import (
	"github.com/labstack/echo/v5"
)

func RegisterRoutes(g *echo.Group, h *Handler) {
	g.POST("", h.Check)
}
