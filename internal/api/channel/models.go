package channel

type OutputType string

const (
	OutputTypeID   OutputType = "id"
	OutputTypeURL  OutputType = "url"
	OutputTypeBoth OutputType = "both"
)

type ChannelRequest struct {
	URL        string     `json:"url" validate:"required"`
	Provider   string     `json:"provider" validate:"omitempty"`
	Output     OutputType `json:"output" validate:"omitempty"`
	LimitMode  string     `json:"limit_mode" validate:"omitempty"`
	Limit      int        `json:"limit" validate:"omitempty"`
	RangeStart int        `json:"range_start" validate:"omitempty"`
	RangeEnd   int        `json:"range_end" validate:"omitempty"`
}

type VideoEntry struct {
	ID       string `json:"id,omitempty"`
	URL      string `json:"url,omitempty"`
	Position int    `json:"position"`
	Title    string `json:"title,omitempty"`
	Duration int    `json:"duration,omitempty"`
}

type ChannelMetadataResponse struct {
	ChannelID   string `json:"channel_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

type ChannelVideosResponse struct {
	ChannelID string       `json:"channel_id"`
	Total     int          `json:"total"`
	Videos    []VideoEntry `json:"videos"`
}
