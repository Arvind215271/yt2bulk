package export

import (
	subtitleapi "github.com/Arvind215271/yt2bulk/internal/api/subtitle"
	"github.com/Arvind215271/yt2bulk/internal/youtube/signal"
	"github.com/Arvind215271/yt2bulk/internal/youtube/subtitle"
	"github.com/Arvind215271/yt2bulk/internal/youtube/transcript"
)

type CommonExportFields struct {
	Subtitle    *subtitle.DownloadRequest               `json:"subtitle,omitempty"`
	Preferences []subtitleapi.SubtitlePreferenceRequest `json:"preferences,omitempty"`
	Transcript  *transcript.ProcessingRequest           `json:"transcript,omitempty"`
	Signal      *signal.SignalRequest                   `json:"signal,omitempty"`
	Format      string                                  `json:"format"`
	Fields      []string                                `json:"fields,omitempty"`
}

type ExportRequest struct {
	Inputs       []string `json:"inputs"`
	PlayerClient string   `json:"player_client,omitempty"`
	CommonExportFields
}
