package yt_bench

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Arvind215271/yt2bulk/internal/cache"
	"github.com/Arvind215271/yt2bulk/internal/logger"
)

func RunBenchmarks(playlistIDs []string) {
	ctx := context.Background()
	outputDir := "debug/yt_bench/output"
	// os.RemoveAll(outputDir)
	os.MkdirAll(outputDir, 0755)

	cacheMgr := cache.NewManager(cache.Config{
		CacheDir: filepath.Join(".", ".cache", "ytdlp"),
		TTLDays:  7,
		MaxFiles: 100,
	}, logger.New("development"))
	loggerInstance := log.New(os.Stdout, "", log.LstdFlags)

	for _, playlistID := range playlistIDs {
		fmt.Printf("Starting Benchmarks for: %s\n", playlistID)

		workers := []int{16, 32, 64, 128, 256}

		for _, w := range workers {
			resD, err := BenchmarkPythonPoolWarm(ctx, playlistID, outputDir, w, cacheMgr, loggerInstance)
			if err != nil {
				log.Fatalf("Strategy D.2 (%d) failed: %v", w, err)
			}
			fmt.Printf("Strategy D.2 (Python Pool Warm %d) Time: %v\n", w, resD.TotalBenchmarkDuration)
		}
	}
}

func SaveResult(res BenchmarkResult) {
	dir := "docs"
	os.MkdirAll(dir, 0755)
	historyFile := filepath.Join(dir, "benchmark_history.json")

	var history []BenchmarkResult
	if data, err := os.ReadFile(historyFile); err == nil {
		json.Unmarshal(data, &history)
	}
	history = append(history, res)
	data, _ := json.MarshalIndent(history, "", "  ")
	err := os.WriteFile(historyFile, data, 0644)
	if err != nil {
		fmt.Printf("ERROR: Failed to save result to %s: %v\n", historyFile, err)
		return
	}
	fmt.Printf("Saved result for %s strategy: %s\n", res.Strategy, res.PlaylistID)
}
