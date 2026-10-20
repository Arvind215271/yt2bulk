package preflight

import (
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/api"
)

type PreflightErrors struct{}

var Err = PreflightErrors{}

func (PreflightErrors) BadRequest(msg string) *api.AppError {
	return api.NewError("BAD_REQUEST", msg, http.StatusBadRequest)
}

func (PreflightErrors) InternalError(err error) *api.AppError {
	return api.NewError("INTERNAL_ERROR", "Something went wrong", http.StatusInternalServerError).Wrap(err)
}
