package youtubeurl

type InputType string

const (
	InputTypeVideo    InputType = "video"
	InputTypePlaylist InputType = "playlist"
	InputTypeChannel  InputType = "channel"
)

type YouTubeInput struct {
	InputType     InputType
	ID            string
	OriginalURL   string
	NormalizedURL string
}

type ParseItemResult struct {
	Input *YouTubeInput
	Error error
	Raw   string
}

type LineResult struct {
	Line  int
	Raw   string
	Input *YouTubeInput
	Error error
}
