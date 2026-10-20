package python

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Arvind215271/yt2bulk/internal/cache"
	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube/stats"
)

type WorkerExecution struct {
	VideoID    string        `json:"video_id"`
	StartedAt  time.Time     `json:"started_at"`
	FinishedAt time.Time     `json:"finished_at"`
	Duration   time.Duration `json:"duration"`
}

type SinglePool struct {
	manager *WorkerManager
	cache   *cache.Manager
	once    sync.Once
}

func NewSinglePool(
	workerCount int,
	log *logger.Logger,
	cache *cache.Manager,
) (*SinglePool, error) {
	manager, err := NewWorkerManager(
		workerCount,
		log,
	)
	if err != nil {
		return nil, err
	}

	return &SinglePool{
		manager: manager,
		cache:   cache,
	}, nil
}

func (p *SinglePool) WarmUp(ctx context.Context) error {
	return p.manager.WarmUp(ctx)
}

func (p *SinglePool) GetVideo(
	ctx context.Context,
	videoID string,
	st *stats.MetadataStats,
) (map[string]any, error) {
	if st != nil {
		st.YTDLPRequests++
	}
	key := p.cache.VideoKey()

	data, err := p.cache.Get(videoID, key)
	if err == nil {
		var result map[string]any

		if err := json.Unmarshal(data, &result); err == nil {
			if st != nil {
				st.CacheHits++
			}
			return result, nil
		}
	}

	if st != nil {
		st.CacheMisses++
		st.UpstreamFetches++
	}

	result, err := p.manager.GetVideo(ctx, videoID)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(result)
	_ = p.cache.Save(videoID, key, jsonData)

	return result, nil
}

func (p *SinglePool) GetVideos(
	ctx context.Context,
	videoIDs []string,
) ([]map[string]any, error) {
	results := make([]map[string]any, len(videoIDs))

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for i, videoID := range videoIDs {
		wg.Add(1)

		go func(index int, id string) {
			defer wg.Done()

			result, err := p.GetVideo(ctx, id, nil)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}

			results[index] = result
		}(i, videoID)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
		return results, nil
	}
}

func (p *SinglePool) GetPlaylist(
	ctx context.Context,
	playlistID string,
	st *stats.MetadataStats,
) (map[string]any, error) {
	if st != nil {
		st.YTDLPRequests++
	}
	key := p.cache.PlaylistKey()

	data, err := p.cache.Get(playlistID, key)
	if err == nil {
		var result map[string]any

		if err := json.Unmarshal(data, &result); err == nil {
			if st != nil {
				st.CacheHits++
			}
			return result, nil
		}
	}

	if st != nil {
		st.CacheMisses++
		st.UpstreamFetches++
	}

	result, err := p.manager.GetPlaylist(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(result)
	_ = p.cache.Save(playlistID, key, jsonData)

	return result, nil
}

func (p *SinglePool) GetPlaylistMetadata(
	ctx context.Context,
	playlistID string,
	st *stats.MetadataStats,
) (map[string]any, error) {
	if st != nil {
		st.YTDLPRequests++
	}
	key := p.cache.PlaylistMetadataKey()

	data, err := p.cache.Get(playlistID, key)
	if err == nil {
		var result map[string]any

		if err := json.Unmarshal(data, &result); err == nil {
			if st != nil {
				st.CacheHits++
			}
			return result, nil
		}
	}

	if st != nil {
		st.CacheMisses++
		st.UpstreamFetches++
	}

	result, err := p.manager.GetPlaylistMetadata(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(result)
	_ = p.cache.Save(playlistID, key, jsonData)

	return result, nil
}

func (p *SinglePool) GetPlaylistSurfaceMetadata(
	ctx context.Context,
	playlistID string,
	st *stats.MetadataStats,
) (map[string]any, error) {
	if st != nil {
		st.YTDLPRequests++
	}
	key := p.cache.PlaylistSurfaceMetadataKey()

	data, err := p.cache.Get(playlistID, key)
	if err == nil {
		var result map[string]any

		if err := json.Unmarshal(data, &result); err == nil {
			if st != nil {
				st.CacheHits++
			}
			return result, nil
		}
	}

	if st != nil {
		st.CacheMisses++
		st.UpstreamFetches++
	}

	result, err := p.manager.GetPlaylistSurfaceMetadata(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	jsonData, _ := json.Marshal(result)
	_ = p.cache.Save(playlistID, key, jsonData)

	return result, nil
}

func (p *SinglePool) GetSubtitle(
	ctx context.Context,
	videoID,
	language,
	subType,
	format string,
) ([]byte, error) {
	key := p.cache.SubtitleKey(subType, language, format)

	outputPath := p.cache.GetPath(
		videoID,
		p.cache.SubtitlePath(subType),
	)

	content, err := p.cache.Get(videoID, key)
	if err == nil {
		return content, nil
	}

	_, err = p.manager.GetSubtitle(
		ctx,
		videoID,
		language,
		subType,
		format,
		outputPath,
	)
	if err != nil {
		return nil, err
	}

	data, err := p.cache.Get(videoID, key)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *SinglePool) GetSubtitles(
	ctx context.Context,
	videoIDs []string,
	language,
	subType,
	format string,
) ([][]byte, error) {
	results := make([][]byte, len(videoIDs))

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for i, videoID := range videoIDs {
		wg.Add(1)

		go func(index int, id string) {
			defer wg.Done()

			result, err := p.GetSubtitle(
				ctx,
				id,
				language,
				subType,
				format,
			)
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}

			results[index] = result
		}(i, videoID)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
		return results, nil
	}
}

func (p *SinglePool) Close() {
	p.once.Do(func() {
		_ = p.manager.Close()
	})
}
