package subtitle

import (
	"context"
	"fmt"

	"github.com/Arvind215271/yt2bulk/internal/cache"
	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

type SubtitleFetcher interface {
	GetSubtitle(ctx context.Context, videoID, language, subType, format string) ([]byte, error)
}

type SubtitleService struct {
	cache  *cache.Manager
	logger *logger.Logger
	pool   SubtitleFetcher
}

func NewSubtitleService(
	cache *cache.Manager,
	logger *logger.Logger,
	pool SubtitleFetcher,
) *SubtitleService {
	return &SubtitleService{
		cache:  cache,
		logger: logger,
		pool:   pool,
	}
}

// ResolveDownloadRequest evaluates preferences against metadata to produce a concrete DownloadRequest.
func (s *SubtitleService) ResolveDownloadRequest(
	meta SubtitleMetadata,
	preferences []SubtitlePreference,
	format string,
) (DownloadRequest, error) {
	selected, err := SelectSubtitle(&meta, preferences)
	if err != nil {
		return DownloadRequest{}, err
	}

	reqFormat := format
	if reqFormat == "" {
		reqFormat = selected.Format
	}

	return DownloadRequest{
		Language: selected.Language,
		Type:     selected.Type,
		Format:   reqFormat,
	}, nil
}

func (s *SubtitleService) DownloadSubtitleWithStats(
	ctx context.Context,
	req DownloadRequest,
	meta SubtitleMetadata,
) (*SubtitleResult, stats.SubtitleStats, error) {
	var st stats.SubtitleStats
	st.Attempts++

	s.logger.Debug(
		"subtitle download requested",
		"videoID", req.VideoID,
		"language", req.Language,
		"type", req.Type,
		"format", req.Format,
	)

	req, err := validateRequest(req)
	if err != nil {
		st.Failures++
		s.logger.Warn(
			"subtitle request validation failed",
			"videoID", req.VideoID,
			"language", req.Language,
			"type", req.Type,
			"format", req.Format,
			"error", err,
		)

		return nil, st, err
	}

	if req.Type == "manual" {
		st.ManualRequests++
	} else if req.Type == "automatic" {
		st.AutomaticRequests++
	}

	s.logger.Debug(
		"subtitle request validated",
		"videoID", req.VideoID,
		"language", req.Language,
		"type", req.Type,
		"format", req.Format,
	)

	if err := validateTrack(req, meta); err != nil {
		st.Failures++
		s.logger.Warn(
			"subtitle track validation failed",
			"videoID", req.VideoID,
			"language", req.Language,
			"type", req.Type,
			"format", req.Format,
			"error", err,
		)

		return nil, st, err
	}

	s.logger.Debug(
		"subtitle track validated",
		"videoID", req.VideoID,
		"language", req.Language,
		"type", req.Type,
		"format", req.Format,
	)

	st.UpstreamFetches++
	s.logger.Debug(
		"downloading subtitle",
		"videoID", req.VideoID,
		"language", req.Language,
		"type", req.Type,
		"format", req.Format,
	)

	data, err := s.pool.GetSubtitle(
		ctx,
		req.VideoID,
		req.Language,
		req.Type,
		req.Format,
	)
	if err != nil {
		st.Failures++
		s.logger.Warn(
			"subtitle download failed",
			"videoID", req.VideoID,
			"language", req.Language,
			"type", req.Type,
			"format", req.Format,
			"error", err,
		)

		return nil, st, err
	}

	s.logger.Debug(
		"subtitle downloaded",
		"videoID", req.VideoID,
		"language", req.Language,
		"type", req.Type,
		"format", req.Format,
		"bytes", len(data),
	)

	st.Successes++
	s.logger.Debug(
		"subtitle download completed",
		"videoID", req.VideoID,
		"language", req.Language,
		"type", req.Type,
		"format", req.Format,
		"bytes", len(data),
	)

	return &SubtitleResult{
		Content:  data,
		Format:   req.Format,
		Language: req.Language,
	}, st, nil
}

func (s *SubtitleService) DownloadSubtitle(
	ctx context.Context,
	req DownloadRequest,
	meta SubtitleMetadata,
) (*SubtitleResult, error) {
	res, _, err := s.DownloadSubtitleWithStats(ctx, req, meta)
	return res, err
}

func validateRequest(req DownloadRequest) (DownloadRequest, error) {

	if req.VideoID == "" {
		return req, fmt.Errorf("video ID is required")
	}

	if req.Language == "" {
		return req, fmt.Errorf("subtitle language is required")
	}

	switch req.Type {
	case "manual", "automatic":
	default:
		return req, fmt.Errorf(
			"invalid subtitle type: %s",
			req.Type,
		)
	}

	if req.Format == "" {
		req.Format = "json3"
	}

	switch req.Format {
	case "json3", "vtt", "srt", "srv1", "srv2", "srv3", "ttml":
	default:
		return req, fmt.Errorf(
			"unsupported subtitle format: %s",
			req.Format,
		)
	}

	return req, nil
}

func validateTrack(
	req DownloadRequest,
	meta SubtitleMetadata,
) error {

	var tracks []SubtitleTrack

	switch req.Type {
	case "manual":
		tracks = meta.Manual

	case "automatic":
		tracks = meta.Automatic
	}

	for _, track := range tracks {

		if track.LanguageCode != req.Language {
			continue
		}

		if !formatSupported(track, req.Format) {
			return fmt.Errorf(
				"subtitle format %s is not available for language %s",
				req.Format,
				req.Language,
			)
		}

		return nil
	}

	return fmt.Errorf(
		"subtitle track not found for type %s and language %s",
		req.Type,
		req.Language,
	)
}

func formatSupported(
	track SubtitleTrack,
	format string,
) bool {

	if len(track.Formats) == 0 {
		return true
	}

	for _, supported := range track.Formats {

		if supported == format {
			return true
		}
	}

	return false
}

func subtitleCacheKey(req DownloadRequest) string {
	return fmt.Sprintf(
		"subtitle.%s.%s",
		req.Language,
		req.Format,
	)
}
