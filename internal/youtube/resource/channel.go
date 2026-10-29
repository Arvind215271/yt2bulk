package resource

import (
	"context"

	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata"
)

// ChannelMetadata represents basic information about a YouTube channel.
type ChannelMetadata struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// ProcessChannel fetches channel metadata using the provided service.
func ProcessChannel(
	ctx context.Context,
	metadataService *metadata.Service,
	channelURL string,
) (*ChannelMetadata, error) {
	playlist, err := metadataService.GetChannel(ctx, channelURL, metadata.ProviderYTDLP)
	if err != nil {
		return nil, err
	}

	return &ChannelMetadata{
		ID:          playlist.ID,
		Title:       playlist.Title,
		Description: playlist.Description,
		URL:         channelURL,
	}, nil
}
