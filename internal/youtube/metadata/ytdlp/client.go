package ytdlp

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata/ytdlp/python"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

type Client struct {
	pool   *python.SinglePool
	logger *logger.Logger
}

const (
	maxRetries = 3
	timeout    = 30 * time.Second
)

func NewClient(pool *python.SinglePool, logger *logger.Logger) *Client {
	return &Client{
		pool:   pool,
		logger: logger,
	}
}

// Cleanup runs the cache cleanup routine.
func (c *Client) Cleanup() error {
	return nil
}

// ValidateYTDLP checks if ytdlp is present.
func (c *Client) ValidateYTDLP() error {
	_, err := exec.LookPath("yt-dlp")
	return err
}

// Fetch fetches the request user asked from ytdlp...
func (c *Client) Fetch(ctx context.Context, args ...string) ([]byte, error) {
	var lastErr error

	c.logger.Debug("running yt-dlp", "args", args)

	for i := 0; i < maxRetries; i++ {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)

		cmd := exec.CommandContext(ctxWithTimeout, "yt-dlp", args...)
		output, err := cmd.CombinedOutput()

		cancel()

		if err == nil {
			c.logger.Debug("yt-dlp command succeeded")
			return output, nil
		}

		c.logger.Warn("yt-dlp attempt failed", "attempt", i+1, "error", err)
		lastErr = fmt.Errorf("%w\n%s", err, output)
	}

	c.logger.Error("yt-dlp command failed after retries", "error", lastErr)
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func (c *Client) GetVideo(ctx context.Context, videoID string, st *stats.MetadataStats) (YTOutput, error) {
	result, err := c.pool.GetVideo(ctx, videoID, st)
	if err != nil {
		return YTOutput{}, err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return YTOutput{}, err
	}

	var meta YTOutput
	err = json.Unmarshal(output, &meta)
	return meta, err
}

func (c *Client) GetPlaylist(ctx context.Context, playlistID string, st *stats.MetadataStats) (YTPlaylistOutput, error) {
	result, err := c.pool.GetPlaylist(ctx, playlistID, st)
	if err != nil {
		return YTPlaylistOutput{}, err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return YTPlaylistOutput{}, err
	}

	var meta YTPlaylistOutput
	err = json.Unmarshal(output, &meta)
	return meta, err
}

func (c *Client) GetPlaylistSurfaceMetadata(ctx context.Context, playlistID string, st *stats.MetadataStats) (PlaylistSurfaceMetadata, error) {
	result, err := c.pool.GetPlaylistSurfaceMetadata(ctx, playlistID, st)
	if err != nil {
		return PlaylistSurfaceMetadata{}, err
	}

	return MapToPlaylistSurfaceMetadata(result), nil
}

func (c *Client) GetPlaylistMetadata(ctx context.Context, playlistID string, st *stats.MetadataStats) (YTPlaylistOutput, error) {
	result, err := c.pool.GetPlaylistMetadata(ctx, playlistID, st)
	if err != nil {
		return YTPlaylistOutput{}, err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return YTPlaylistOutput{}, err
	}

	var meta YTPlaylistOutput
	err = json.Unmarshal(output, &meta)
	return meta, err
}

func (c *Client) GetChannel(ctx context.Context, channelURL string, st *stats.MetadataStats) (YTPlaylistOutput, error) {
	result, err := c.pool.GetChannel(ctx, channelURL, st)
	if err != nil {
		return YTPlaylistOutput{}, err
	}

	output, err := json.Marshal(result)
	if err != nil {
		return YTPlaylistOutput{}, err
	}

	var meta YTPlaylistOutput
	err = json.Unmarshal(output, &meta)
	return meta, err
}
