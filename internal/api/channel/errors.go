package channel

import (
	"net/http"

	"github.com/Arvind215271/yt2bulk/internal/api"
)

type ChannelErrors struct{}

var Err = ChannelErrors{}

func (ChannelErrors) BadRequest(msg string) *api.AppError {
	return api.NewError(
		"CHANNEL_BAD_REQUEST",
		msg,
		http.StatusBadRequest,
	)
}

func (ChannelErrors) InvalidURL() *api.AppError {
	return api.NewError(
		"INVALID_CHANNEL_URL",
		"Invalid channel URL",
		http.StatusBadRequest,
	)
}

func (ChannelErrors) NotAChannel() *api.AppError {
	return api.NewError(
		"NOT_A_CHANNEL",
		"The provided URL is not a channel",
		http.StatusBadRequest,
	)
}

func (ChannelErrors) InvalidProvider() *api.AppError {
	return api.NewError(
		"INVALID_PROVIDER",
		"Invalid provider",
		http.StatusBadRequest,
	)
}

func (ChannelErrors) FetchFailed(err error) *api.AppError {
	return api.NewError(
		"CHANNEL_FETCH_FAILED",
		"Failed to fetch channel items",
		http.StatusInternalServerError,
	).Wrap(err)
}
