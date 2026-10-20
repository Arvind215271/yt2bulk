package subtitle

import (
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/api"
)

type SubtitleErrors struct{}

var Err = SubtitleErrors{}

func (SubtitleErrors) BadRequest(msg string) *api.AppError {
	return api.NewError("BAD_REQUEST", msg, http.StatusBadRequest)
}

func (SubtitleErrors) InvalidURL() *api.AppError {
	return api.NewError("INVALID_URL", "Invalid URL", http.StatusBadRequest)
}

func (SubtitleErrors) FetchFailed(err error) *api.AppError {
	return api.NewError("FETCH_FAILED", "Failed to fetch video", http.StatusInternalServerError).Wrap(err)
}

func (SubtitleErrors) InvalidPreference(err error) *api.AppError {
	return api.NewError("INVALID_PREFERENCE", "Invalid subtitle preference", http.StatusBadRequest).Wrap(err)
}

func (SubtitleErrors) SubtitleNotFound(err error) *api.AppError {
	return api.NewError("SUBTITLE_NOT_FOUND", "No matching subtitle found", http.StatusNotFound).Wrap(err)
}

func (SubtitleErrors) MultiplePlaylistsNotAllowed() *api.AppError {
	return api.NewError("MULTIPLE_PLAYLISTS_NOT_ALLOWED", "Multiple playlists are not allowed for subtitle download", http.StatusBadRequest)
}

func (SubtitleErrors) MixedResourcesNotAllowed() *api.AppError {
	return api.NewError("MIXED_RESOURCES_NOT_ALLOWED", "Mixed resources (playlists and videos) are not allowed for subtitle download", http.StatusBadRequest)
}

func (SubtitleErrors) NoSubtitlesFound() *api.AppError {
	return api.NewError("NO_SUBTITLES_FOUND", "No downloadable subtitles found for the given inputs", http.StatusBadRequest)
}

func (SubtitleErrors) ZipCreationFailed(err error) *api.AppError {
	return api.NewError("ZIP_CREATION_FAILED", "Failed to create ZIP archive", http.StatusInternalServerError).Wrap(err)
}

func (SubtitleErrors) InternalError(err error) *api.AppError {
	return api.NewError("INTERNAL_ERROR", "Something went wrong", http.StatusInternalServerError).Wrap(err)
}
