package python

import (
	"context"
	"errors"
	"sync"

	"github.com/Arvind215271/yt2bulk/internal/logger"
)

// ErrManagerClosed signals the manager is no longer accepting requests.
var ErrManagerClosed = errors.New("worker manager closed")

type request struct {
	execute  func(*SingleClient) (any, error)
	response chan ManagerResponse
}

type ManagerResponse struct {
	Result interface{}
	Err    error
}

type WorkerManager struct {
	mu      sync.Mutex
	workers []*SingleClient
	queue   chan request

	logger *logger.Logger

	closing bool
	wg      sync.WaitGroup
}

// WarmUp initializes all workers.
func (m *WorkerManager) WarmUp(ctx context.Context) error {
	var wg sync.WaitGroup
	errCh := make(chan error, len(m.workers))
	for _, w := range m.workers {
		wg.Add(1)
		go func(worker *SingleClient) {
			defer wg.Done()
			if err := worker.WarmUp(ctx); err != nil {
				errCh <- err
			}
		}(w)
	}
	wg.Wait()
	close(errCh)
	if len(errCh) > 0 {
		return <-errCh
	}
	return nil
}

func (m *WorkerManager) replaceWorker(index int) (*SingleClient, error) {
	// Only the dedicated workerLoop for this worker calls replaceWorker,
	// therefore replacements for a worker are serialized.
	m.mu.Lock()
	if m.closing {
		m.mu.Unlock()
		return nil, ErrWorkerDied
	}

	oldWorker := m.workers[index]
	m.mu.Unlock()

	m.logger.Warn(
		"replacing dead python worker",
		"worker_id", index,
	)

	// oldWorker.Close() errors are ignored because the worker
	// has already died and we are replacing it.
	_ = oldWorker.Close()

	newWorker, err := NewSingleWorker(
		index,
		m.logger,
	)
	if err != nil {
		m.logger.Error(
			"failed to replace python worker",
			"worker_id", index,
			"error", err,
		)
		return nil, err
	}

	m.mu.Lock()
	if m.closing {
		m.mu.Unlock()
		_ = newWorker.Close()
		return nil, ErrWorkerDied
	}
	m.workers[index] = newWorker
	m.mu.Unlock()

	m.logger.Info(
		"python worker replaced",
		"worker_id", index,
	)

	return newWorker, nil
}

// workerLoop processes requests from the manager queue.
//
// 1. Send request.
// 2. If worker returns a normal response, return it.
// 3. If communication with Python fails (ErrWorkerDied), replace worker.
// 4. Retry exactly once.
// 5. Return whatever retry returns.
func (m *WorkerManager) workerLoop(index int, worker *SingleClient) {
	currWorker := worker
	for req := range m.queue {
		result, err := req.execute(currWorker)

		if err != nil && errors.Is(err, ErrWorkerDied) {
			m.logger.Warn(
				"python worker died during request",
				"worker_id", index,
				"error", err,
			)

			newWorker, rErr := m.replaceWorker(index)
			if rErr == nil {
				currWorker = newWorker
				result, err = req.execute(currWorker)
			} else {
				err = rErr
			}
		}

		req.response <- ManagerResponse{Result: result, Err: err}
	}
}

func NewWorkerManager(
	workerCount int,
	log *logger.Logger,
) (*WorkerManager, error) {
	if workerCount <= 0 {
		return nil, errors.New("worker count must be greater than zero")
	}

	manager := &WorkerManager{
		workers: make([]*SingleClient, 0, workerCount),
		queue:   make(chan request),
		logger: log.With(
			"component", "python_worker_manager",
		),
	}

	manager.logger.Info(
		"starting python worker manager",
		"worker_count", workerCount,
	)

	for i := 0; i < workerCount; i++ {
		worker, err := NewSingleWorker(
			i,
			manager.logger,
		)
		if err != nil {
			manager.logger.Error(
				"failed to create python worker",
				"worker_id", i,
				"error", err,
			)
			// Cleanup
			for _, w := range manager.workers {
				_ = w.Close()
			}
			return nil, err
		}
		manager.workers = append(manager.workers, worker)
		manager.wg.Add(1)
		go func(i int, w *SingleClient) {
			defer manager.wg.Done()
			manager.workerLoop(i, w)
		}(i, worker)
	}

	manager.logger.Info(
		"python worker manager started",
		"worker_count", workerCount,
	)

	return manager, nil
}

