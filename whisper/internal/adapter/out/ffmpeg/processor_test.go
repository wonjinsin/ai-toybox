package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
)

func TestSegmentAudioForVADRejectsNonFiniteManifestInterval(t *testing.T) {
	tempDir := t.TempDir()
	ffmpegPath := filepath.Join(tempDir, "ffmpeg")
	writeExecutable(t, ffmpegPath, `#!/bin/sh
set -eu
manifest=""
previous=""
for argument do
  if [ "$previous" = "-segment_list" ]; then manifest="$argument"; fi
  previous="$argument"
done
directory=${manifest%/*}
: > "$directory/vad_00000.wav"
printf '%s\n' 'vad_00000.wav,NaN,15.000000' > "$manifest"
`)

	_, err := SegmentForVAD(context.Background(), ffmpegPath, "audio.wav", tempDir)
	if err == nil {
		t.Fatal("SegmentForVAD() error = nil, want non-finite interval error")
	}
}

func TestExtractTranscriptionChunksUsesFourWorkers(t *testing.T) {
	chunks := make([]domain.Chunk, 5)
	for index := range chunks {
		chunks[index] = domain.Chunk{Start: time.Duration(index) * time.Second, End: time.Duration(index+1) * time.Second}
	}

	started := make(chan struct{}, len(chunks))
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseAll := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseAll()
	var active atomic.Int32
	var maximum atomic.Int32
	directory := t.TempDir()
	execute := func(ctx context.Context, _ string, _ []string) error {
		current := active.Add(1)
		for current > maximum.Load() && !maximum.CompareAndSwap(maximum.Load(), current) {
		}
		started <- struct{}{}
		select {
		case <-release:
		case <-ctx.Done():
			active.Add(-1)
			return ctx.Err()
		}
		active.Add(-1)
		return nil
	}

	result := make(chan error, 1)
	go func() {
		_, err := extractTranscriptionChunksFromIndexUsing(context.Background(), execute, "ffmpeg", "audio.wav", directory, chunks, 0)
		result <- err
	}()

	for worker := 0; worker < 4; worker++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatalf("started workers = %d, want 4", worker)
		}
	}
	if got := maximum.Load(); got != 4 {
		t.Errorf("maximum concurrent extractions = %d, want 4", got)
	}
	releaseAll()
	if err := <-result; err != nil {
		t.Fatalf("extractTranscriptionChunksFromIndexUsing() error = %v", err)
	}
}

func TestExtractTranscriptionChunksPreservesInputOrder(t *testing.T) {
	chunks := []domain.Chunk{
		{Start: 5 * time.Second, End: 6 * time.Second},
		{Start: 10 * time.Second, End: 11 * time.Second},
		{Start: 15 * time.Second, End: 16 * time.Second},
	}
	execute := func(_ context.Context, _ string, arguments []string) error {
		outputPath := arguments[len(arguments)-1]
		switch {
		case strings.Contains(outputPath, "chunk_0007"):
			time.Sleep(30 * time.Millisecond)
		case strings.Contains(outputPath, "chunk_0008"):
			time.Sleep(20 * time.Millisecond)
		default:
			time.Sleep(10 * time.Millisecond)
		}
		return nil
	}

	got, err := extractTranscriptionChunksFromIndexUsing(context.Background(), execute, "ffmpeg", "audio.wav", t.TempDir(), chunks, 7)
	if err != nil {
		t.Fatalf("extractTranscriptionChunksFromIndexUsing() error = %v", err)
	}
	for index, chunk := range got {
		if chunk.Chunk != chunks[index] {
			t.Errorf("chunk %d interval = %#v, want %#v", index, chunk.Chunk, chunks[index])
		}
		wantPrefix := fmt.Sprintf("chunk_%04d_", index+7)
		if !strings.HasPrefix(filepath.Base(chunk.Path), wantPrefix) {
			t.Errorf("chunk %d path = %q, want prefix %q", index, chunk.Path, wantPrefix)
		}
	}
}

