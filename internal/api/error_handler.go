// ./internal/api/error_handler.go

package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/Arvind215271/yt2bulk/internal/logger"
)

type ErrorHandler struct {
	Log *logger.Logger
}

func NewErrorHandler(log *logger.Logger) *ErrorHandler {
	return &ErrorHandler{
		Log: log,
	}
}

func (h *ErrorHandler) Handle(c *echo.Context, err error) {

	// response already sent
	if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if resp.Committed {
			return
		}
	}

	// application error
	var appErr *AppError

	if errors.As(err, &appErr) {
		req := c.Request()
		var path, method string
		if req != nil {
			path = req.URL.Path
			method = req.Method
		}

		h.Log.Error(
			"application error",

			"code", appErr.Code,
			"status", appErr.Status,

			"path", path,
			"method", method,

			"error", appErr.Err,
		)

		_ = c.JSON(appErr.Status, ErrorResponse(appErr))
		return
	}

	// route not found
	var sc echo.HTTPStatusCoder

	if errors.As(err, &sc) {

		if sc.StatusCode() == http.StatusNotFound {

			appErr := Err.Common.RouteNotFound()

			h.Log.Warn(
				"route not found",
				"path", c.Request().URL.Path,
				"method", c.Request().Method,
			)

			_ = c.JSON(appErr.Status, ErrorResponse(appErr))
			return
		}
	}

	// fallback error
	appErr = Err.Common.Internal()

	req := c.Request()
	var path, method string
	if req != nil {
		path = req.URL.Path
		method = req.Method
	}

	h.Log.Error(
		"unexpected error",
		"path", path,
		"method", method,
		"error", err,
	)

	_ = c.JSON(appErr.Status, ErrorResponse(appErr))
}