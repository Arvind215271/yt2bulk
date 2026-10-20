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
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata/ytdlp"
)

func BenchmarkParallel(ctx context.Context, playlistID string, outputDir string, workers int) (BenchmarkResult, error) {
	start := time.Now()
	
	// Memory tracking
	var peakMemory uint64
	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for range ticker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			if m.Alloc > peakMemory {
				atomic.StoreUint64(&peakMemory, m.Alloc)
			}
		}
	}()
	defer ticker.Stop()

	// Ensure log directory exists
	logDir := "debug/yt_bench/output/logs"
	os.MkdirAll(logDir, 0755)
	logFile, err := os.Create(filepath.Join(logDir, fmt.Sprintf("parallel_workers_%d.log", workers)))
	if err != nil {
		return BenchmarkResult{}, err
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	flatStart := time.Now()
	flatOutput, err := FetchPlaylistFlatMetadata(ctx, playlistID)
	flatFetchTime := time.Since(flatStart)
	if err != nil {
		return BenchmarkResult{}, err
	}
	var flat ytdlp.YTPlaylistOutput
	scanner := bufio.NewScanner(bytes.NewReader(flatOutput))
	for scanner.Scan() {
		var entry ytdlp.YTPlaylistEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			flat.Entries = append(flat.Entries, entry)
		}
	}

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var activeWorkers int32
	var metadataFetchTime int64 // nanoseconds
	var diskWriteTime int64     // nanoseconds
	errChan := make(chan error, 1)

	for _, e := range flat.Entries {
		wg.Add(1)
		sem <- struct{}{}
		go func(id string) {
			defer wg.Done()
			defer func() { <-sem }()
			
			current := atomic.AddInt32(&activeWorkers, 1)
			if int(current) > workers {
				select {
				case errChan <- fmt.Errorf("active workers %d exceeded limit %d", current, workers):
				default:
				}
				atomic.AddInt32(&activeWorkers, -1)
				return
			}

			msg := fmt.Sprintf("START worker videoID=%s active=%d\n", id, current)
			fmt.Print(msg)
			logger.Print(msg)
			
			fetchStart := time.Now()
			data, err := FetchFullVideoMetadata(ctx, id)
			fetchDuration := time.Since(fetchStart)
			atomic.AddInt64(&metadataFetchTime, int64(fetchDuration))

			if err != nil {
				fmt.Printf("ERROR: FetchFullVideoMetadata failed for %s: %v\n", id, err)
				atomic.AddInt32(&activeWorkers, -1)
				return
			}
			
			writeStart := time.Now()
			fileName := filepath.Join(outputDir, fmt.Sprintf("C_%d_%s.json", workers, id))
			err = os.WriteFile(fileName, data, 0644)
			writeDuration := time.Since(writeStart)
			atomic.AddInt64(&diskWriteTime, int64(writeDuration))
			
			if err != nil {
				fmt.Printf("ERROR: os.WriteFile failed for %s: %v\n", id, err)
			}

			atomic.AddInt32(&activeWorkers, -1)
			msgEnd := fmt.Sprintf("END worker videoID=%s active=%d duration=%v\n", id, atomic.LoadInt32(&activeWorkers), fetchDuration)
			fmt.Print(msgEnd)
			logger.Print(msgEnd)
		}(e.ID)
	}
	wg.Wait()

	select {
	case err := <-errChan:
		return BenchmarkResult{}, err
	default:
	}

	res := BenchmarkResult{
		PlaylistID:             playlistID,
		Strategy:               fmt.Sprintf("Parallel (%d)", workers),
		PlaylistSize:           len(flat.Entries),
		TotalExecutionTime:     time.Since(start),
		StartedAt:              start,
		ProcessesSpawned:       1 + len(flat.Entries),
		PeakMemoryUsage:        atomic.LoadUint64(&peakMemory),
		FlatPlaylistFetchTime:  flatFetchTime,
		VideoMetadataFetchTime: time.Duration(atomic.LoadInt64(&metadataFetchTime)),
		DiskWriteTime:          time.Duration(atomic.LoadInt64(&diskWriteTime)),
	}
	SaveResult(res)
	return res, nil
}
