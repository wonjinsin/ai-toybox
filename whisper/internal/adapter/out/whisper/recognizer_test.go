package whisper

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

func TestRecognizerPreservesEmptyChunkSlotsAndOriginIndices(t *testing.T) {
	directory := t.TempDir()
	chunks := []portout.ChunkAudio{
		{Chunk: domain.Chunk{Start: 10 * time.Second, End: 12 * time.Second}, Index: 9, Path: filepath.Join(directory, "first.wav")},
		{Chunk: domain.Chunk{Start: 20 * time.Second, End: 23 * time.Second}, Index: 14, Path: filepath.Join(directory, "second.wav")},
	}
	runner := NewCommandRunner(1, func(context.Context, string, []string) (string, error) {
		if err := os.WriteFile(chunks[0].Path+".json", []byte(`{"transcription":[]}`), 0o644); err != nil {
			return "", err
		}
		err := os.WriteFile(chunks[1].Path+".json", []byte(`{"transcription":[{"offsets":{"from":100,"to":900},"text":"hello","tokens":[{"text":" hello","offsets":{"from":100,"to":900},"p":0.9}]}]}`), 0o644)
		return "", err
	})
	recognizer := NewRecognizer("whisper-cli", "model.bin", "en", runner)
	var completed []int
	got, err := recognizer.Recognize(context.Background(), chunks, portout.RecognitionStandard, func(count int) { completed = append(completed, count) })
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || len(got[0]) != 0 || len(got[1]) != 1 {
		t.Fatalf("chunk slots = %#v", got)
	}
	cue := got[1][0]
	if cue.Origin.Index != 14 || cue.Origin.Start != chunks[1].Chunk.Start || cue.Start != 20100*time.Millisecond || cue.Tokens[0].Origin != cue.Origin {
		t.Fatalf("cue = %#v", cue)
	}
	if !reflect.DeepEqual(completed, []int{2}) {
		t.Fatalf("completed = %v", completed)
	}
}

func TestRecognizerRetryUsesDeterministicBeamAndPreservesFailure(t *testing.T) {
	wantError := errors.New("decoder failed")
	var arguments []string
	runner := NewCommandRunner(1, func(_ context.Context, _ string, args []string) (string, error) {
		arguments = append([]string(nil), args...)
		return "", wantError
	})
	recognizer := NewRecognizer("whisper-cli", "model.bin", "ja", runner)
	_, err := recognizer.Recognize(context.Background(), []portout.ChunkAudio{{Path: "retry.wav"}}, portout.RecognitionRetry, nil)
	if !errors.Is(err, wantError) || !strings.Contains(err.Error(), "retry low-confidence speech chunks") {
		t.Fatalf("error = %v", err)
	}
	if !containsConsecutiveArguments(arguments, "-tp", "0") || !containsConsecutiveArguments(arguments, "-bs", "8") {
		t.Fatalf("retry arguments = %q", arguments)
	}
}

func TestRecognizerReportsMissingOrMalformedOutputs(t *testing.T) {
	for _, test := range []struct{ name, payload, want string }{
		{"missing", "", "read whisper JSON"},
		{"malformed", "{", "parse whisper JSON"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "audio.wav")
			runner := NewCommandRunner(1, func(context.Context, string, []string) (string, error) {
				if test.payload != "" {
					return "", os.WriteFile(path+".json", []byte(test.payload), 0o644)
				}
				return "", nil
			})
			recognizer := NewRecognizer("whisper-cli", "model.bin", "ja", runner)
			_, err := recognizer.Recognize(context.Background(), []portout.ChunkAudio{{Path: path}}, portout.RecognitionStandard, nil)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestRecognizerEmptyBatchDoesNotExecute(t *testing.T) {
	runner := NewCommandRunner(1, func(context.Context, string, []string) (string, error) {
		t.Fatal("empty batch executed inference")
		return "", nil
	})
	recognizer := NewRecognizer("whisper-cli", "model.bin", "ja", runner)
	for _, mode := range []portout.RecognitionMode{portout.RecognitionStandard, portout.RecognitionRetry} {
		got, err := recognizer.Recognize(context.Background(), nil, mode, nil)
		if err != nil || len(got) != 0 {
			t.Fatalf("empty batch = %#v, %v", got, err)
		}
	}
}

func TestDetectorUsesConfiguredExecutableAndRestoresTimeline(t *testing.T) {
	directory := t.TempDir()
	ffmpeg := filepath.Join(directory, "ffmpeg")
	writeFakeFFmpeg(t, directory, "#!/bin/sh\nset -eu\nexit 0\n")
	vad := filepath.Join(directory, "vad")
	writeExecutable(t, vad, "#!/bin/sh\nprintf '%s\\n' 'Final speech segments after filtering: 1' 'VAD segment 0: start = 1.00, end = 2.00' >&2\n")
	runner := NewCommandRunner(0, nil)
	detector := NewDetector(ffmpeg, vad, "audio.wav", "vad.bin", runner.Execute)
	got, err := detector.Detect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []domain.SpeechSegment{{Start: time.Second, End: 2 * time.Second}}) {
		t.Fatalf("segments = %#v", got)
	}
}
