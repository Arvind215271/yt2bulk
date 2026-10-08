package python

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/Arvind215271/askito/internal/logger"
)

type SingleClient struct {
	worker *PythonWorker
	logger *logger.Logger
}

func NewSingleWorker(
	workerID int,
	log *logger.Logger,
) (*SingleClient, error) {
	_, file, _, _ := runtime.Caller(0)

	script := filepath.Join(
		filepath.Dir(file),
		"python_worker_single.py",
	)

	worker, err := NewWorker(
		script,
		workerID,
		log,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create single worker %d: %w",
			workerID,
			err,
		)
	}

	return &SingleClient{
		worker: worker,
		logger: log.With(
			"worker_id", workerID,
		),
	}, nil
}

func (c *SingleClient) GetPlaylist(
	ctx context.Context,
	playlistID string,
) (map[string]any, error) {
	c.logger.Debug(
		"fetching playlist items",
		"playlist_id", playlistID,
	)

	req := PlaylistRequest{
		Cmd:        "playlist",
		PlaylistID: playlistID,
	}

	if err := c.worker.SendCommand(req); err != nil {
		return nil, fmt.Errorf(
			"send playlist request for playlist %s: %w",
			playlistID,
			err,
		)
	}

	respData, err := c.worker.ReceiveResponse()
	if err != nil {
		return nil, fmt.Errorf(
			"receive playlist response for playlist %s: %w",
			playlistID,
			err,
		)
	}

	var resp WorkerResponse

	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf(
			"decode playlist response for playlist %s: %w",
			playlistID,
			err,
		)
	}

	if !resp.Ok {
		c.logger.Warn(
			"playlist request failed",
			"playlist_id", playlistID,
			"error", resp.Err,
		)

		return nil, fmt.Errorf(
			"worker playlist error for playlist %s: %s",
			playlistID,
			resp.Err,
		)
	}

	c.logger.Debug(
		"playlist items fetched",
		"playlist_id", playlistID,
	)

	return resp.Data, nil
}

func (c *SingleClient) GetPlaylistMetadata(
	ctx context.Context,
	playlistID string,
) (map[string]any, error) {
	c.logger.Debug(
		"fetching playlist metadata",
		"playlist_id", playlistID,
	)

	req := PlaylistMetadataRequest{
		Cmd:        "playlist_metadata",
		PlaylistID: playlistID,
	}

	if err := c.worker.SendCommand(req); err != nil {
		return nil, fmt.Errorf(
			"send playlist metadata request for playlist %s: %w",
			playlistID,
			err,
		)
	}

	respData, err := c.worker.ReceiveResponse()
	if err != nil {
		return nil, fmt.Errorf(
			"receive playlist metadata response for playlist %s: %w",
			playlistID,
			err,
		)
	}

	var resp WorkerResponse

	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf(
			"decode playlist metadata response for playlist %s: %w",
			playlistID,
			err,
		)
	}

	if !resp.Ok {
		c.logger.Warn(
			"playlist metadata request failed",
			"playlist_id", playlistID,
			"error", resp.Err,
		)

		return nil, fmt.Errorf(
			"worker playlist metadata error for playlist %s: %s",
			playlistID,
			resp.Err,
		)
	}

	c.logger.Debug(
		"playlist metadata fetched",
		"playlist_id", playlistID,
	)

	return resp.Data, nil
}

