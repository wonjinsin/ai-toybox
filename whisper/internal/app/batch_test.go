package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDiscoverInputPathsReturnsSupportedFilesInNameOrder(t *testing.T) {
	t.Parallel()

	inputDir := t.TempDir()
	for _, name := range []string{"voice.amr", "audio.mka", "track.ac3", "b.wav", "a.mp4", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(inputDir, name), []byte("input"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(inputDir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "nested", "ignored.mp3"), []byte("input"), 0o644); err != nil {
		t.Fatal(err)
	}

	paths, err := discoverInputPaths(inputDir)
	if err != nil {
		t.Fatalf("discoverInputPaths() error = %v", err)
	}

	want := []string{
		filepath.Join(inputDir, "a.mp4"),
		filepath.Join(inputDir, "audio.mka"),
		filepath.Join(inputDir, "b.wav"),
		filepath.Join(inputDir, "track.ac3"),
		filepath.Join(inputDir, "voice.amr"),
	}
	if !reflect.DeepEqual(paths, want) {
		t.Errorf("discoverInputPaths() = %q, want %q", paths, want)
	}
}

func TestTranscribeAllRejectsOutputNameCollisionsBeforeStarting(t *testing.T) {
	inputDir := t.TempDir()
	for _, name := range []string{"same.mp3", "same.wav"} {
		if err := os.WriteFile(filepath.Join(inputDir, name), []byte("input"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	var calls atomic.Int32
	_, err := transcribeAll(context.Background(), Options{
		InputPath: inputDir,
		Format:    "txt",
		Parallel:  2,
	}, func(_ context.Context, _ Options) (string, error) {
		calls.Add(1)
		return "", nil
	})
	if err == nil {
		t.Fatal("transcribeAll() error = nil, want output collision error")
	}
	if !strings.Contains(err.Error(), "same.txt") {
		t.Errorf("transcribeAll() error = %q, want conflicting output name", err)
	}
	if calls.Load() != 0 {
		t.Errorf("transcriber calls = %d, want 0", calls.Load())
	}
}

func TestDiscoverInputPathsRejectsDirectoryWithoutMedia(t *testing.T) {
	t.Parallel()

	_, err := discoverInputPaths(t.TempDir())
	if err == nil {
		t.Fatal("discoverInputPaths() error = nil, want empty directory error")
	}
	if !strings.Contains(err.Error(), "no supported media files") {
		t.Errorf("discoverInputPaths() error = %q, want supported media context", err)
	}
}

func TestDiscoverInputPathsKeepsSingleFileInput(t *testing.T) {
	t.Parallel()

	inputPath := filepath.Join(t.TempDir(), "recording.custom")
	if err := os.WriteFile(inputPath, []byte("input"), 0o644); err != nil {
		t.Fatal(err)
	}

	paths, err := discoverInputPaths(inputPath)
	if err != nil {
		t.Fatalf("discoverInputPaths() error = %v", err)
	}
	if !reflect.DeepEqual(paths, []string{inputPath}) {
		t.Errorf("discoverInputPaths() = %q, want single input %q", paths, inputPath)
	}
}

func TestTranscribeAllKeepsStartingFilesUntilDirectoryIsFinished(t *testing.T) {
	inputDir := t.TempDir()
	for _, name := range []string{"a.mp3", "b.mp3", "c.mp3"} {
		if err := os.WriteFile(filepath.Join(inputDir, name), []byte("input"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	started := make(chan string, 3)
	release := make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })

	resultChannel := make(chan []TranscriptionResult, 1)
	go func() {
		results, _ := transcribeAll(context.Background(), Options{
			InputPath: inputDir,
			Parallel:  2,
		}, func(_ context.Context, options Options) (string, error) {
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
	inputDir := t.TempDir()
	for _, name := range []string{"a.mp3", "b.mp3", "c.mp3"} {
		if err := os.WriteFile(filepath.Join(inputDir, name), []byte("input"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	processed := make([]string, 0, 3)
	results, err := transcribeAll(context.Background(), Options{
		InputPath: inputDir,
		Parallel:  1,
	}, func(_ context.Context, options Options) (string, error) {
		processed = append(processed, filepath.Base(options.InputPath))
		if filepath.Base(options.InputPath) == "b.mp3" {
			return "", errors.New("decode failed")
		}
		return options.InputPath + ".txt", nil
	})
	if err != nil {
		t.Fatalf("transcribeAll() error = %v", err)
	}
	if !reflect.DeepEqual(processed, []string{"a.mp3", "b.mp3", "c.mp3"}) {
		t.Errorf("processed files = %q, want all files", processed)
	}
	if len(results) != 3 || results[1].Err == nil {
		t.Errorf("results = %+v, want only middle file to fail", results)
	}
}

func TestTranscribeAllCancelsActiveWorkWithoutStartingQueuedFiles(t *testing.T) {
	inputDir := t.TempDir()
	for _, name := range []string{"a.mp3", "b.mp3", "c.mp3"} {
		if err := os.WriteFile(filepath.Join(inputDir, name), []byte("input"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	var calls atomic.Int32
	resultChannel := make(chan []TranscriptionResult, 1)
	go func() {
		results, _ := transcribeAll(ctx, Options{
			InputPath: inputDir,
			Parallel:  1,
		}, func(ctx context.Context, _ Options) (string, error) {
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
		t.Fatal("transcribeAll() did not finish after cancellation")
	}
}
