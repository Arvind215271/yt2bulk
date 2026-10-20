package youtubeapi


import (
	yt "google.golang.org/api/youtube/v3"
	youtube 	"github.com/Arvind215271/yt2bulk/internal/youtube"

)

func (p *Provider) mapVideo(video *yt.Video) youtube.Video {
	seconds := youtube.ParseYouTubeDuration(video.ContentDetails.Duration)

	return youtube.Video{
		ID:          video.Id,
		Title:       video.Snippet.Title,
		Description: video.Snippet.Description,

		ChannelID:    video.Snippet.ChannelId,
		ChannelTitle: video.Snippet.ChannelTitle,

		Thumbnails: p.getVideoThumbnails(video),

		PublishedAt: p.parseTime(video.Snippet.PublishedAt),

		Duration:          video.ContentDetails.Duration,
		DurationSeconds:   int64(seconds),
		DurationMinutes:   youtube.GetDurationMinutes(seconds),
		DurationTimestamp: youtube.GetDurationTimestamp(seconds),

		ViewCount:    video.Statistics.ViewCount,
		LikeCount:    video.Statistics.LikeCount,
		CommentCount: video.Statistics.CommentCount,

		Tags: video.Snippet.Tags,

		CategoryID: video.Snippet.CategoryId,

		CaptionAvailable: video.ContentDetails.Caption == "true",

		PrivacyStatus:       video.Status.PrivacyStatus,
		LiveBroadcastStatus: video.Snippet.LiveBroadcastContent,
	}
}



func (p *Provider) mapPlaylistVideos(
	items []*yt.PlaylistItem,
	videos []*yt.Video,
) []youtube.PlaylistVideo {

	videoMap := make(map[string]*yt.Video, len(videos))
	for _, video := range videos {
		videoMap[video.Id] = video
	}

	result := make([]youtube.PlaylistVideo, 0, len(items))

	for _, item := range items {

		if item == nil ||
			item.Snippet == nil ||
			item.ContentDetails == nil {
			continue
		}

		video := videoMap[item.ContentDetails.VideoId]

		if video == nil ||
			video.Snippet == nil ||
			video.ContentDetails == nil ||
			video.Statistics == nil ||
			video.Status == nil {
			p.logger.Warn("video missing metadata in playlist", "videoID", item.ContentDetails.VideoId)
			continue
		}

		result = append(result, youtube.PlaylistVideo{
			Video:    p.mapVideo(video),
			Position: int(item.Snippet.Position),
			AddedAt:  p.parseTime(item.Snippet.PublishedAt),
		})
	}

	return result
}