func (c *SingleClient) GetPlaylistSurfaceMetadata(
	ctx context.Context,
	playlistID string,
) (map[string]any, error) {
	c.logger.Debug(
		"fetching playlist surface metadata",
		"playlist_id", playlistID,
	)

	req := PlaylistSurfaceMetadataRequest{
		Cmd:        "playlist_metadata",
		PlaylistID: playlistID,
	}

	if err := c.worker.SendCommand(req); err != nil {
		return nil, fmt.Errorf(
			"send playlist surface metadata request for playlist %s: %w",
			playlistID,
			err,
		)
	}

	respData, err := c.worker.ReceiveResponse()
	if err != nil {
		return nil, fmt.Errorf(
			"receive playlist surface metadata response for playlist %s: %w",
			playlistID,
			err,
		)
	}

	var resp WorkerResponse

	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf(
			"decode playlist surface metadata response for playlist %s: %w",
			playlistID,
			err,
		)
	}

	if !resp.Ok {
		c.logger.Warn(
			"playlist surface metadata request failed",
			"playlist_id", playlistID,
			"error", resp.Err,
		)

		return nil, fmt.Errorf(
			"worker playlist surface metadata error for playlist %s: %s",
			playlistID,
			resp.Err,
		)
	}

	c.logger.Debug(
		"playlist surface metadata fetched",
		"playlist_id", playlistID,
	)

	return resp.Data, nil
}

func (c *SingleClient) GetVideo(
	ctx context.Context,
	id string,
) (map[string]any, error) {
	c.logger.Debug(
		"fetching video metadata",
		"video_id", id,
	)

	if err := c.worker.SendCommand(
		VideoRequest{VideoID: id},
	); err != nil {
		return nil, fmt.Errorf(
			"send metadata request for video %s: %w",
			id,
			err,
		)
	}

	respData, err := c.worker.ReceiveResponse()
	if err != nil {
		return nil, fmt.Errorf(
			"receive metadata response for video %s: %w",
			id,
			err,
		)
	}

	var resp WorkerResponse

	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf(
			"decode metadata response for video %s: %w",
			id,
			err,
		)
	}

	if !resp.Ok {
		c.logger.Warn(
			"metadata request failed",
			"video_id", id,
			"error", resp.Err,
		)

		return nil, fmt.Errorf(
			"worker metadata error for video %s: %s",
			id,
			resp.Err,
		)
	}

	c.logger.Debug(
		"video metadata fetched",
		"video_id", id,
	)

	return resp.Data, nil
}

func (c *SingleClient) GetSubtitle(
	ctx context.Context,
	videoID,
	language,
	subType,
	format,
	outputPath string,
) ([]byte, error) {
	c.logger.Debug(
		"fetching subtitle",
		"video_id", videoID,
		"language", language,
		"type", subType,
		"format", format,
		"output_path", outputPath,
	)

	req := SubtitleRequest{
		Cmd:        "subtitle",
		VideoID:    videoID,
		Language:   language,
		Type:       subType,
		Format:     format,
		OutputPath: outputPath,
	}

	if err := c.worker.SendCommand(req); err != nil {
		return nil, fmt.Errorf(
			"send subtitle request for video %s: %w",
			videoID,
			err,
		)
	}

	respData, err := c.worker.ReceiveResponse()
	if err != nil {
		return nil, fmt.Errorf(
			"receive subtitle response for video %s: %w",
			videoID,
			err,
		)
	}

	var resp SubtitleResponse

	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf(
			"decode subtitle response for video %s: %w",
			videoID,
			err,
		)
	}

	if !resp.Ok {
		c.logger.Warn(
			"subtitle request failed",
			"video_id", videoID,
			"language", language,
			"type", subType,
			"format", format,
			"error", resp.Err,
		)

		return nil, fmt.Errorf(
			"worker subtitle error for video %s: %s",
			videoID,
			resp.Err,
		)
	}

	c.logger.Debug(
		"subtitle fetched",
		"video_id", videoID,
		"language", language,
		"format", format,
	)

	return nil, nil // Return nil, caller reads the output path
}

func (c *SingleClient) WarmUp(ctx context.Context) error {
	if err := c.worker.SendCommand(WarmupRequest{Cmd: "warmup"}); err != nil {
		return err
	}
	respData, err := c.worker.ReceiveResponse()
	if err != nil {
		return err
	}
	var resp WorkerResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return err
	}
	if !resp.Ok {
		return fmt.Errorf("worker warmup error: %s", resp.Err)
	}
	return nil
}

func (c *SingleClient) Close() error {
	return c.worker.Close()
}
