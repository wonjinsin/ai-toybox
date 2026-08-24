package app

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWhisperCommandRunnerCancelsWhileWaitingForSlot(t *testing.T) {
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	defer close(release)

	runner := newWhisperCommandRunner(1, func(ctx context.Context, _ string, _ []string) (string, error) {
		started <- struct{}{}
		select {
		case <-release:
			return "", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	})
	go runner.Run(context.Background(), "whisper-cli", nil)
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("first command did not acquire the slot")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := make(chan error, 1)
	go func() {
		_, err := runner.Run(ctx, "whisper-cli", nil)
		result <- err
	}()

	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run() error = %v, want context.Canceled", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Run() did not return after context cancellation")
	}
}

func TestWhisperCommandRunnerRetriesOnceAfterGPUOutOfMemory(t *testing.T) {
	var calls atomic.Int32
	runner := newWhisperCommandRunner(2, func(context.Context, string, []string) (string, error) {
		if calls.Add(1) == 1 {
			return "", errors.New("error: Insufficient Memory (00000008:kIOGPUCommandBufferCallbackErrorOutOfMemory)")
		}
		return "retry output", nil
	})

	output, err := runner.Run(context.Background(), "whisper-cli", []string{"-m", "model.bin"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if output != "retry output" {
		t.Errorf("Run() output = %q, want retry output", output)
	}
	if got := calls.Load(); got != 2 {
		t.Errorf("command calls = %d, want 2", got)
	}
}

func TestWhisperCommandRunnerWaitsForActiveCommandBeforeOOMRetry(t *testing.T) {
	holdStarted := make(chan struct{})
	holdRelease := make(chan struct{})
	oomFailed := make(chan struct{})
	retryStarted := make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(holdRelease) }) })
	var oomCalls atomic.Int32

	runner := newWhisperCommandRunner(2, func(_ context.Context, executable string, _ []string) (string, error) {
		switch executable {
		case "hold":
			close(holdStarted)
			<-holdRelease
			return "", nil
		case "oom":
			if oomCalls.Add(1) == 1 {
				close(oomFailed)
				return "", errors.New("kIOGPUCommandBufferCallbackErrorOutOfMemory")
			}
			close(retryStarted)
			return "retry output", nil
		default:
			return "", errors.New("unexpected executable")
		}
	})

	go runner.Run(context.Background(), "hold", nil)
	select {
	case <-holdStarted:
	case <-time.After(time.Second):
		t.Fatal("active command did not start")
	}
	result := make(chan error, 1)
	go func() {
		_, err := runner.Run(context.Background(), "oom", nil)
		result <- err
	}()
	select {
	case <-oomFailed:
	case <-time.After(time.Second):
		t.Fatal("OOM command did not fail")
	}
	select {
	case <-retryStarted:
		t.Fatal("OOM retry started while another command was active")
	case <-time.After(100 * time.Millisecond):
	}

	releaseOnce.Do(func() { close(holdRelease) })
	select {
	case <-retryStarted:
	case <-time.After(time.Second):
		t.Fatal("OOM retry did not start after active command completed")
	}
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("OOM command did not finish after retry")
	}
}

