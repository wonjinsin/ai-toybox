package whisper

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/out/process"
)

type CommandExecutor func(context.Context, string, []string) (string, error)

type CommandRunner struct {
	limiter *commandLimiter
	execute CommandExecutor
}

func NewCommandRunner(maxConcurrent int, execute CommandExecutor) *CommandRunner {
	if execute == nil {
		execute = process.RunCaptureStderr
	}
	return &CommandRunner{
		limiter: newCommandLimiter(maxConcurrent),
		execute: execute,
	}
}

func (runner *CommandRunner) Run(ctx context.Context, executable string, args []string) (string, error) {
	if err := runner.limiter.acquire(ctx); err != nil {
		return "", err
	}
	output, err := runner.execute(ctx, executable, args)
	if err == nil || !isWhisperOutOfMemory(err) {
		runner.limiter.release()
		return output, err
	}

	runner.limiter.reduceLimitToOne()
	runner.limiter.release()
	if err := runner.limiter.acquire(ctx); err != nil {
		return "", err
	}
	defer runner.limiter.release()
	output, err = runner.execute(ctx, executable, args)
	if err != nil {
		return "", fmt.Errorf("retry whisper command after GPU out of memory: %w", err)
	}
	return output, nil
}

func isWhisperOutOfMemory(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "insufficient memory") ||
		strings.Contains(message, "outofmemory") ||
		strings.Contains(message, "out of memory")
}

type commandLimiter struct {
	mu      sync.Mutex
	limit   int
	active  int
	changed chan struct{}
}

func newCommandLimiter(limit int) *commandLimiter {
	if limit < 1 {
		limit = 1
	}
	return &commandLimiter{
		limit:   limit,
		changed: make(chan struct{}),
	}
}

func (limiter *commandLimiter) acquire(ctx context.Context) error {
	for {
		limiter.mu.Lock()
		if limiter.active < limiter.limit {
			limiter.active++
			limiter.mu.Unlock()
			return nil
		}
		changed := limiter.changed
		limiter.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changed:
		}
	}
}

func (limiter *commandLimiter) release() {
	limiter.mu.Lock()
	limiter.active--
	limiter.notifyLocked()
	limiter.mu.Unlock()
}

func (limiter *commandLimiter) reduceLimitToOne() {
	limiter.mu.Lock()
	if limiter.limit > 1 {
		limiter.limit = 1
		limiter.notifyLocked()
	}
	limiter.mu.Unlock()
}

func (limiter *commandLimiter) notifyLocked() {
	close(limiter.changed)
	limiter.changed = make(chan struct{})
}

// Execute runs a command without consuming an inference slot.
func (runner *CommandRunner) Execute(ctx context.Context, executable string, args []string) (string, error) {
	return runner.execute(ctx, executable, args)
}
