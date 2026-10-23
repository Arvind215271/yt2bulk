package subtitle

type SubtitleOptionsRequest struct {
	Inputs []string `json:"inputs" validate:"required,min=1"`
}

type SubtitlePreferenceRequest struct {
	Language string `json:"language" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=manual automatic manual>automatic automatic>manual"`
}

type SubtitleDownloadRequest struct {
	Inputs       []string                    `json:"inputs" validate:"required,min=1"`
	PlayerClient string                      `json:"player_client,omitempty"`
	Preferences  []SubtitlePreferenceRequest `json:"preferences"`
	Format       string                      `json:"format,omitempty" validate:"omitempty,oneof=json3 vtt srt srv1 srv2 srv3 ttml"`
}
