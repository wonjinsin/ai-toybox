package app

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type commandExecutor func(context.Context, string, []string) (string, error)

type whisperCommandRunner struct {
	limiter *commandLimiter
	execute commandExecutor
}

func newWhisperCommandRunner(maxConcurrent int, execute commandExecutor) *whisperCommandRunner {
	return &whisperCommandRunner{
		limiter: newCommandLimiter(maxConcurrent),
		execute: execute,
	}
}

func (runner *whisperCommandRunner) Run(ctx context.Context, executable string, args []string) (string, error) {
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
