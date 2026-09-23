package bootstrap

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/out/whisper"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

func TestFactoryOpensSessionsWithoutExecutingTools(t *testing.T) {
	request, directory := factoryRequest(t)
	writeTools(t, directory, []string{"ffmpeg", "ffprobe", "whisper-cli", "whisper-vad-speech-segments"})
	t.Setenv("PATH", directory)
	var calls int
	runner := whisper.NewCommandRunner(1, func(context.Context, string, []string) (string, error) {
		calls++
		return "", nil
	})
	factory := NewFactoryWithRunner(runner)
	for _, format := range []string{"srt", "txt", "vtt"} {
		t.Run(format, func(t *testing.T) {
			configured := request
			configured.Format = format
			session, err := factory.Open(context.Background(), configured)
			if err != nil {
				t.Fatal(err)
			}
			if session.Audio == nil || session.Detector == nil || session.Recognizer == nil || session.Transcripts == nil || session.Close == nil {
				t.Fatalf("incomplete session: %#v", session)
			}
			if (session.Checkpoints != nil) != (format == "srt") {
				t.Fatalf("checkpoints = %#v for format %q", session.Checkpoints, format)
			}
			if format == "srt" {
				if _, found, err := session.Checkpoints.Load(); err != nil || found {
					t.Fatalf("new session Load() = %t, %v", found, err)
				}
			}
			if err := session.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
	if calls != 0 {
		t.Fatalf("Open executed %d engine commands", calls)
	}
	reservations, err := filepath.Glob(filepath.Join(directory, ".whisper-local-output-*"))
	if err != nil || len(reservations) != 0 {
		t.Fatalf("closed session reservations = %v, %v", reservations, err)
	}
}

func TestFactoryReportsMissingToolsInPreflightOrder(t *testing.T) {
	request, directory := factoryRequest(t)
	t.Setenv("PATH", directory)
	factory := NewFactory()
	for _, tool := range []string{"ffmpeg", "ffprobe", "whisper-cli", "whisper-vad-speech-segments"} {
		if _, err := factory.Open(context.Background(), request); err == nil || !strings.Contains(err.Error(), tool+" not found:") {
			t.Fatalf("Open() error = %v, want missing %s", err, tool)
		}
		writeTools(t, directory, []string{tool})
	}
	missing := request
	missing.InputPath = filepath.Join(directory, "missing.wav")
	if _, err := factory.Open(context.Background(), missing); err == nil || !strings.Contains(err.Error(), "open input file") {
		t.Fatalf("Open() error = %v, want input validation", err)
	}
	blockedOutput := request
	blockedOutput.OutputDir = request.InputPath
	if _, err := factory.Open(context.Background(), blockedOutput); err == nil {
		t.Fatal("Open() accepted a regular file as output directory")
	}
}

func factoryRequest(t *testing.T) (portout.SessionRequest, string) {
	t.Helper()
	directory := t.TempDir()
	inputPath := filepath.Join(directory, "input.wav")
	modelPath := filepath.Join(directory, "model.bin")
	vadPath := filepath.Join(directory, "vad.bin")
	for _, path := range []string{inputPath, modelPath, vadPath} {
		if err := os.WriteFile(path, []byte("fixture"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return portout.SessionRequest{InputPath: inputPath, ModelPath: modelPath, VADModelPath: vadPath, Language: "ko", Format: "srt"}, directory
}

func writeTools(t *testing.T, directory string, names []string) {
	t.Helper()
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(directory, name), []byte("#!/bin/sh\nexit 99\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}
