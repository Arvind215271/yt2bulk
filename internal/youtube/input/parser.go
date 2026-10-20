package youtubeurl

import (
	"net/url"
	"strings"

	"github.com/Arvind215271/yt2bulk/internal/youtube"
)


// ParseURL validates a YouTube URL,
// extracts the resource ID,
// and returns a normalized YouTube input.
func Parse(raw string) (*YouTubeInput, error) {
	// get the raw URL and  trim it... because there might be some extra sspace around that would create problem if not in proper form
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return nil, youtube.Err.URL.EmptyURL()
	}

	u, err := url.Parse(raw)
	// here it will parse the url into  different object like way. Like it will have stuff like u.query, u.params,etc, that any url contain. So then we can easily extract infromatoin from it.
	if err != nil {
		return nil, youtube.Err.URL.InvalidURL().Wrap(err)
	}

	if !isYouTubeHost(u.Host) {
		return nil, youtube.Err.URL.InvalidDomain()
	}

	// first check for video URLs.
	//
	// this intentionally takes priority over playlists because a URL like:
	// https://www.youtube.com/watch?v=abc123&list=PL123
	//
	// is still a video being watched inside playlist context.
	videoID := extractVideoID(u)
	if videoID != "" {
		return &YouTubeInput{
			InputType:    InputTypeVideo,
			ID:           videoID,
			OriginalURL:  raw,
			NormalizedURL: normalizeVideoURL(videoID),
		}, nil
	}

	// youtube playlist have the query params of ?list=playlist_id that we need to find
	playlistID := u.Query().Get("list")
	if playlistID != "" {
		return &YouTubeInput{
			InputType:    InputTypePlaylist,
			ID:           playlistID,
			OriginalURL:  raw,
			NormalizedURL: normalizePlaylistURL(playlistID),
		}, nil
	}

	return nil, youtube.Err.URL.MissingID()
}

// extractVideoID extracts a video ID from supported YouTube URL formats.
func extractVideoID(u *url.URL) string {
	host := strings.ToLower(u.Host)

	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimPrefix(host, "m.")

	// https://youtu.be/VIDEO_ID
	if host == "youtu.be" {
		return strings.Trim(u.Path, "/")
	}

	// https://www.youtube.com/watch?v=VIDEO_ID
	videoID := u.Query().Get("v")
	if videoID != "" {
		return videoID
	}

	// https://www.youtube.com/shorts/VIDEO_ID
	//
	// shorts are treadted same as video. So... we can normally use those as well.
	path := strings.Trim(u.Path, "/")
	if strings.HasPrefix(path, "shorts/") {
		return strings.TrimPrefix(path, "shorts/")
	}

	return ""
}

// isYouTubeHost validates supported YouTube hosts.
func isYouTubeHost(host string) bool {
	host = strings.ToLower(host)

	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimPrefix(host, "m.")

	return host == "youtube.com" ||
		host == "youtu.be" ||
		strings.HasPrefix(host, "youtube.")
}

// normalizePlaylistURL converts a playlist ID into a canonical URL.
func normalizePlaylistURL(id string) string {
	return "https://www.youtube.com/playlist?list=" + id
}

// normalizeVideoURL converts a video ID into a canonical URL.
func normalizeVideoURL(id string) string {
	return "https://www.youtube.com/watch?v=" + id
}