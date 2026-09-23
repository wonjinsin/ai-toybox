package whisper

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

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

func TestDetectSpeechSegmentsReturnsCancellationAfterWorkerCompletes(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeFFmpeg(t, binDir, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\n: > \"$output\"\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	ctx, cancel := context.WithCancel(context.Background())
	execute := func(_ context.Context, _ string, _ []string) (string, error) {
		cancel()
		return "whisper_vad_segments_from_probs: Final speech segments after filtering: 0\n", nil
	}

	_, err := detectSpeechSegments(ctx, execute, filepath.Join(binDir, "ffmpeg"), "whisper-vad-speech-segments", "audio.wav", "vad.bin")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("detectSpeechSegments() error = %v, want context cancellation", err)
	}
}

func TestDetectSpeechSegmentsResetsVADStateAndRestoresAbsoluteTimeline(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(binDir, "ffmpeg"), `#!/bin/sh
set -eu
manifest=""
previous=""
for argument do
  if [ "$previous" = "-segment_list" ]; then manifest="$argument"; fi
  previous="$argument"
done
directory=${manifest%/*}
: > "$directory/vad_00000.wav"
: > "$directory/vad_00001.wav"
: > "$directory/vad_00002.wav"
printf '%s\n' \
  'vad_00000.wav,0.000000,15.000000' \
  'vad_00001.wav,15.000000,30.000000' \
  'vad_00002.wav,30.000000,45.000000' > "$manifest"
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	var mu sync.Mutex
	var audioPaths []string
	execute := func(_ context.Context, _ string, arguments []string) (string, error) {
		for index, argument := range arguments {
			if argument != "-f" || index+1 >= len(arguments) {
				continue
			}
			mu.Lock()
			audioPaths = append(audioPaths, arguments[index+1])
			mu.Unlock()
			break
		}
		return "whisper_vad_segments_from_probs: Final speech segments after filtering: 1\n" +
			"whisper_vad_segments_from_probs: VAD segment 0: start = 1.00, end = 2.00 (duration: 1.00)\n", nil
	}

	segments, err := detectSpeechSegments(context.Background(), execute, filepath.Join(binDir, "ffmpeg"), "whisper-vad-speech-segments", "audio.wav", "vad.bin")
	if err != nil {
		t.Fatalf("detectSpeechSegments() error = %v", err)
	}
	want := []domain.SpeechSegment{
		{Start: time.Second, End: 2 * time.Second},
		{Start: 16 * time.Second, End: 17 * time.Second},
		{Start: 31 * time.Second, End: 32 * time.Second},
	}
	if !reflect.DeepEqual(segments, want) {
		t.Errorf("segments = %#v, want %#v", segments, want)
	}
	if len(audioPaths) != 3 {
		t.Fatalf("VAD audio paths = %q, want 3 independently processed chunks", audioPaths)
	}
	for _, path := range audioPaths {
		if filepath.Base(path) == "audio.wav" {
			t.Errorf("VAD audio paths = %q, original long audio must not be processed as one stateful stream", audioPaths)
		}
	}
}

func TestDetectSpeechSegmentsUsesBoundedParallelVADWorkers(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(binDir, "ffmpeg"), `#!/bin/sh
set -eu
manifest=""
previous=""
for argument do
  if [ "$previous" = "-segment_list" ]; then manifest="$argument"; fi
  previous="$argument"
done
directory=${manifest%/*}
: > "$directory/vad_00000.wav"
: > "$directory/vad_00001.wav"
: > "$directory/vad_00002.wav"
: > "$directory/vad_00003.wav"
: > "$directory/vad_00004.wav"
printf '%s\n' \
  'vad_00000.wav,0.000000,15.000000' \
  'vad_00001.wav,15.000000,30.000000' \
  'vad_00002.wav,30.000000,45.000000' \
  'vad_00003.wav,45.000000,60.000000' \
  'vad_00004.wav,60.000000,75.000000' > "$manifest"
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	var active atomic.Int32
	var maximum atomic.Int32
	execute := func(_ context.Context, _ string, _ []string) (string, error) {
		current := active.Add(1)
		for current > maximum.Load() && !maximum.CompareAndSwap(maximum.Load(), current) {
		}
		time.Sleep(20 * time.Millisecond)
		active.Add(-1)
		return "whisper_vad_segments_from_probs: Final speech segments after filtering: 0\n", nil
	}

	if _, err := detectSpeechSegments(context.Background(), execute, filepath.Join(binDir, "ffmpeg"), "whisper-vad-speech-segments", "audio.wav", "vad.bin"); err != nil {
		t.Fatalf("detectSpeechSegments() error = %v", err)
	}
	if got := maximum.Load(); got != 4 {
		t.Errorf("maximum concurrent VAD calls = %d, want 4", got)
	}
}

func TestDetectSpeechSegmentsUsesDedicatedVADTool(t *testing.T) {
	t.Parallel()

	var gotExecutable string
	var gotArguments []string
	execute := func(_ context.Context, executable string, arguments []string) (string, error) {
		gotExecutable = executable
		gotArguments = append([]string(nil), arguments...)
		return "whisper_vad_segments_from_probs: Final speech segments after filtering: 1\n" +
			"whisper_vad_segments_from_probs: VAD segment 0: start = 1.00, end = 2.00 (duration: 1.00)\n", nil
	}

	segments, err := detectSpeechSegmentsInFile(context.Background(), execute, "/usr/local/bin/whisper-vad-speech-segments", "audio.wav", "vad.bin")
	if err != nil {
		t.Fatalf("detectSpeechSegmentsInFile() error = %v", err)
	}
	if gotExecutable != "/usr/local/bin/whisper-vad-speech-segments" {
		t.Errorf("executable = %q, want dedicated VAD tool", gotExecutable)
	}
	wantArguments := []string{
		"-t", "1",
		"-f", "audio.wav",
		"-vm", "vad.bin",
		"-vt", "0.35",
		"--vad-min-speech-duration-ms", "100",
		"--vad-speech-pad-ms", "250",
		"--vad-samples-overlap", "0.20",
		"--vad-max-speech-duration-s", "20",
	}
	if !reflect.DeepEqual(gotArguments, wantArguments) {
		t.Errorf("arguments = %#v, want %#v", gotArguments, wantArguments)
	}
	wantSegments := []domain.SpeechSegment{{Start: time.Second, End: 2 * time.Second}}
	if !reflect.DeepEqual(segments, wantSegments) {
		t.Errorf("segments = %#v, want %#v", segments, wantSegments)
	}
}

func TestLoadTranscriptionCuesPreservesChunkOrigin(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	firstPath := filepath.Join(directory, "first.wav")
	secondPath := filepath.Join(directory, "second.wav")
	payload := []byte(`{"transcription":[{"offsets":{"from":100,"to":900},"text":"발화","tokens":[{"text":" 발화","offsets":{"from":100,"to":900},"p":0.9}]}]}`)
	for _, path := range []string{firstPath, secondPath} {
		if err := os.WriteFile(path+".json", payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	chunkCues, err := loadTranscriptionCues([]portout.ChunkAudio{
		{Chunk: domain.Chunk{Start: 10 * time.Second, End: 12 * time.Second}, Path: firstPath},
		{Chunk: domain.Chunk{Start: 20 * time.Second, End: 23 * time.Second}, Path: secondPath, Index: 1},
	})
	if err != nil {
		t.Fatalf("loadTranscriptionCues() error = %v", err)
	}
	cues := append(chunkCues[0], chunkCues[1]...)
	if len(cues) != 2 {
		t.Fatalf("len(cues) = %d, want 2", len(cues))
	}
	if got := cues[1].Origin.Index; got != 1 {
		t.Errorf("second origin index = %d, want 1", got)
	}
	if got := cues[1].Origin.End; got != 23*time.Second {
		t.Errorf("second origin end = %s, want 23s", got)
	}
	if got := cues[1].Tokens[0].Origin; got != cues[1].Origin {
		t.Errorf("token origin = %#v, want %#v", got, cues[1].Origin)
	}
}

func TestTranscribeAudioChunksUsesSingleBeam(t *testing.T) {
	t.Parallel()

	var gotArguments []string
	runner := NewCommandRunner(1, func(_ context.Context, _ string, arguments []string) (string, error) {
		gotArguments = append([]string(nil), arguments...)
		return "", nil
	})

	err := transcribeAudioChunks(context.Background(), runner, "whisper-cli", "model.bin", "ja", []portout.ChunkAudio{{Path: "chunk.wav"}})
	if err != nil {
		t.Fatalf("transcribeAudioChunks() error = %v", err)
	}
	if !containsConsecutiveArguments(gotArguments, "-bs", "1") {
		t.Errorf("arguments = %q, want beam size 1", gotArguments)
	}
}

func TestTranscribeAudioChunksReportsCompletedOutputWhileBatchStillRuns(t *testing.T) {
	directory := t.TempDir()
	chunks := []portout.ChunkAudio{
		{Path: filepath.Join(directory, "first.wav")},
		{Path: filepath.Join(directory, "second.wav")},
	}
	firstOutputWritten := make(chan struct{})
	releaseSecondOutput := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseSecondOutput) }) }
	t.Cleanup(release)
	runner := NewCommandRunner(1, func(ctx context.Context, _ string, _ []string) (string, error) {
		if err := os.WriteFile(chunks[0].Path+".json", []byte(`{"transcription":[]}`), 0o644); err != nil {
			return "", err
		}
		close(firstOutputWritten)
		select {
		case <-releaseSecondOutput:
		case <-ctx.Done():
			return "", ctx.Err()
		}
		if err := os.WriteFile(chunks[1].Path+".json", []byte(`{"transcription":[]}`), 0o644); err != nil {
			return "", err
		}
		return "", nil
	})

	completed := make(chan int, len(chunks))
	result := make(chan error, 1)
	go func() {
		result <- transcribeAudioChunksWithProgressInterval(
			context.Background(),
			runner,
			"whisper-cli",
			"model.bin",
			"ko",
			chunks,
			time.Millisecond,
			func(count int) { completed <- count },
		)
	}()

	select {
	case <-firstOutputWritten:
	case <-time.After(time.Second):
		t.Fatal("first transcription output was not written")
	}
	select {
	case count := <-completed:
		if count != 1 {
			t.Fatalf("completed chunks = %d, want 1 before batch completion", count)
		}
	case <-time.After(time.Second):
		t.Fatal("first completed transcription was not reported")
	}

	release()
	select {
	case count := <-completed:
		if count != 2 {
			t.Errorf("completed chunks = %d, want 2 after batch completion", count)
		}
	case <-time.After(time.Second):
		t.Fatal("final completed transcription was not reported")
	}
	if err := <-result; err != nil {
		t.Fatalf("transcribeAudioChunksWithProgressInterval() error = %v", err)
	}
}
func TestParseSpeechSegmentsRejectsMissingSummary(t *testing.T) {
	t.Parallel()

	_, err := parseSpeechSegments("whisper changed its VAD log format")
	if err == nil || !strings.Contains(err.Error(), "VAD segment summary") {
		t.Fatalf("parseSpeechSegments() error = %v, want VAD summary error", err)
	}
}

func writeFakeFFmpeg(t *testing.T, binDir, content string) {
	t.Helper()
	const setupMarker = "set -eu\n"
	const segmentSupport = `manifest=""
previous=""
for argument do
  if [ "$previous" = "-segment_list" ]; then manifest="$argument"; fi
  previous="$argument"
done
if [ -n "$manifest" ]; then
  directory=${manifest%/*}
  : > "$directory/vad_00000.wav"
  printf '%s\n' 'vad_00000.wav,0.000000,100000.000000' > "$manifest"
  exit 0
fi
`
	if strings.Contains(content, setupMarker) {
		content = strings.Replace(content, setupMarker, setupMarker+segmentSupport, 1)
	}
	writeExecutable(t, filepath.Join(binDir, "ffmpeg"), content)
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func containsConsecutiveArguments(arguments []string, first, second string) bool {
	for index := 0; index+1 < len(arguments); index++ {
		if arguments[index] == first && arguments[index+1] == second {
			return true
		}
	}
	return false
}
