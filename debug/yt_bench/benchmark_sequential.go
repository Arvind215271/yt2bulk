package yt_bench

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata/ytdlp"
)

func BenchmarkSequential(ctx context.Context, playlistID string, outputDir string) (BenchmarkResult, error) {
	start := time.Now()
	
	// Ensure log directory exists
	logDir := "debug/yt_bench/output/logs"
	os.MkdirAll(logDir, 0755)
	logFile, err := os.Create(filepath.Join(logDir, "sequential.log"))
	if err != nil {
		return BenchmarkResult{}, err
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	flatStart := time.Now()
	flatOutput, err := FetchPlaylistFlatMetadata(ctx, playlistID)
	flatFetchTime := time.Since(flatStart)
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("flat fetch failed: %w", err)
	}

	videoIDs := []string{}
	scanner := bufio.NewScanner(bytes.NewReader(flatOutput))
	for scanner.Scan() {
		var entry ytdlp.YTPlaylistEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			videoIDs = append(videoIDs, entry.ID)
		}
	}
	fmt.Printf("DEBUG: Found %d videos in flat playlist\n", len(videoIDs))

	var totalMetadataFetchTime time.Duration
	var totalDiskWriteTime time.Duration

	for _, id := range videoIDs {
		msg := fmt.Sprintf("START videoID=%s\n", id)
		fmt.Print(msg)
		logger.Print(msg)
		
		fetchStart := time.Now()
		data, _ := FetchFullVideoMetadata(ctx, id)
		fetchDuration := time.Since(fetchStart)
		totalMetadataFetchTime += fetchDuration

		writeStart := time.Now()
		os.WriteFile(filepath.Join(outputDir, fmt.Sprintf("A_%s.json", id)), data, 0644)
		totalDiskWriteTime += time.Since(writeStart)

		msgEnd := fmt.Sprintf("END videoID=%s duration=%v\n", id, fetchDuration)
		fmt.Print(msgEnd)
		logger.Print(msgEnd)
	}

	res := BenchmarkResult{
		PlaylistID:             playlistID,
		Strategy:               "Sequential",
		PlaylistSize:           len(videoIDs),
		TotalExecutionTime:     time.Since(start),
		StartedAt:              start,
		ProcessesSpawned:       1 + len(videoIDs),
		FlatPlaylistFetchTime:  flatFetchTime,
		VideoMetadataFetchTime: totalMetadataFetchTime,
		DiskWriteTime:          totalDiskWriteTime,
	}
	SaveResult(res)
	return res, nil
}