func TestWhisperCommandRunnerKeepsSingleCommandLimitAfterOOM(t *testing.T) {
	var calls atomic.Int32
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	runner := newWhisperCommandRunner(2, func(context.Context, string, []string) (string, error) {
		switch calls.Add(1) {
		case 1:
			return "", errors.New("OutOfMemory")
		case 2:
			return "retry output", nil
		default:
			started <- struct{}{}
			<-release
			return "", nil
		}
	})
	if _, err := runner.Run(context.Background(), "whisper-cli", nil); err != nil {
		t.Fatalf("initial Run() error = %v", err)
	}

	var workers sync.WaitGroup
	workers.Add(2)
	for range 2 {
		go func() {
			defer workers.Done()
			runner.Run(context.Background(), "whisper-cli", nil)
		}()
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("post-OOM command did not start")
	}
	select {
	case <-started:
		t.Fatal("two commands ran concurrently after OOM")
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	workers.Wait()
}

func TestWhisperCommandRunnerSerializesConcurrentOOMRetries(t *testing.T) {
	initialReady := make(chan struct{}, 2)
	releaseInitial := make(chan struct{})
	retryStarted := make(chan string, 2)
	releaseRetry := make(chan struct{})
	calls := make(map[string]int, 2)
	var callsMu sync.Mutex
	var activeRetries atomic.Int32
	var maximumRetries atomic.Int32

	runner := newWhisperCommandRunner(2, func(_ context.Context, executable string, _ []string) (string, error) {
		callsMu.Lock()
		calls[executable]++
		attempt := calls[executable]
		callsMu.Unlock()
		if attempt == 1 {
			initialReady <- struct{}{}
			<-releaseInitial
			return "", errors.New("kIOGPUCommandBufferCallbackErrorOutOfMemory")
		}

		active := activeRetries.Add(1)
		for {
			maximum := maximumRetries.Load()
			if active <= maximum || maximumRetries.CompareAndSwap(maximum, active) {
				break
			}
		}
		retryStarted <- executable
		<-releaseRetry
		activeRetries.Add(-1)
		return "", nil
	})

	results := make(chan error, 2)
	for _, executable := range []string{"first", "second"} {
		go func() {
			_, err := runner.Run(context.Background(), executable, nil)
			results <- err
		}()
	}
	for range 2 {
		select {
		case <-initialReady:
		case <-time.After(time.Second):
			t.Fatal("initial OOM commands did not start together")
		}
	}
	close(releaseInitial)
	select {
	case <-retryStarted:
	case <-time.After(time.Second):
		t.Fatal("first OOM retry did not start")
	}
	select {
	case second := <-retryStarted:
		t.Fatalf("second OOM retry %q started concurrently", second)
	case <-time.After(100 * time.Millisecond):
	}
	releaseRetry <- struct{}{}
	select {
	case <-retryStarted:
	case <-time.After(time.Second):
		t.Fatal("second OOM retry did not start serially")
	}
	releaseRetry <- struct{}{}
	for range 2 {
		select {
		case err := <-results:
			if err != nil {
				t.Errorf("Run() error = %v", err)
			}
		case <-time.After(time.Second):
			t.Fatal("OOM command did not finish")
		}
	}
	if got := maximumRetries.Load(); got != 1 {
		t.Errorf("maximum concurrent OOM retries = %d, want 1", got)
	}
}

func TestWhisperCommandRunnerCancelsOOMRetryWithoutLeakingSlot(t *testing.T) {
	holdStarted := make(chan struct{})
	holdRelease := make(chan struct{})
	oomFailed := make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(holdRelease) }) })
	var oomCalls atomic.Int32

	runner := newWhisperCommandRunner(2, func(_ context.Context, executable string, _ []string) (string, error) {
		switch executable {
		case "hold":
			close(holdStarted)
			<-holdRelease
			return "", nil
		case "oom":
			if oomCalls.Add(1) == 1 {
				close(oomFailed)
				return "", errors.New("Insufficient Memory")
			}
			return "", errors.New("retry should not start")
		case "after":
			return "available", nil
		default:
			return "", errors.New("unexpected executable")
		}
	})
	go runner.Run(context.Background(), "hold", nil)
	select {
	case <-holdStarted:
	case <-time.After(time.Second):
		t.Fatal("active command did not start")
	}

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		_, err := runner.Run(ctx, "oom", nil)
		result <- err
	}()
	select {
	case <-oomFailed:
	case <-time.After(time.Second):
		t.Fatal("OOM command did not fail")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("OOM retry wait did not honor context cancellation")
	}
	if got := oomCalls.Load(); got != 1 {
		t.Errorf("OOM command calls = %d, want 1", got)
	}

	releaseOnce.Do(func() { close(holdRelease) })
	output, err := runner.Run(context.Background(), "after", nil)
	if err != nil {
		t.Fatalf("post-cancellation Run() error = %v", err)
	}
	if output != "available" {
		t.Errorf("post-cancellation output = %q, want available", output)
	}
}

func TestWhisperCommandRunnerDoesNotRetryNormalError(t *testing.T) {
	wantErr := errors.New("invalid model")
	var calls atomic.Int32
	runner := newWhisperCommandRunner(2, func(context.Context, string, []string) (string, error) {
		calls.Add(1)
		return "", wantErr
	})

	_, err := runner.Run(context.Background(), "whisper-cli", nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("command calls = %d, want 1", got)
	}
}

func TestWhisperCommandRunnerReturnsSecondFailureDetails(t *testing.T) {
	var calls atomic.Int32
	runner := newWhisperCommandRunner(2, func(context.Context, string, []string) (string, error) {
		if calls.Add(1) == 1 {
			return "", errors.New("Insufficient Memory")
		}
		return "", errors.New("second OOM detail")
	})

	_, err := runner.Run(context.Background(), "whisper-cli", nil)
	if err == nil {
		t.Fatal("Run() error = nil, want retry failure")
	}
	for _, expected := range []string{"retry whisper command after GPU out of memory", "second OOM detail"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("Run() error = %q, want %q", err, expected)
		}
	}
	if got := calls.Load(); got != 2 {
		t.Errorf("command calls = %d, want 2", got)
	}
}

func TestWhisperOOMDetectionRequiresMemoryMarker(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		message string
		want    bool
	}{
		{message: "error: Insufficient Memory", want: true},
		{message: "kIOGPUCommandBufferCallbackErrorOutOfMemory", want: true},
		{message: "allocator: out of memory", want: true},
		{message: "whisper_full_with_state: failed to encode", want: false},
		{message: "invalid model", want: false},
	} {
		if got := isWhisperOutOfMemory(errors.New(test.message)); got != test.want {
			t.Errorf("isWhisperOutOfMemory(%q) = %t, want %t", test.message, got, test.want)
		}
	}
}
