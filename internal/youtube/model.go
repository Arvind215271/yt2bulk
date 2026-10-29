// ./internal/youtube/model.go

package youtube

import (
	"time"

	"github.com/Arvind215271/yt2bulk/internal/youtube/description"
	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
	"github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
)

// ResourceType defines the type of resource wrapper (e.g. video or playlist).
type ResourceType string

const (
	ResourceTypeVideo    ResourceType = "video"
	ResourceTypePlaylist ResourceType = "playlist"
)

// Resource represents a mixed-resource API wrapper container.
type Resource struct {
	ID       string       `json:"id"`
	Type     ResourceType `json:"type"`
	Video    *Video       `json:"video,omitempty"`
	Playlist *Playlist    `json:"playlist,omitempty"`
}

// Error represents an error encountered during video or playlist processing.
type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// Thumbnail represents a video or playlist thumbnail image at a specific resolution.
type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

// PlaylistItem is a lightweight snapshot returned by the playlist API.
//
// It exists only as an optimization during pipeline execution and is
// never exposed directly by the public API.
type PlaylistItem struct {
	VideoID string `json:"video_id"`

	// Playlist metadata
	Position int       `json:"position"`
	AddedAt  time.Time `json:"added_at"`

	// Video snapshot
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	Duration  time.Duration `json:"duration,omitempty"`
	ViewCount uint64        `json:"view_count,omitempty"`

	// Channel snapshot
	ChannelID    string `json:"channel_id,omitempty"`
	ChannelTitle string `json:"channel_title,omitempty"`

	// URLs
	URL string `json:"url,omitempty"`

	// Thumbnails
	Thumbnails []Thumbnail `json:"thumbnails,omitempty"`

	// Playlist/video status
	PrivacyStatus string `json:"privacy_status,omitempty"`
}

// PlaylistVideo represents a fully processed video that belongs to a playlist,
// including its playlist-specific context like position and added timestamp.
type PlaylistVideo struct {
	Video

	Position int       `json:"position"`
	AddedAt  time.Time `json:"added_at"`
}

// Playlist represents a YouTube playlist along with its aggregated metadata,
// statistics, internal snapshot items, and fully processed videos.
type Playlist struct {
	// Identity
	ID string `json:"id"`

	// Metadata
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	ChannelID     string      `json:"channel_id"`
	ChannelTitle  string      `json:"channel_title"`
	Thumbnails    []Thumbnail `json:"thumbnails,omitempty"`
	Tags          []string    `json:"tags,omitempty"`
	PrivacyStatus string      `json:"privacy_status"`
	PublishedAt   time.Time   `json:"published_at"`
	ModifiedAt    time.Time   `json:"modified_at"`

	// Statistics
	ItemCount int `json:"item_count"`

	// Internal
	// Items is a lightweight snapshot returned by the playlist API.
	// It exists only as an optimization during pipeline execution and is
	// never exposed directly by the public API.
	Items []PlaylistItem `json:"-"`

	// Public
	Videos []PlaylistVideo `json:"videos,omitempty"`

	// Errors
	Errors []Error `json:"errors,omitempty"`
}

// ChannelThumbnail represents a channel thumbnail with extra yt-dlp specific fields.
type ChannelThumbnail struct {
	URL        string `json:"url"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Preference int    `json:"preference,omitempty"`
	ID         string `json:"id,omitempty"`
	Resolution string `json:"resolution,omitempty"`
}

// CommonChannelFields represents the common fields found in channel entries and data.
type CommonChannelFields struct {
	ID                   string             `json:"id"`
	Channel              string             `json:"channel,omitempty"`
	ChannelID            string             `json:"channel_id,omitempty"`
	Title                string             `json:"title,omitempty"`
	Availability         string             `json:"availability,omitempty"`
	ChannelFollowerCount int64              `json:"channel_follower_count,omitempty"`
	Description          string             `json:"description,omitempty"`
	Tags                 []string           `json:"tags,omitempty"`
	Thumbnails           []ChannelThumbnail `json:"thumbnails,omitempty"`
	UploaderID           string             `json:"uploader_id,omitempty"`
	UploaderURL          string             `json:"uploader_url,omitempty"`
	ChannelIsVerified    bool               `json:"channel_is_verified,omitempty"`
	ModifiedDate         interface{}        `json:"modified_date,omitempty"`
	ViewCount            interface{}        `json:"view_count,omitempty"`
	PlaylistCount        interface{}        `json:"playlist_count,omitempty"`
	Uploader             string             `json:"uploader,omitempty"`
	ChannelURL           string             `json:"channel_url,omitempty"`
	Type                 string             `json:"_type,omitempty"`
}

// ChannelEntry represents an entry (playlist or video) inside a channel response.
type ChannelEntry struct {
	CommonChannelFields
	Duration   float64        `json:"duration,omitempty"`
	Timestamp  interface{}    `json:"timestamp,omitempty"`
	LiveStatus string         `json:"live_status,omitempty"`
	IEKey      string         `json:"ie_key,omitempty"`
	URL        string         `json:"url,omitempty"`
	Entries    []ChannelEntry `json:"entries,omitempty"`
}

// ChannelResponse represents the top-level yt-dlp worker response for a channel.
type ChannelResponse struct {
	Ok   bool        `json:"ok"`
	Data ChannelData `json:"data"`
}

// ChannelData represents the channel object inside ChannelResponse.
type ChannelData struct {
	CommonChannelFields
	Entries []ChannelEntry `json:"entries,omitempty"`
}

// Video represents everything known about a single YouTube video,
// including metadata, transcripts, descriptions, subtitles, and errors.
type Video struct {
	// Identity
	ID string `json:"id"`

	// Basic metadata
	Title       string    `json:"title"`
	Tags        []string  `json:"tags,omitempty"`
	CategoryID  string    `json:"category_id"`
	PublishedAt time.Time `json:"published_at"`

	// Channel
	ChannelID    string `json:"channel_id"`
	ChannelTitle string `json:"channel_title"`

	// Statistics
	ViewCount    uint64 `json:"view_count"`
	LikeCount    uint64 `json:"like_count"`
	CommentCount uint64 `json:"comment_count"`

	// Duration
	Duration          string  `json:"duration"`
	DurationSeconds   int64   `json:"duration_seconds"`
	DurationMinutes   float64 `json:"duration_minutes"`
	DurationTimestamp string  `json:"duration_timestamp"`

	// Description
	Description         string               `json:"description"`
	DescriptionMetadata description.Metadata `json:"description_metadata,omitempty"`
	DescriptionChapters string               `json:"description_chapters"`
	DescriptionLinks    []string             `json:"description_links"`
	DescriptionEmails   []string             `json:"description_emails"`
	DescriptionCleaned  string               `json:"description_cleaned"`

	// Subtitle
	SubtitleMetadata subtitle.SubtitleMetadata `json:"subtitle_metadata"`

	// Transcript
	Transcript       *transcript.Transcript `json:"transcript,omitempty"`
	TranscriptText   string                 `json:"transcript_text"`
	TranscriptSignal string                 `json:"transcript_signal"`

	// Analysis
	// (Reserved for future analysis extensions)

	// Misc
	Thumbnails          []Thumbnail `json:"thumbnails,omitempty"`
	CaptionAvailable    bool        `json:"caption_available"`
	PrivacyStatus       string      `json:"privacy_status"`
	LiveBroadcastStatus string      `json:"live_broadcast_status"`

	// Errors
	Errors []Error `json:"errors,omitempty"`
}
