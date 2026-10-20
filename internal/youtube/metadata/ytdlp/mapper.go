package ytdlp

import (
	"fmt"
	"time"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
)

func MapVideo(meta YTOutput) youtube.Video {
	durationSeconds := int64(meta.Duration)
	duration := time.Duration(durationSeconds) * time.Second

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	durationTimestamp := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	if hours == 0 {
		durationTimestamp = fmt.Sprintf("%02d:%02d", minutes, seconds)
	}

	return youtube.Video{
		ID:                meta.ID,
		Title:             meta.Title,
		Description:       meta.Description,
		DurationSeconds:   durationSeconds,
		DurationMinutes:   duration.Minutes(),
		DurationTimestamp: durationTimestamp,
		Duration:          durationTimestamp,
		ViewCount:         meta.ViewCount,
		LikeCount:         meta.LikeCount,
		CommentCount:      meta.CommentCount,
		ChannelTitle:      meta.Channel,
		ChannelID:         meta.ChannelID,
		Thumbnails:        mapThumbnails(meta),
		Tags:              meta.Tags,
		CategoryID:        "", // Need to map categories if available
		PrivacyStatus:     meta.Availability,
		SubtitleMetadata:  mapSubtitleMetadata(meta),
	}
}

func mapThumbnails(meta YTOutput) []youtube.Thumbnail {
	var thumbnails []youtube.Thumbnail
	for _, th := range meta.Thumbnails {
		thumbnails = append(thumbnails, youtube.Thumbnail{
			URL:    th.URL,
			Width:  th.Width,
			Height: th.Height,
		})
	}
	if len(thumbnails) == 0 && meta.Thumbnail != "" {
		thumbnails = append(thumbnails, youtube.Thumbnail{
			URL: meta.Thumbnail,
		})
	}
	return thumbnails
}

func mapSubtitleMetadata(meta YTOutput) subtitle.SubtitleMetadata {
	return subtitle.SubtitleMetadata{
		Manual:    mapTracks(meta.Subtitles),
		Automatic: mapTracks(meta.AutomaticCaptions),
	}
}

func mapTracks(raw map[string][]SubtitleFormat) []subtitle.SubtitleTrack {
	var tracks []subtitle.SubtitleTrack
	for lang, formats := range raw {
		track := subtitle.SubtitleTrack{
			LanguageCode: lang,
			LanguageName: lang, // Could be enhanced with a lookup
			Formats:      []string{},
		}
		seen := make(map[string]bool)
		for _, f := range formats {
			if !seen[f.Ext] {
				track.Formats = append(track.Formats, f.Ext)
				seen[f.Ext] = true
			}
		}
		tracks = append(tracks, track)
	}
	return tracks
}

func MapPlaylist(meta YTPlaylistOutput) youtube.Playlist {
	var thumbnails []youtube.Thumbnail
	// Map primary thumbnail
	if meta.Thumbnail != "" {
		thumbnails = append(thumbnails, youtube.Thumbnail{
			URL: meta.Thumbnail,
		})
	}

	itemCount := meta.PlaylistCount
	if itemCount == 0 {
		itemCount = len(meta.Entries)
	}

	playlist := youtube.Playlist{
		ID:           meta.ID,
		Title:        meta.Title,
		Description:  meta.Description,
		ChannelTitle: meta.Channel,
		ChannelID:    meta.ChannelID,
		Thumbnails:   thumbnails,
		Tags:         meta.Tags,
		ItemCount:    itemCount,
	}

	// Map entries as lightweight items
	for _, entry := range meta.Entries {
		item := youtube.PlaylistItem{
			VideoID:     entry.ID,
			Title:       entry.Title,
			Description: entry.Description,
			// Duration is in seconds for entries in YTPlaylistEntry
			Duration:     time.Duration(entry.Duration) * time.Second,
			ViewCount:    entry.ViewCount,
			ChannelID:    entry.ChannelID,
			ChannelTitle: entry.Channel,
			URL:          entry.WebpageURL,
		}

		// Map thumbnails for items
		for _, t := range entry.Thumbnails {
			item.Thumbnails = append(item.Thumbnails, youtube.Thumbnail{
				URL:    t.URL,
				Width:  t.Width,
				Height: t.Height,
			})
		}

		playlist.Items = append(playlist.Items, item)
	}

	return playlist
}

func MapToPlaylistSurfaceMetadata(data map[string]any) PlaylistSurfaceMetadata {
	return PlaylistSurfaceMetadata{
		ID:            getString(data, "id"),
		Title:         getString(data, "title"),
		Description:   getString(data, "description"),
		Channel:       getString(data, "channel"),
		ChannelID:     getString(data, "channel_id"),
		Thumbnail:     getString(data, "thumbnail"),
		Tags:          getStringSlice(data, "tags"),
		WebpageURL:    getString(data, "webpage_url"),
		Uploader:      getString(data, "uploader"),
		PlaylistCount: getInt(data, "playlist_count"),
	}
}

func getString(data map[string]any, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getStringSlice(data map[string]any, key string) []string {
	if val, ok := data[key].([]string); ok {
		return val
	}
	if val, ok := data[key].([]any); ok {
		res := make([]string, len(val))
		for i, v := range val {
			if s, ok := v.(string); ok {
				res[i] = s
			}
		}
		return res
	}
	return nil
}

func getInt(data map[string]any, key string) int {
	if val, ok := data[key].(int); ok {
		return val
	}
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	return 0
}
