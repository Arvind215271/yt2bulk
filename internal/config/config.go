package config

import (
	"os"
	"strconv"

	"github.com/Arvind215271/yt2bulk/internal/cache"
	"github.com/joho/godotenv"
)

type Config struct {
	Env  string
	Port string

	YouTubeAPIKey string
	YtdlpCache    cache.Config
	PythonWorkers int
}

// Load reads environment variables once and builds the config
func Load() Config {
	// load the env file into the system, so it can be used by os
	_ = godotenv.Load()

	return Config{
		Env:  getEnv("APP_ENV", "development"),
		Port: getEnv("PORT", "8080"),

		YouTubeAPIKey: getEnv("YOUTUBE_API_KEY", ""),

		YtdlpCache: cache.Config{
			CacheDir: getEnv("YTDLP_CACHE_DIR", "./.cache/ytdlp"),
			TTLDays:  getEnvAsInt("YTDLP_CACHE_TTL_DAYS", 28),
			MaxFiles: getEnvAsInt("YTDLP_CACHE_MAX_FILES", 2000), // Updated to 2000
		},

		PythonWorkers: getEnvAsInt("PYTHON_WORKERS", 16),
	}
}

// getEnv returns fallback if env is not set
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

// getEnvAsInt returns fallback if env is not set or not an int
func getEnvAsInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	if i, err := strconv.Atoi(val); err == nil {
		return i
	}
	return fallback
}
