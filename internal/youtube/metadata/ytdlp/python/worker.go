package python

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/Arvind215271/yt2bulk/internal/logger"
)

var ErrWorkerDied = errors.New("python worker died")

type PythonWorker struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	logger   *logger.Logger
	workerID int
}

func NewWorker(
	scriptPath string,
	workerID int,
	log *logger.Logger,
) (*PythonWorker, error) {
	cmd := exec.Command(
		"./venv/bin/python3",
		"-u",
		scriptPath,
		"--worker-id",
		fmt.Sprintf("%d", workerID),
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create worker stdin: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create worker stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("create worker stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start python worker: %w", err)
	}

	workerLogger := log.With(
		"worker_id", workerID,
		"pid", cmd.Process.Pid,
	)

	workerLogger.Debug(
		"python worker started",
		"script", scriptPath,
	)

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			workerLogger.Warn("python worker stderr", "output", line)
		}
	}()

	return &PythonWorker{
		cmd:      cmd,
		stdin:    stdin,
		stdout:   stdout,
		stderr:   stderr,
		logger:   workerLogger,
		workerID: workerID,
	}, nil
}

func (w *PythonWorker) SendCommand(req any) error {
	reqData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal worker request: %w", err)
	}

	if err := binary.Write(
		w.stdin,
		binary.BigEndian,
		uint32(len(reqData)),
	); err != nil {
		w.logger.Warn(
			"failed to write worker request length",
			"error", err,
		)

		return fmt.Errorf("%w: write request length: %v", ErrWorkerDied, err)
	}

	if _, err := w.stdin.Write(reqData); err != nil {
		w.logger.Warn(
			"failed to write worker request",
			"error", err,
		)

		return fmt.Errorf("%w: write request: %v", ErrWorkerDied, err)
	}

	return nil
}

func (w *PythonWorker) ReceiveResponse() ([]byte, error) {
	var length uint32

	if err := binary.Read(
		w.stdout,
		binary.BigEndian,
		&length,
	); err != nil {
		w.logger.Warn(
			"failed to read worker response length",
			"error", err,
		)

		return nil, fmt.Errorf("%w: read response length: %v", ErrWorkerDied, err)
	}

	respData := make([]byte, length)

	if _, err := io.ReadFull(w.stdout, respData); err != nil {
		w.logger.Warn(
			"failed to read worker response",
			"response_length", length,
			"error", err,
		)

		return nil, fmt.Errorf("%w: read response: %v", ErrWorkerDied, err)
	}

	return respData, nil
}

func (w *PythonWorker) Close() error {
	w.logger.Debug("closing python worker")

	if err := w.stdin.Close(); err != nil {
		w.logger.Debug(
			"failed to close worker stdin",
			"error", err,
		)
	}

	if err := w.cmd.Wait(); err != nil {
		w.logger.Debug(
			"python worker exited with error",
			"error", err,
		)

		return err
	}

	w.logger.Debug("python worker closed")

	return nil
}
