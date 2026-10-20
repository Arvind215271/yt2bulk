package resource

import (
	"github.com/Arvind215271/yt2bulk/internal/youtube"
)

// ItemToVideo maps a playlist item snapshot into a fully populated youtube.Video.
// This allows the pipeline to receive a standard Video model for any playlist item.
func ItemToVideo(item youtube.PlaylistItem) youtube.Video {
	return youtube.Video{
		ID:              item.VideoID,
		Title:           item.Title,
		Description:     item.Description,
		Duration:        item.Duration.String(),
		DurationSeconds: int64(item.Duration.Seconds()),
		ViewCount:       item.ViewCount,
		ChannelID:       item.ChannelID,
		ChannelTitle:    item.ChannelTitle,
		Thumbnails:      item.Thumbnails,
		PrivacyStatus:   item.PrivacyStatus,
	}
}