func (m *WorkerManager) GetVideo(ctx context.Context, videoID string) (map[string]any, error) {
	respCh := make(chan ManagerResponse, 1)
	req := request{
		execute: func(w *SingleClient) (any, error) {
			return w.GetVideo(ctx, videoID)
		},
		response: respCh,
	}

	m.mu.Lock()
	if m.closing {
		m.mu.Unlock()
		return nil, ErrManagerClosed
	}
	m.mu.Unlock()

	select {
	case m.queue <- req:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	res := <-respCh
	if res.Err != nil {
		return nil, res.Err
	}
	return res.Result.(map[string]any), nil
}

func (m *WorkerManager) GetSubtitle(ctx context.Context, videoID, language, subType, format, outputPath string) ([]byte, error) {
	respCh := make(chan ManagerResponse, 1)
	req := request{
		execute: func(w *SingleClient) (any, error) {
			return w.GetSubtitle(ctx, videoID, language, subType, format, outputPath)
		},
		response: respCh,
	}

	m.mu.Lock()
	if m.closing {
		m.mu.Unlock()
		return nil, ErrManagerClosed
	}
	m.mu.Unlock()

	select {
	case m.queue <- req:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	res := <-respCh
	if res.Err != nil {
		return nil, res.Err
	}
	// Note: result is nil now, but the worker writes the file to outputPath
	return nil, nil
}

func (m *WorkerManager) GetPlaylist(
	ctx context.Context,
	playlistID string,
) (map[string]any, error) {
	respCh := make(chan ManagerResponse, 1)

	req := request{
		execute: func(w *SingleClient) (any, error) {
			return w.GetPlaylist(ctx, playlistID)
		},
		response: respCh,
	}

	m.mu.Lock()
	if m.closing {
		m.mu.Unlock()
		return nil, ErrManagerClosed
	}
	m.mu.Unlock()

	select {
	case m.queue <- req:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	res := <-respCh

	if res.Err != nil {
		return nil, res.Err
	}

	return res.Result.(map[string]any), nil
}

func (m *WorkerManager) GetPlaylistMetadata(
	ctx context.Context,
	playlistID string,
) (map[string]any, error) {
	respCh := make(chan ManagerResponse, 1)

	req := request{
		execute: func(w *SingleClient) (any, error) {
			return w.GetPlaylistMetadata(ctx, playlistID)
		},
		response: respCh,
	}

	m.mu.Lock()
	if m.closing {
		m.mu.Unlock()
		return nil, ErrManagerClosed
	}
	m.mu.Unlock()

	select {
	case m.queue <- req:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	res := <-respCh

	if res.Err != nil {
		return nil, res.Err
	}

	return res.Result.(map[string]any), nil
}

func (m *WorkerManager) GetPlaylistSurfaceMetadata(
	ctx context.Context,
	playlistID string,
) (map[string]any, error) {
	respCh := make(chan ManagerResponse, 1)

	req := request{
		execute: func(w *SingleClient) (any, error) {
			return w.GetPlaylistSurfaceMetadata(ctx, playlistID)
		},
		response: respCh,
	}

	m.mu.Lock()
	if m.closing {
		m.mu.Unlock()
		return nil, ErrManagerClosed
	}
	m.mu.Unlock()

	select {
	case m.queue <- req:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	res := <-respCh

	if res.Err != nil {
		return nil, res.Err
	}

	return res.Result.(map[string]any), nil
}

func (m *WorkerManager) Close() error {
	m.mu.Lock()
	m.closing = true
	m.mu.Unlock()

	close(m.queue)
	m.wg.Wait()

	var firstErr error
	for _, w := range m.workers {
		err := w.Close()
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
