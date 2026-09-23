package usecase

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
)

func TestTranscribeAllRejectsOutputNameCollisionsBeforeStarting(t *testing.T) {
	inputDir := "recordings"
	paths := []string{}
	for _, name := range []string{"same.mp3", "same.wav"} {
		paths = append(paths, filepath.Join(inputDir, name))
	}

	var calls atomic.Int32
	_, err := runBatch(context.Background(), portin.Request{
		InputPath: inputDir,
		Format:    "txt",
		Parallel:  2,
	}, paths, func(_ context.Context, _ portin.Request) (string, error) {
		calls.Add(1)
		return "", nil
	})
	if err == nil {
		t.Fatal("runBatch() error = nil, want output collision error")
	}
	if !strings.Contains(err.Error(), "same.txt") {
		t.Errorf("runBatch() error = %q, want conflicting output name", err)
	}
	if calls.Load() != 0 {
		t.Errorf("transcriber calls = %d, want 0", calls.Load())
	}
}

func TestTranscribeAllKeepsStartingFilesUntilDirectoryIsFinished(t *testing.T) {
	inputDir := "recordings"
	paths := []string{}
	for _, name := range []string{"a.mp3", "b.mp3", "c.mp3"} {
		paths = append(paths, filepath.Join(inputDir, name))
	}

	started := make(chan string, 3)
	release := make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })

	resultChannel := make(chan []portin.Result, 1)
	go func() {
		results, _ := runBatch(context.Background(), portin.Request{
			InputPath: inputDir,
			Parallel:  2,
		}, paths, func(_ context.Context, options portin.Request) (string, error) {
			started <- options.InputPath
			<-release
			return options.InputPath + ".txt", nil
		})
		resultChannel <- results
	}()

	for index := 0; index < 2; index++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("parallel workers did not start")
		}
	}
	select {
	case path := <-started:
		t.Fatalf("third file %q started before a worker became available", path)
	case <-time.After(50 * time.Millisecond):
	}

	releaseOnce.Do(func() { close(release) })
	select {
	case results := <-resultChannel:
		if len(results) != 3 {
			t.Errorf("len(results) = %d, want 3", len(results))
		}
	case <-time.After(time.Second):
		t.Fatal("directory processing did not finish")
	}
}

func TestTranscribeAllContinuesAfterOneFileFails(t *testing.T) {
	inputDir := "recordings"
	paths := []string{}
	for _, name := range []string{"a.mp3", "b.mp3", "c.mp3"} {
		paths = append(paths, filepath.Join(inputDir, name))
	}

	processed := make([]string, 0, 3)
	results, err := runBatch(context.Background(), portin.Request{
		InputPath: inputDir,
		Parallel:  1,
	}, paths, func(_ context.Context, options portin.Request) (string, error) {
		processed = append(processed, filepath.Base(options.InputPath))
		if filepath.Base(options.InputPath) == "b.mp3" {
			return "", errors.New("decode failed")
		}
		return options.InputPath + ".txt", nil
	})
	if err != nil {
		t.Fatalf("runBatch() error = %v", err)
	}
	if !reflect.DeepEqual(processed, []string{"a.mp3", "b.mp3", "c.mp3"}) {
		t.Errorf("processed files = %q, want all files", processed)
	}
	if len(results) != 3 || results[1].Err == nil {
		t.Errorf("results = %+v, want only middle file to fail", results)
	}
}

func TestTranscribeAllCancelsActiveWorkWithoutStartingQueuedFiles(t *testing.T) {
	inputDir := "recordings"
	paths := []string{}
	for _, name := range []string{"a.mp3", "b.mp3", "c.mp3"} {
		paths = append(paths, filepath.Join(inputDir, name))
	}

	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	var calls atomic.Int32
	resultChannel := make(chan []portin.Result, 1)
	go func() {
		results, _ := runBatch(ctx, portin.Request{
			InputPath: inputDir,
			Parallel:  1,
		}, paths, func(ctx context.Context, _ portin.Request) (string, error) {
			calls.Add(1)
			close(started)
			<-ctx.Done()
			return "", ctx.Err()
		})
		resultChannel <- results
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("active transcription did not start")
	}
	cancel()

	select {
	case results := <-resultChannel:
		if calls.Load() != 1 {
			t.Errorf("transcriber calls = %d, want only active file", calls.Load())
		}
		if len(results) != 3 {
			t.Fatalf("len(results) = %d, want 3 cancellation results", len(results))
		}
		for _, result := range results {
			if !errors.Is(result.Err, context.Canceled) {
				t.Errorf("result for %q error = %v, want context cancellation", result.InputPath, result.Err)
			}
		}
	case <-time.After(time.Second):
		t.Fatal("runBatch() did not finish after cancellation")
	}
}
