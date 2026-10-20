package playlist

import (
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/api"
)

type PlaylistErrors struct{}

var Err = PlaylistErrors{}

func (PlaylistErrors) BadRequest(msg string) *api.AppError {
	return api.NewError(
		"PLAYLIST_BAD_REQUEST",
		msg,
		http.StatusBadRequest,
	)
}

func (PlaylistErrors) InvalidURL() *api.AppError {
	return api.NewError(
		"INVALID_PLAYLIST_URL",
		"Invalid playlist URL",
		http.StatusBadRequest,
	)
}

func (PlaylistErrors) NotAPlaylist() *api.AppError {
	return api.NewError(
		"NOT_A_PLAYLIST",
		"The provided URL is not a playlist",
		http.StatusBadRequest,
	)
}

func (PlaylistErrors) InvalidProvider() *api.AppError {
	return api.NewError(
		"INVALID_PROVIDER",
		"Invalid provider",
		http.StatusBadRequest,
	)
}

func (PlaylistErrors) FetchFailed(err error) *api.AppError {
	return api.NewError(
		"PLAYLIST_FETCH_FAILED",
		"Failed to fetch playlist items",
		http.StatusInternalServerError,
	).Wrap(err)
}
