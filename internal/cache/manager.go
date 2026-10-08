package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Arvind215271/askito/internal/logger"
)

type Config struct {
	CacheDir string
	TTLDays  int
	MaxFiles int
}

type Manager struct {
	config Config
	logger *logger.Logger
}

func NewManager(cfg Config, logger *logger.Logger) *Manager {
	return &Manager{config: cfg, logger: logger}
}

func (m *Manager) GetPath(id, filename string) string {
	return filepath.Join(m.config.CacheDir, id, filename)
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
	if os.IsNotExist(err) {
		return nil, err
	}
	if time.Since(info.ModTime()) > time.Duration(m.config.TTLDays)*24*time.Hour {
		return nil, os.ErrNotExist
	}
	return os.ReadFile(path)
}

func (m *Manager) Save(id, filename string, data []byte) error {
	dir := filepath.Dir(m.GetPath(id, filename))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(m.GetPath(id, filename), data, 0644)
}

func (m *Manager) VideoKey() string {
	return "metadata." + ".json"
}

// HasMetadata checks if valid metadata exists for a given video ID.
func (m *Manager) HasMetadata(id string) bool {
	return m.Exists(id, m.VideoKey())
}

// HasSubtitle checks if any subtitle file exists for a given video ID.
func (m *Manager) HasSubtitle(id string) bool {
	dir := filepath.Join(m.config.CacheDir, id)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 10 && entry.Name()[0:10] == "subtitles." {
			return true
		}
	}
	return false
}

func (m *Manager) PlaylistKey() string {
	return "playlist." + ".json"
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
	entries, err := os.ReadDir(m.config.CacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

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
			if err := os.RemoveAll(path); err != nil {
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
