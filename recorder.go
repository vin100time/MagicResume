package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type RecorderState int

const (
	StateIdle RecorderState = iota
	StateRecording
	StatePaused
)

type Recorder struct {
	mu        sync.Mutex
	state     RecorderState
	cmd       *exec.Cmd
	audioPath string
	startTime time.Time
	storage   *Storage
	ctx       context.Context
	stopTimer chan struct{}
}

func NewRecorder(storage *Storage, ctx context.Context) *Recorder {
	return &Recorder{
		state:   StateIdle,
		storage: storage,
		ctx:     ctx,
	}
}

func (r *Recorder) IsRecording() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state != StateIdle
}

func (r *Recorder) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state != StateIdle {
		return fmt.Errorf("déjà en enregistrement")
	}

	timestamp := time.Now().Format("2006-01-02_15h04")
	r.audioPath = filepath.Join(r.storage.audioDir, fmt.Sprintf("rec_%s.wav", timestamp))

	r.cmd = exec.Command("rec", "-r", "44100", "-c", "1", r.audioPath)
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Suppress sox output
	r.cmd.Stdout = nil
	r.cmd.Stderr = nil

	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("impossible de lancer rec: %w", err)
	}

	r.state = StateRecording
	r.startTime = time.Now()

	// Start timer goroutine
	r.stopTimer = make(chan struct{})
	go r.emitTimer()

	return nil
}

func (r *Recorder) emitTimer() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.mu.Lock()
			if r.state == StateIdle {
				r.mu.Unlock()
				return
			}
			elapsed := time.Since(r.startTime)
			r.mu.Unlock()

			runtime.EventsEmit(r.ctx, "recording_timer", int(elapsed.Seconds()))
		case <-r.stopTimer:
			return
		}
	}
}

func (r *Recorder) Pause() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state != StateRecording {
		return fmt.Errorf("pas en enregistrement")
	}

	if r.cmd != nil && r.cmd.Process != nil {
		r.cmd.Process.Signal(syscall.SIGSTOP)
	}
	r.state = StatePaused
	return nil
}

func (r *Recorder) Resume() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state != StatePaused {
		return fmt.Errorf("pas en pause")
	}

	if r.cmd != nil && r.cmd.Process != nil {
		r.cmd.Process.Signal(syscall.SIGCONT)
	}
	r.state = StateRecording
	return nil
}

func (r *Recorder) Stop() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state == StateIdle {
		return "", fmt.Errorf("pas en enregistrement")
	}

	// If paused, resume first so the process can receive SIGINT
	if r.state == StatePaused && r.cmd != nil && r.cmd.Process != nil {
		r.cmd.Process.Signal(syscall.SIGCONT)
	}

	// Stop timer
	if r.stopTimer != nil {
		close(r.stopTimer)
	}

	// Send SIGINT to stop rec gracefully
	if r.cmd != nil && r.cmd.Process != nil {
		r.cmd.Process.Signal(syscall.SIGINT)
		r.cmd.Wait()
	}

	r.state = StateIdle
	path := r.audioPath

	// Verify file exists and has content
	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 {
		os.Remove(path)
		return "", fmt.Errorf("aucun audio enregistré")
	}

	return path, nil
}
