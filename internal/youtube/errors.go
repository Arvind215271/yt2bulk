// ./internal/youtube/errors.go
package youtube

import (
	"fmt"
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/api"
)

type Errors struct {
	Playlist PlaylistErrors
	Video    VideoErrors
	Export   ExportErrors
	URL      URLErrors
}

var Err = Errors{
	Playlist: PlaylistErrors{},
	Video:    VideoErrors{},
	Export:   ExportErrors{},
	URL:      URLErrors{},
}

type PlaylistErrors struct{}
type VideoErrors struct{}
type ExportErrors struct{}
type URLErrors struct{}

// Playlist Errors

func (PlaylistErrors) InvalidURL() *api.AppError {
	return api.NewError(
		"PLAYLIST_INVALID_URL",
		"Invalid YouTube playlist URL",
		http.StatusBadRequest,
	)
}

func (PlaylistErrors) InvalidDomain() *api.AppError {
	return api.NewError(
		"PLAYLIST_INVALID_DOMAIN",
		"URL must be a YouTube URL",
		http.StatusBadRequest,
	)
}

func (PlaylistErrors) MissingID() *api.AppError {
	return api.NewError(
		"PLAYLIST_MISSING_ID",
		"Playlist ID is missing",
		http.StatusBadRequest,
	)
}

func (PlaylistErrors) InvalidType() *api.AppError {
	return api.NewError(
		"PLAYLIST_INVALID_TYPE",
		"URL is not a playlist",
		http.StatusBadRequest,
	)
}

func (PlaylistErrors) NotFound() *api.AppError {
	return api.NewError(
		"PLAYLIST_NOT_FOUND",
		"Playlist not found",
		http.StatusNotFound,
	)
}

func (PlaylistErrors) FetchFailed() *api.AppError {
	return api.NewError(
		"PLAYLIST_FETCH_FAILED",
		"Failed to fetch playlist",
		http.StatusInternalServerError,
	)
}

// video Errors

func (VideoErrors) NotFound() *api.AppError {
	return api.NewError(
		"VIDEO_NOT_FOUND",
		"Video not found",
		http.StatusNotFound,
	)
}

func (VideoErrors) FetchFailed() *api.AppError {
	return api.NewError(
		"VIDEO_FETCH_FAILED",
		"Failed to fetch video",
		http.StatusInternalServerError,
	)
}

// export errors
func (ExportErrors) InvalidFormat() *api.AppError {
	return api.NewError(
		"EXPORT_INVALID_FORMAT",
		"Invalid export format",
		http.StatusBadRequest,
	)
}

func (ExportErrors) MarshalFailed() *api.AppError {
	return api.NewError(
		"EXPORT_MARSHAL_FAILED",
		"Failed to export data",
		http.StatusInternalServerError,
	)
}

func (ExportErrors) InvalidField() *api.AppError {
	return api.NewError(
		"EXPORT_INVALID_FIELD",
		"One or more export fields are invalid",
		http.StatusBadRequest,
	)
}

// URL Errors
func (URLErrors) EmptyURL() *api.AppError {
	return api.NewError(
		"YOUTUBE_EMPTY_URL",
		"URL cannot be empty",
		http.StatusBadRequest,
	)
}

func (URLErrors) InvalidURL() *api.AppError {
	return api.NewError(
		"YOUTUBE_INVALID_URL",
		"Invalid YouTube URL",
		http.StatusBadRequest,
	)
}

func (URLErrors) InvalidDomain() *api.AppError {
	return api.NewError(
		"YOUTUBE_INVALID_DOMAIN",
		"URL must be a YouTube URL",
		http.StatusBadRequest,
	)
}

func (URLErrors) MissingID() *api.AppError {
	return api.NewError(
		"YOUTUBE_MISSING_ID",
		"Could not extract a YouTube resource ID from URL",
		http.StatusBadRequest,
	)
}

// Preflight Errors
func (ExportErrors) LimitExceeded(reached, max int) *api.AppError {
	return api.NewError(
		"EXPORT_LIMIT_EXCEEDED",
		"Export limit exceeded",
		http.StatusBadRequest,
	).AddField("reached", fmt.Sprintf("%d", reached)).AddField("max", fmt.Sprintf("%d", max))
}
