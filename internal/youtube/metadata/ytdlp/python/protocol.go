package python

import "context"

type VideoRequest struct {
	Ctx     context.Context `json:"-"`
	VideoID string          `json:"video_id"`
}

type WarmupRequest struct {
	Ctx context.Context `json:"-"`
	Cmd string          `json:"cmd"`
}

type PlaylistRequest struct {
	Ctx        context.Context `json:"-"`
	Cmd        string          `json:"cmd"`
	PlaylistID string          `json:"playlist_id"`
}

type PlaylistMetadataRequest struct {
	Ctx        context.Context `json:"-"`
	Cmd        string          `json:"cmd"`
	PlaylistID string          `json:"playlist_id"`
}

type PlaylistSurfaceMetadataRequest struct {
	Ctx        context.Context `json:"-"`
	Cmd        string          `json:"cmd"`
	PlaylistID string          `json:"playlist_id"`
}

type WorkerResponse struct {
	Ok   bool           `json:"ok"`
	Data map[string]any `json:"data"`
	Err  string         `json:"error"`
}

type SubtitleRequest struct {
	Ctx        context.Context `json:"-"`
	Cmd        string          `json:"cmd"`
	VideoID    string          `json:"video_id"`
	Language   string          `json:"language"`
	Type       string          `json:"type"`   // manual | automatic
	Format     string          `json:"format"` // json3, vtt...
	OutputPath string          `json:"output_path"`
}

type SubtitleResponse struct {
	Ok       bool   `json:"ok"`
	Filename string `json:"filename"`
	Language string `json:"language"`
	Format   string `json:"format"`
	Err      string `json:"error"`
}
