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
	"sync/atomic"
	"time"

	"github.com/Arvind215271/yt2bulk/internal/cache"
	"github.com/Arvind215271/yt2bulk/internal/logger"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata/ytdlp"
	"github.com/Arvind215271/yt2bulk/internal/youtube/metadata/ytdlp/python"
)

func BenchmarkPythonPool(ctx context.Context, playlistID string, outputDir string, workers int, cacheMgr *cache.Manager, loggerInstance *log.Logger) (BenchmarkResult, error) {
	benchmarkStart := time.Now()

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

	// Flat playlist fetch timing
	flatFetchStart := time.Now()
	flatOutput, err := FetchPlaylistFlatMetadata(ctx, playlistID)
	flatFetchEnd := time.Now()
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

	var aggregateWorkerTime int64 // nanoseconds
	var diskWriteTime int64       // nanoseconds

	// Atomic values for concurrent timing updates
	var firstWorkerStartedAt atomic.Value // time.Time
	var lastWorkerFinishedAt atomic.Value // time.Time
	var firstDiskWriteAt atomic.Value     // time.Time
	var lastDiskWriteAt atomic.Value      // time.Time

	// Worker spawn timing
	workerSpawnStart := time.Now()
	// Use logger from internal/logger package
	l := logger.New("development")
	pool, err := python.NewSinglePool(workers, l, cacheMgr)
	workerSpawnEnd := time.Now()
	if err != nil {
		return BenchmarkResult{}, err
	}
	defer pool.Close()

	// Metadata dispatch timing
	dispatchStart := time.Now()

	for _, e := range flat.Entries {
		loggerInstance.Printf("Submitting videoID=%s to worker", e.ID)
		data, err := pool.GetVideo(ctx, e.ID, nil)
		if err != nil {
			loggerInstance.Printf("Error processing video %s: %v", e.ID, err)
			continue
		}

		// Disk write timing
		writeStart := time.Now()
		fileName := filepath.Join(outputDir, fmt.Sprintf("D_%d_%s.json", workers, e.ID))
		respData, _ := json.Marshal(data)
		os.WriteFile(fileName, respData, 0644)
		writeDuration := time.Since(writeStart)
		atomic.AddInt64(&diskWriteTime, int64(writeDuration))

		// Update first/last disk write times
		for {
			old := firstDiskWriteAt.Load()
			if old == nil || writeStart.Before(old.(time.Time)) {
				if firstDiskWriteAt.CompareAndSwap(old, writeStart) {
					break
				}
			} else {
				break
			}
		}
		for {
			old := lastDiskWriteAt.Load()
			if old == nil || writeStart.Add(writeDuration).After(old.(time.Time)) {
				if lastDiskWriteAt.CompareAndSwap(old, writeStart.Add(writeDuration)) {
					break
				}
			} else {
				break
			}
		}

		loggerInstance.Printf("Processed videoID=%s, dataLen=%d", e.ID, len(respData))
	}

	dispatchEnd := time.Now()

	// Wait for all workers to complete
	pool.Close()

	benchmarkEnd := time.Now()

	// Calculate derived durations
	firstWorkerStart := time.Time{}
	if v := firstWorkerStartedAt.Load(); v != nil {
		firstWorkerStart = v.(time.Time)
	}
	lastWorkerEnd := time.Time{}
	if v := lastWorkerFinishedAt.Load(); v != nil {
		lastWorkerEnd = v.(time.Time)
	}
	firstDiskWrite := time.Time{}
	if v := firstDiskWriteAt.Load(); v != nil {
		firstDiskWrite = v.(time.Time)
	}
	lastDiskWrite := time.Time{}
	if v := lastDiskWriteAt.Load(); v != nil {
		lastDiskWrite = v.(time.Time)
	}

	actualMetadataFetchDuration := time.Duration(0)
	if !firstWorkerStart.IsZero() && !lastWorkerEnd.IsZero() {
		actualMetadataFetchDuration = lastWorkerEnd.Sub(firstWorkerStart)
	}

	diskWriteDuration := time.Duration(0)
	if !firstDiskWrite.IsZero() && !lastDiskWrite.IsZero() {
		diskWriteDuration = lastDiskWrite.Sub(firstDiskWrite)
	}

	res := BenchmarkResult{
		PlaylistID:             playlistID,
		Strategy:               fmt.Sprintf("Python Pool (%d)", workers),
		PlaylistSize:           len(flat.Entries),
		TotalExecutionTime:     benchmarkEnd.Sub(benchmarkStart),
		StartedAt:              benchmarkStart,
		ProcessesSpawned:       1 + workers + len(flat.Entries),
		PeakMemoryUsage:        atomic.LoadUint64(&peakMemory),
		FlatPlaylistFetchTime:  flatFetchEnd.Sub(flatFetchStart),
		VideoMetadataFetchTime: time.Duration(atomic.LoadInt64(&aggregateWorkerTime)), // Deprecated: aggregate time
		DiskWriteTime:          time.Duration(atomic.LoadInt64(&diskWriteTime)),
		WorkerSpawnTime:        workerSpawnEnd.Sub(workerSpawnStart),

		// New detailed fields
		BenchmarkStartedAt:         benchmarkStart,
		BenchmarkFinishedAt:        benchmarkEnd,
		FlatFetchStartedAt:         flatFetchStart,
		FlatFetchFinishedAt:        flatFetchEnd,
		WorkerSpawnStartedAt:       workerSpawnStart,
		WorkerSpawnFinishedAt:      workerSpawnEnd,
		MetadataDispatchStartedAt:  dispatchStart,
		MetadataDispatchFinishedAt: dispatchEnd,
		FirstWorkerStartedAt:       firstWorkerStart,
		LastWorkerFinishedAt:       lastWorkerEnd,
		DiskWriteStartedAt:         firstDiskWrite,
		DiskWriteFinishedAt:        lastDiskWrite,

		// Derived durations
		FlatFetchDuration:           flatFetchEnd.Sub(flatFetchStart),
		WorkerSpawnDuration:         workerSpawnEnd.Sub(workerSpawnStart),
		DispatchDuration:            dispatchEnd.Sub(dispatchStart),
		ActualMetadataFetchDuration: actualMetadataFetchDuration,
		DiskWriteDuration:           diskWriteDuration,
		TotalBenchmarkDuration:      benchmarkEnd.Sub(benchmarkStart),
		AggregateWorkerTime:         time.Duration(atomic.LoadInt64(&aggregateWorkerTime)),
	}
	SaveResult(res)
	return res, nil
}