func TestExtractTranscriptionChunksCancelsSiblingWorkersAfterFailure(t *testing.T) {
	chunks := make([]domain.Chunk, 6)
	for index := range chunks {
		chunks[index] = domain.Chunk{Start: time.Duration(index) * time.Second, End: time.Duration(index+1) * time.Second}
	}

	wantError := errors.New("ffmpeg failed")
	allWorkersStarted := make(chan struct{})
	var started atomic.Int32
	var canceled atomic.Int32
	execute := func(ctx context.Context, _ string, arguments []string) error {
		if started.Add(1) == 4 {
			close(allWorkersStarted)
		}
		<-allWorkersStarted
		outputPath := arguments[len(arguments)-1]
		if strings.Contains(outputPath, "chunk_0009") {
			return wantError
		}
		<-ctx.Done()
		canceled.Add(1)
		return ctx.Err()
	}

	_, err := extractTranscriptionChunksFromIndexUsing(context.Background(), execute, "ffmpeg", "audio.wav", t.TempDir(), chunks, 9)
	if !errors.Is(err, wantError) {
		t.Fatalf("extractTranscriptionChunksFromIndexUsing() error = %v, want %v", err, wantError)
	}
	if !strings.Contains(err.Error(), "extract speech chunk 9") {
		t.Errorf("error = %q, want failing chunk index", err)
	}
	if canceled.Load() == 0 {
		t.Error("sibling workers were not canceled")
	}
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func TestProcessorExtractIsolatesOutputFilesBetweenBatches(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "ffmpeg")
	writeExecutable(t, executable, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\n: > \"$output\"\n")
	processor := New(executable, "ffprobe", "input", "audio.wav", directory)
	chunks := []domain.Chunk{{Start: time.Second, End: 2 * time.Second}}
	first, err := processor.Extract(context.Background(), chunks, 7)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(first.Chunks[0].Path+".json", []byte(`{"transcription":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := processor.Extract(context.Background(), chunks, 7)
	if err != nil {
		t.Fatal(err)
	}
	if first.Chunks[0].Path == second.Chunks[0].Path {
		t.Fatal("repeated extraction reused a previous batch's output path")
	}
	if _, err := os.Stat(second.Chunks[0].Path + ".json"); !os.IsNotExist(err) {
		t.Fatalf("new batch has stale JSON: %v", err)
	}
	if filepath.Base(first.Chunks[0].Path) != filepath.Base(second.Chunks[0].Path) || second.Chunks[0].Index != 7 {
		t.Fatalf("chunk filename/index contract changed: %#v, %#v", first, second)
	}
}

func TestProcessorExtractEmptyBatchDoesNotAllocateDirectory(t *testing.T) {
	directory := t.TempDir()
	processor := New("unused-ffmpeg", "ffprobe", "input", "audio.wav", directory)
	if _, err := processor.Extract(context.Background(), nil, 0); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("empty extraction created %d filesystem entries, want none", len(entries))
	}
}

func TestProcessorExtractReleaseRemovesOnlyOwnedBatch(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "ffmpeg")
	writeExecutable(t, executable, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\n: > \"$output\"\n")
	source := filepath.Join(directory, "source.mp4")
	normalized := filepath.Join(directory, "normalized.wav")
	for _, path := range []string{source, normalized} {
		if err := os.WriteFile(path, []byte("audio"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	processor := New(executable, "ffprobe", source, normalized, directory)
	chunks := []domain.Chunk{{Start: time.Second, End: 2 * time.Second}}
	first, err := processor.Extract(context.Background(), chunks, 7)
	if err != nil {
		t.Fatal(err)
	}
	second, err := processor.Extract(context.Background(), chunks, 7)
	if err != nil {
		t.Fatal(err)
	}
	if first.Release == nil || second.Release == nil {
		t.Fatal("allocated audio batches must provide a release callback")
	}
	t.Cleanup(func() {
		if err := second.Release(); err != nil {
			t.Error(err)
		}
	})
	for _, path := range []string{first.Chunks[0].Path, second.Chunks[0].Path} {
		if err := os.WriteFile(path+".json", []byte(`{"transcription":[]}`), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("extracted WAV %q is unavailable before release: %v", path, err)
		}
	}
	if err := first.Release(); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{first.Chunks[0].Path, first.Chunks[0].Path + ".json", filepath.Dir(first.Chunks[0].Path)} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("released batch path %q still exists: %v", path, err)
		}
	}
	for _, path := range []string{directory, source, normalized, executable, second.Chunks[0].Path, second.Chunks[0].Path + ".json"} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("release removed unrelated path %q: %v", path, err)
		}
	}
	if err := first.Release(); err != nil {
		t.Fatalf("repeated release must be harmless: %v", err)
	}
}

func TestProcessorExtractFailureRemovesPartialBatch(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "ffmpeg")
	writeExecutable(t, executable, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\n: > \"$output\"\n: > \"$output.json\"\nprintf 'extraction failed' >&2\nexit 7\n")
	processor := New(executable, "ffprobe", "input", "audio.wav", directory)
	_, err := processor.Extract(context.Background(), []domain.Chunk{{Start: time.Second, End: 2 * time.Second}}, 7)
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 7 || !strings.Contains(err.Error(), "extract speech chunk 7") {
		t.Fatalf("extraction error = %v, want wrapped chunk 7 exit error", err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "ffmpeg" {
		t.Fatalf("failed extraction left batch resources: %v", entries)
	}
}

func TestProcessorExtractCancellationLeavesNoBatchResources(t *testing.T) {
	for _, test := range []struct {
		name   string
		chunks []domain.Chunk
	}{
		{name: "empty"},
		{name: "nonempty", chunks: []domain.Chunk{{End: time.Second}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			processor := New("unused-ffmpeg", "ffprobe", "input", "audio.wav", directory)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if _, err := processor.Extract(ctx, test.chunks, 0); !errors.Is(err, context.Canceled) {
				t.Fatalf("canceled extraction error = %v, want context.Canceled", err)
			}
			entries, err := os.ReadDir(directory)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatalf("canceled extraction left %d filesystem entries, want none", len(entries))
			}
		})
	}
}

func TestProcessorNormalizesAudioAndProbesMedia(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "ffmpeg")
	normalized := filepath.Join(directory, "normalized.wav")
	source := filepath.Join(directory, "source input.mp4")
	if err := os.WriteFile(source, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, executable, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\nprintf '%s\\n' \"$@\" > \"$output\"\n")
	probe := filepath.Join(directory, "ffprobe")
	writeExecutable(t, probe, "#!/bin/sh\nset -eu\nfor argument do input=\"$argument\"; done\ntest -f \"$input\"\nprintf '12.345\\n'\n")
	processor := New(executable, probe, source, normalized, directory)
	if err := processor.Normalize(context.Background()); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(normalized)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-nostdin", "-hide_banner", "-loglevel", "error", "-y", "-i", source, "-vn", "-af", "dynaudnorm=f=500:g=31:p=0.95:m=100:b=1", "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", normalized}
	if string(content) != strings.Join(want, "\n")+"\n" {
		t.Fatalf("normalization arguments = %q, want %q", content, want)
	}
	duration, err := processor.Duration(context.Background())
	if err != nil || duration != 12345*time.Millisecond {
		t.Fatalf("duration = %v, error = %v", duration, err)
	}
}

func TestProcessorReportsCommandAndDurationErrors(t *testing.T) {
	for _, test := range []struct{ name, script, want string }{
		{"invalid duration", "printf 'unknown\\n'", "parse media duration"},
		{"zero duration", "printf '0\\n'", "parse media duration"},
		{"negative duration", "printf '%s\\n' '-1'", "parse media duration"},
		{"probe stderr", "printf 'probe failed' >&2; exit 7", "probe failed"},
		{"probe failure", "exit 7", "probe media duration"},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			executable := filepath.Join(directory, "ffprobe")
			writeExecutable(t, executable, "#!/bin/sh\n"+test.script+"\n")
			processor := New(executable, executable, "input", "audio", directory)
			if _, err := processor.Duration(context.Background()); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("duration error = %v, want %q", err, test.want)
			}
		})
	}
	directory := t.TempDir()
	executable := filepath.Join(directory, "ffmpeg")
	writeExecutable(t, executable, "#!/bin/sh\nprintf 'audio failed' >&2\nexit 7\n")
	processor := New(executable, "probe", "input", "audio", directory)
	if err := processor.Normalize(context.Background()); err == nil || !strings.Contains(err.Error(), "extract audio with ffmpeg") || !strings.Contains(err.Error(), "audio failed") {
		t.Fatalf("normalization error = %v", err)
	}
}

func TestSegmentForVADReadsBoundedManifestIntervals(t *testing.T) {
	directory := t.TempDir()
	executable := filepath.Join(directory, "ffmpeg")
	writeExecutable(t, executable, "#!/bin/sh\nexit 0\n")
	if err := os.WriteFile(filepath.Join(directory, "vad-chunks.csv"), []byte("vad_00000.wav,0.000,15.000\nvad_00001.wav,15.000,22.250\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"vad_00000.wav", "vad_00001.wav"} {
		if err := os.WriteFile(filepath.Join(directory, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	chunks, err := SegmentForVAD(context.Background(), executable, "audio", directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 2 || chunks[1].Start != 15*time.Second || chunks[1].End != 22250*time.Millisecond || chunks[1].Path != filepath.Join(directory, "vad_00001.wav") {
		t.Fatalf("chunks = %#v", chunks)
	}
}

func TestSegmentForVADRejectsUnsafeOrMalformedManifest(t *testing.T) {
	for _, test := range []struct{ name, manifest, want string }{
		{"empty", "", "no audio chunks created"},
		{"field count", "vad.wav,0\n", "got 2 fields"},
		{"invalid csv", "\"unclosed,0,1\n", "read VAD chunk manifest"},
		{"path traversal", "../outside.wav,0,1\n", "unsafe chunk path"},
		{"absolute path", "/tmp/outside.wav,0,1\n", "unsafe chunk path"},
		{"reversed interval", "vad.wav,2,1\n", "invalid interval"},
		{"missing chunk", "vad.wav,0,1\n", "open VAD audio chunk file"},
		{"directory chunk", "folder,0,1\n", "not a regular file"},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			executable := filepath.Join(directory, "ffmpeg")
			writeExecutable(t, executable, "#!/bin/sh\nexit 0\n")
			if err := os.Mkdir(filepath.Join(directory, "folder"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "vad-chunks.csv"), []byte(test.manifest), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := SegmentForVAD(context.Background(), executable, "audio", directory)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestProcessorDurationUsesOriginalMediaTimelineWhenAudioIsShorter(t *testing.T) {
	directory := t.TempDir()
	probe := filepath.Join(directory, "ffprobe")
	writeExecutable(t, probe, `#!/bin/sh
set -eu
for argument do input="$argument"; done
case "$input" in
 */source.mp4) printf '25.000\n' ;;
 *) printf '10.000\n' ;;
esac
`)
	processor := New("ffmpeg", probe, filepath.Join(directory, "source.mp4"), filepath.Join(directory, "audio.wav"), directory)
	duration, err := processor.Duration(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if duration != 25*time.Second {
		t.Fatalf("media duration = %v, want 25s original video duration", duration)
	}
}
