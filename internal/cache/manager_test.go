package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Arvind215271/askito/internal/logger"
)

func TestManager_Exists(t *testing.T) {
	// Setup
	cacheDir, err := os.MkdirTemp("", "cache-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cacheDir)

	cfg := Config{
		CacheDir: cacheDir,
		TTLDays:  1,
		MaxFiles: 10,
	}
	m := NewManager(cfg, &logger.Logger{})

	id := "test-id"
	filename := "test.json"
	path := m.GetPath(id, filename)
	os.MkdirAll(filepath.Dir(path), 0755)

	// Case 1: File does not exist
	if m.Exists(id, filename) {
		t.Error("Expected false for non-existent file")
	}

	// Create file
	data := []byte("hello")
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Case 2: File exists, not expired
	if !m.Exists(id, filename) {
		t.Error("Expected true for fresh file")
	}

	// Case 3: File exists, expired
	expiredTime := time.Now().Add(-2 * 24 * time.Hour)
	err = os.Chtimes(path, expiredTime, expiredTime)
	if err != nil {
		t.Fatal(err)
	}

	if m.Exists(id, filename) {
		t.Error("Expected false for expired file")
	}
}

func TestManager_HasMetadata(t *testing.T) {
	// Setup
	cacheDir, err := os.MkdirTemp("", "cache-test-metadata")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cacheDir)

	cfg := Config{
		CacheDir: cacheDir,
		TTLDays:  1,
		MaxFiles: 10,
	}
	m := NewManager(cfg, &logger.Logger{})

	id := "test-id"
	filename := m.VideoKey()
	path := m.GetPath(id, filename)
	os.MkdirAll(filepath.Dir(path), 0755)

	// Case 1: Metadata does not exist
	if m.HasMetadata(id) {
		t.Error("Expected false for non-existent metadata")
	}

	// Create file
	data := []byte("{}")
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Case 2: Metadata exists, not expired
	if !m.HasMetadata(id) {
		t.Error("Expected true for fresh metadata")
	}

	// Case 3: Metadata exists, expired
	expiredTime := time.Now().Add(-2 * 24 * time.Hour)
	err = os.Chtimes(path, expiredTime, expiredTime)
	if err != nil {
		t.Fatal(err)
	}

	if m.HasMetadata(id) {
		t.Error("Expected false for expired metadata")
	}
}
