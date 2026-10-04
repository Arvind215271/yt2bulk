package fields

// Field constants
const (
	FieldID                  = "id"
	FieldErrors              = "errors"
	FieldTitle               = "title"
	FieldDescription         = "description"
	FieldChannelID           = "channel_id"
	FieldChannelTitle        = "channel_title"
	FieldThumbnails          = "thumbnails"
	FieldPublishedAt         = "published_at"
	FieldDuration            = "duration"
	FieldDurationSeconds     = "duration_seconds"
	FieldDurationMinutes     = "duration_minutes"
	FieldDurationTimestamp   = "duration_timestamp"
	FieldViewCount           = "view_count"
	FieldLikeCount           = "like_count"
	FieldCommentCount        = "comment_count"
	FieldTags                = "tags"
	FieldCategoryID          = "category_id"
	FieldCaptionAvailable    = "caption_available"
	FieldPrivacyStatus       = "privacy_status"
	FieldLiveBroadcastStatus = "live_broadcast_status"

	FieldDescriptionChapters = "description_chapters"
	FieldDescriptionLinks    = "description_links"
	FieldDescriptionEmails   = "description_emails"
	FieldDescriptionCleaned  = "description_cleaned"

	FieldTranscriptText   = "transcript_text"
	FieldTranscriptSignal = "transcript_signal"
	FieldSubtitleMetadata = "subtitle_metadata"
	FieldVideoID          = "video_id"
	FieldPosition         = "position"
	FieldAddedAt          = "added_at"
)

// Registry is the single source of truth for valid fields and their definitions.
var Registry = map[string]FieldDefinition{
	FieldID: {
		Name:    FieldID,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldErrors: {
		Name:    FieldErrors,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldTitle: {
		Name:    FieldTitle,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldDescription: {
		Name:    FieldDescription,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldChannelID: {
		Name:    FieldChannelID,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldChannelTitle: {
		Name:    FieldChannelTitle,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldThumbnails: {
		Name:    FieldThumbnails,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldPublishedAt: {
		Name:    FieldPublishedAt,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldDuration: {
		Name:    FieldDuration,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldDurationSeconds: {
		Name:    FieldDurationSeconds,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldDurationMinutes: {
		Name:    FieldDurationMinutes,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldDurationTimestamp: {
		Name:    FieldDurationTimestamp,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldViewCount: {
		Name:    FieldViewCount,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldLikeCount: {
		Name:    FieldLikeCount,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldCommentCount: {
		Name:    FieldCommentCount,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldTags: {
		Name:    FieldTags,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldCategoryID: {
		Name:    FieldCategoryID,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldCaptionAvailable: {
		Name:    FieldCaptionAvailable,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldPrivacyStatus: {
		Name:    FieldPrivacyStatus,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldLiveBroadcastStatus: {
		Name:    FieldLiveBroadcastStatus,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldDescriptionChapters: {
		Name:    FieldDescriptionChapters,
		Group:   GroupDescription,
		Default: true,
	},
	FieldDescriptionLinks: {
		Name:    FieldDescriptionLinks,
		Group:   GroupDescription,
		Default: true,
	},
	FieldDescriptionEmails: {
		Name:    FieldDescriptionEmails,
		Group:   GroupDescription,
		Default: true,
	},
	FieldDescriptionCleaned: {
		Name:    FieldDescriptionCleaned,
		Group:   GroupDescription,
		Default: true,
	},
	FieldTranscriptText: {
		Name:    FieldTranscriptText,
		Group:   GroupTranscript,
		Default: true,
	},
	FieldTranscriptSignal: {
		Name:    FieldTranscriptSignal,
		Group:   GroupSignal,
		Default: true,
	},
	FieldSubtitleMetadata: {
		Name:    FieldSubtitleMetadata,
		Group:   GroupTranscript,
		Default: true,
	},
	FieldVideoID: {
		Name:    FieldVideoID,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldPosition: {
		Name:    FieldPosition,
		Group:   GroupMetadata,
		Default: true,
	},
	FieldAddedAt: {
		Name:    FieldAddedAt,
		Group:   GroupMetadata,
		Default: true,
	},
}
