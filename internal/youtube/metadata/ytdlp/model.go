package ytdlp

// YTOutput represents the full structure returned by yt-dlp -j
type YTOutput struct {
	ID                string                      `json:"id"`
	Title             string                      `json:"title"`
	FullTitle         string                      `json:"fulltitle"`
	Description       string                      `json:"description"`
	Duration          int                         `json:"duration"`
	ViewCount         uint64                      `json:"view_count"`
	LikeCount         uint64                      `json:"like_count"`
	CommentCount      uint64                      `json:"comment_count"`
	Channel           string                      `json:"channel"`
	ChannelID         string                      `json:"channel_id"`
	Thumbnail         string                      `json:"thumbnail"`
	Tags              []string                    `json:"tags"`
	Categories        []string                    `json:"categories"`
	UploadDate        string                      `json:"upload_date"`
	WebpageURL        string                      `json:"webpage_url"`
	Uploader          string                      `json:"uploader"`
	UploaderID        string                      `json:"uploader_id"`
	UploaderURL       string                      `json:"uploader_url"`
	Availability      string                      `json:"availability"`
	MediaType         string                      `json:"media_type"`
	AgeLimit          int                         `json:"age_limit"`
	Extractor         string                      `json:"extractor"`
	DisplayID         string                      `json:"display_id"`
	Filename          string                      `json:"filename"`
	Ext               string                      `json:"ext"`
	Width             int                         `json:"width"`
	Height            int                         `json:"height"`
	Resolution        string                      `json:"resolution"`
	FPS               int                         `json:"fps"`
	VCodec            string                      `json:"vcodec"`
	ACodec            string                      `json:"acodec"`
	AudioChannels     int                         `json:"audio_channels"`
	Language          string                      `json:"language"`
	Subtitles         map[string][]SubtitleFormat `json:"subtitles"`
	AutomaticCaptions map[string][]SubtitleFormat `json:"automatic_captions"`
	Chapters          []map[string]interface{}    `json:"chapters"`
	Thumbnails        []ThumbnailEntry            `json:"thumbnails"`
}

type SubtitleFormat struct {
	Ext      string `json:"ext"`
	URL      string `json:"url"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Client   string `json:"client,omitempty"`
}

type YTPlaylistOutput struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	Channel       string            `json:"channel"`
	ChannelID     string            `json:"channel_id"`
	Thumbnail     string            `json:"thumbnail"`
	Tags          []string          `json:"tags"`
	WebpageURL    string            `json:"webpage_url"`
	Uploader      string            `json:"uploader"`
	PlaylistCount int               `json:"playlist_count"`
	Entries       []YTPlaylistEntry `json:"entries"`
}

type PlaylistSurfaceMetadata struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Channel       string   `json:"channel"`
	ChannelID     string   `json:"channel_id"`
	Thumbnail     string   `json:"thumbnail"`
	Tags          []string `json:"tags"`
	WebpageURL    string   `json:"webpage_url"`
	Uploader      string   `json:"uploader"`
	PlaylistCount int      `json:"playlist_count"`
}

type YTPlaylistEntry struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Duration     int               `json:"duration"`
	ViewCount    uint64            `json:"view_count"`
	ChannelID    string            `json:"channel_id"`
	Channel      string            `json:"channel"`
	ChannelIDAlt string            `json:"uploader_id"`
	WebpageURL   string            `json:"webpage_url"`
	OriginalURL  string            `json:"original_url"`
	Thumbnail    string            `json:"thumbnail"`
	Thumbnails   []ThumbnailEntry  `json:"thumbnails"`
	Type         string            `json:"_type,omitempty"`
	Entries      []YTPlaylistEntry `json:"entries,omitempty"`
}

type ThumbnailEntry struct {
	URL        string `json:"url"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Preference int    `json:"preference"`
	ID         string `json:"id"`
	Resolution string `json:"resolution"`
}
