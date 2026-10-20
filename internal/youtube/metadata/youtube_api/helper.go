package youtubeapi

import (
	"strconv"
	"time"

	youtube "github.com/Arvind215271/yt2bulk/internal/youtube"
	yt "google.golang.org/api/youtube/v3"
)


func (p *Provider) parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		p.logger.Warn("failed to parse time", "value", s, "error", err)
		return time.Time{}
	}
	return t
}

func (p *Provider) parseUint(s string) uint64 {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		p.logger.Warn("failed to parse uint", "value", s, "error", err)
		return 0
	}
	return v
}

func (p *Provider) getPlaylistThumbnail(playlist *yt.Playlist) string {
	if playlist.Snippet == nil || playlist.Snippet.Thumbnails == nil {
		return ""
	}

	if playlist.Snippet.Thumbnails.High != nil {
		return playlist.Snippet.Thumbnails.High.Url
	}
	if playlist.Snippet.Thumbnails.Medium != nil {
		return playlist.Snippet.Thumbnails.Medium.Url
	}
	if playlist.Snippet.Thumbnails.Default != nil {
		return playlist.Snippet.Thumbnails.Default.Url
	}
	return ""
}

func (p *Provider) getVideoThumbnails(video *yt.Video) []youtube.Thumbnail {
	var thumbnails []youtube.Thumbnail
	if video.Snippet == nil || video.Snippet.Thumbnails == nil {
		return thumbnails
	}

	if video.Snippet.Thumbnails.Default != nil {
		thumbnails = append(thumbnails, youtube.Thumbnail{
			URL:    video.Snippet.Thumbnails.Default.Url,
			Width:  int(video.Snippet.Thumbnails.Default.Width),
			Height: int(video.Snippet.Thumbnails.Default.Height),
		})
	}
	if video.Snippet.Thumbnails.Medium != nil {
		thumbnails = append(thumbnails, youtube.Thumbnail{
			URL:    video.Snippet.Thumbnails.Medium.Url,
			Width:  int(video.Snippet.Thumbnails.Medium.Width),
			Height: int(video.Snippet.Thumbnails.Medium.Height),
		})
	}
	if video.Snippet.Thumbnails.High != nil {
		thumbnails = append(thumbnails, youtube.Thumbnail{
			URL:    video.Snippet.Thumbnails.High.Url,
			Width:  int(video.Snippet.Thumbnails.High.Width),
			Height: int(video.Snippet.Thumbnails.High.Height),
		})
	}
	return thumbnails
}
