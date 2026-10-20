package transcript

import (
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/api"
	subtitleapi "github.com/Arvind215271/yt2bulk/internal/api/subtitle"
)

type TranscriptRequest struct {
	Inputs      []string                                `json:"inputs" validate:"required,min=1"`
	Preferences []subtitleapi.SubtitlePreferenceRequest `json:"preferences,omitempty"`
}

type TranscriptErrors struct{}

var Err = TranscriptErrors{}

func (TranscriptErrors) BadRequest(msg string) *api.AppError {
	return api.NewError("BAD_REQUEST", msg, http.StatusBadRequest)
}

func (TranscriptErrors) InvalidURL() *api.AppError {
	return api.NewError("INVALID_URL", "Invalid URL", http.StatusBadRequest)
}

func (TranscriptErrors) FetchFailed(err error) *api.AppError {
	return api.NewError("FETCH_FAILED", "Failed to fetch video", http.StatusInternalServerError).Wrap(err)
}

func (TranscriptErrors) InternalError(err error) *api.AppError {
	return api.NewError("INTERNAL_ERROR", "Something went wrong", http.StatusInternalServerError).Wrap(err)
}
