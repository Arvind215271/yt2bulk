package cache

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Arvind215271/yt2bulk/internal/logger"
)

type Config struct {
	CacheDir        string
	TTLDays         int
	MaxFiles        int
	CleanupInterval time.Duration
}

type Manager struct {
	config Config
	logger *logger.Logger
}

func NewManager(cfg Config, logger *logger.Logger) *Manager {
	return &Manager{config: cfg, logger: logger}
}

func (m *Manager) safeID(id string) string {
	// Sanitize the ID to be filesystem-friendly.
	// If the ID is a full URL, attempt to extract the identifier (e.g. handle or channel ID).
	if strings.HasPrefix(id, "http") {
		// Try to extract the last part of the path
		parts := strings.Split(strings.TrimRight(id, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	// Replace characters that are invalid or problematic in paths with underscores.
	replacer := strings.NewReplacer(":", "_", "/", "_", "?", "_", "@", "_", "&", "_", "=", "_")
	return replacer.Replace(id)
}

func (m *Manager) GetPath(id, filename string) string {
	return filepath.Join(m.config.CacheDir, m.safeID(id), filename)
}

func (m *Manager) Exists(id, filename string) bool {
	path := m.GetPath(id, filename)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) <= time.Duration(m.config.TTLDays)*24*time.Hour
}

func (m *Manager) Get(id, filename string) ([]byte, error) {
	path := m.GetPath(id, filename)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if time.Since(info.ModTime()) > time.Duration(m.config.TTLDays)*24*time.Hour {
		return nil, os.ErrNotExist
	}
	return os.ReadFile(path)
}

func (m *Manager) Save(id, filename string, data []byte) error {
	path := m.GetPath(id, filename)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	now := time.Now()
	return os.Chtimes(dir, now, now)
}

func (m *Manager) VideoKey() string {
	return "metadata.json"
}

// HasMetadata checks if valid metadata exists for a given video ID.
func (m *Manager) HasMetadata(id string) bool {
	return m.Exists(id, m.VideoKey())
}

// HasSubtitle checks if any subtitle file exists for a given video ID.
func (m *Manager) HasSubtitle(id string) bool {
	dir := filepath.Join(m.config.CacheDir, m.safeID(id))

	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	ttl := time.Duration(m.config.TTLDays) * 24 * time.Hour

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "subtitles.") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if time.Since(info.ModTime()) <= ttl {
			return true
		}
	}

	return false
}

func (m *Manager) PlaylistKey() string {
	return "playlist.json"
}

func (m *Manager) ChannelKey() string {
	return "channel.json"
}

func (m *Manager) PlaylistMetadataKey() string {
	return "playlist_metadata.json"
}

func (m *Manager) GetPlaylistVideoIDs(playlistID string) ([]string, bool) {
	data, err := m.Get(playlistID, m.PlaylistMetadataKey())
	if err != nil {
		return nil, false
	}

	var result struct {
		VideoIDs []string `json:"video_ids"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}
	return result.VideoIDs, true
}

func (m *Manager) PlaylistSurfaceMetadataKey() string {
	return "playlist_surface_metadata.json"
}

func (m *Manager) SubtitleKey(subType, language, format string) string {
	return "subtitles." + subType + "." + language + "." + format
}

func (m *Manager) SubtitlePath(subType string) string {
	return "subtitles." + subType
}

// Cleanup removes expired and excess cache entries.
//
// Each direct child of the cache directory is treated as one cache entry.
// Cache entries may be either files or directories. Expiration and capacity
// are determined using the modification time of the entry itself.
func (m *Manager) Cleanup() error {
	m.logger.Info("starting cache cleanup", "cache_dir", m.config.CacheDir)
	entries, err := os.ReadDir(m.config.CacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			m.logger.Info("cache directory does not exist, skipping cleanup", "cache_dir", m.config.CacheDir)
			return nil
		}

		m.logger.Error("failed to read cache directory during cleanup", "err", err, "cache_dir", m.config.CacheDir)
		return err
	}

	type cacheEntry struct {
		path    string
		modTime time.Time
	}

	var entriesToKeep []cacheEntry

	ttl := time.Duration(
		m.config.TTLDays,
	) * 24 * time.Hour

	now := time.Now()

	for _, entry := range entries {

		path := filepath.Join(
			m.config.CacheDir,
			entry.Name(),
		)

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// TTL-based eviction.
		if now.Sub(info.ModTime()) > ttl {
			m.logger.Info("evicting expired cache entry", "path", path, "mod_time", info.ModTime(), "age", now.Sub(info.ModTime()))
			if err := os.RemoveAll(path); err != nil {
				m.logger.Error("failed to remove expired cache entry", "path", path, "err", err)
				return err
			}

			continue
		}

		entriesToKeep = append(
			entriesToKeep,
			cacheEntry{
				path:    path,
				modTime: info.ModTime(),
			},
		)
	}

	// Capacity-based eviction.
	if len(entriesToKeep) <= m.config.MaxFiles {
		return nil
	}

	sort.Slice(
		entriesToKeep,
		func(i, j int) bool {
			return entriesToKeep[i].modTime.Before(
				entriesToKeep[j].modTime,
			)
		},
	)

	// Keep 90% of the configured capacity after cleanup.
	targetSize := m.config.MaxFiles * 9 / 10

	numToDelete := len(entriesToKeep) - targetSize

	for i := 0; i < numToDelete; i++ {

		if err := os.RemoveAll(
			entriesToKeep[i].path,
		); err != nil {
			return err
		}
	}

	return nil
}

// StartCleanupLoop runs Cleanup periodically.
func (m *Manager) StartCleanupLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := m.Cleanup(); err != nil {
					m.logger.Error("cache cleanup failed", "error", err)
				} else {
					m.logger.Info("periodic cache cleanup completed")
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
