package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteShowsUsageWhenInputIsMissing(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Execute(context.Background(), nil, &stdout, &stderr)

	if exitCode != 2 {
		t.Errorf("Execute() exit code = %d, want 2", exitCode)
	}
	if !strings.Contains(stderr.String(), "사용법:") {
		t.Errorf("stderr = %q, want usage", stderr.String())
	}
}

func TestExecuteShowsHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Execute(context.Background(), []string{"-help"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("Execute() exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout.String(), "사용법:") {
		t.Errorf("stdout = %q, want usage", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestExecuteReportsTranscriptionFailure(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Execute(context.Background(), []string{"missing.mp4"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("Execute() exit code = %d, want 1", exitCode)
	}
	if !strings.Contains(stderr.String(), "전사 실패:") {
		t.Errorf("stderr = %q, want transcription failure", stderr.String())
	}
}

func TestExecuteReportsCreatedTranscript(t *testing.T) {
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(binDir, "ffmpeg"), `#!/bin/sh
set -eu
for argument do output="$argument"; done
: > "$output"
`)
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), `#!/bin/sh
set -eu
while [ "$#" -gt 0 ]; do
  case "$1" in
    -of) output_base="$2"; shift 2 ;;
    *) shift ;;
  esac
done
printf 'transcript\n' > "${output_base}.txt"
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	inputPath := filepath.Join(tempDir, "input.mp4")
	modelPath := filepath.Join(tempDir, "model.bin")
	if err := os.WriteFile(inputPath, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(modelPath, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Execute(context.Background(), []string{
		"-model", modelPath,
		inputPath,
	}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Execute() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), filepath.Join(tempDir, "input.txt")) {
		t.Errorf("stdout = %q, want transcript path", stdout.String())
	}
}
