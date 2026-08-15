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
	for _, option := range []string{"-vad-model", "-corrections"} {
		if !strings.Contains(stdout.String(), option) {
			t.Errorf("stdout = %q, want %q option", stdout.String(), option)
		}
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
	writeFakeFFprobe(t, binDir, "60.0")
	writeSuccessfulWhisperPipeline(t, binDir, "transcript")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	inputPath := filepath.Join(tempDir, "input.mp4")
	modelPath := filepath.Join(tempDir, "model.bin")
	if err := os.WriteFile(inputPath, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(modelPath, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "ggml-silero-v6.2.0.bin"), []byte("vad"), 0o644); err != nil {
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

func TestExecuteProcessesEveryMediaFileInDirectory(t *testing.T) {
	stdout, stderr, exitCode, inputDir, inputLog := executeDirectory(t, []string{"a.mp3", "b.mp4"}, "")

	if exitCode != 0 {
		t.Fatalf("Execute() exit code = %d, want 0; stderr = %q", exitCode, stderr)
	}
	for _, name := range []string{"a.txt", "b.txt"} {
		if !strings.Contains(stdout, filepath.Join(inputDir, name)) {
			t.Errorf("stdout = %q, want output %q", stdout, name)
		}
	}
	loggedInputs, err := os.ReadFile(inputLog)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Fields(string(loggedInputs)); len(lines) != 2 {
		t.Errorf("processed inputs = %q, want 2 files", lines)
	}
}

func TestExecuteSummarizesDirectoryWithOneMediaFile(t *testing.T) {
	stdout, stderr, exitCode, _, _ := executeDirectory(t, []string{"only.mp3"}, "")

	if exitCode != 0 {
		t.Fatalf("Execute() exit code = %d, want 0; stderr = %q", exitCode, stderr)
	}
	if !strings.Contains(stdout, "요약: 성공 1개, 실패 0개") {
		t.Errorf("stdout = %q, want single-file directory summary", stdout)
	}
}

func TestExecuteContinuesDirectoryAfterOneFileFails(t *testing.T) {
	stdout, stderr, exitCode, inputDir, inputLog := executeDirectory(
		t,
		[]string{"a.mp3", "b.mp3", "c.mp3"},
		"b.mp3",
	)

	if exitCode != 1 {
		t.Errorf("Execute() exit code = %d, want 1", exitCode)
	}
	for _, name := range []string{"a.txt", "c.txt"} {
		if !strings.Contains(stdout, filepath.Join(inputDir, name)) {
			t.Errorf("stdout = %q, want successful output %q", stdout, name)
		}
	}
	if !strings.Contains(stderr, filepath.Join(inputDir, "b.mp3")) || !strings.Contains(stderr, "bad media") {
		t.Errorf("stderr = %q, want failed input and command error", stderr)
	}
	if !strings.Contains(stdout, "요약: 성공 2개, 실패 1개") {
		t.Errorf("stdout = %q, want partial failure summary", stdout)
	}
	loggedInputs, err := os.ReadFile(inputLog)
	if err != nil {
		t.Fatal(err)
	}
	if lines := strings.Fields(string(loggedInputs)); len(lines) != 3 {
		t.Errorf("processed inputs = %q, want all 3 files", lines)
	}
}

func executeDirectory(t *testing.T, inputNames []string, failingInput string) (string, string, int, string, string) {
	t.Helper()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(binDir, "ffmpeg"), `#!/bin/sh
set -eu
input=
output=
while [ "$#" -gt 0 ]; do
  case "$1" in
    -i) input="$2"; shift 2 ;;
    *) output="$1"; shift ;;
  esac
done
case "$input" in
  *.mp3|*.mp4)
    printf '%s\n' "$input" >> "$INPUT_LOG"
    if [ "$(basename "$input")" = "$FAILING_INPUT" ]; then
      printf 'bad media' >&2
      exit 3
    fi
    ;;
esac
: > "$output"
`)
	writeFakeFFprobe(t, binDir, "60.0")
	writeSuccessfulWhisperPipeline(t, binDir, "transcript")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	inputLog := filepath.Join(tempDir, "inputs.log")
	t.Setenv("INPUT_LOG", inputLog)
	t.Setenv("FAILING_INPUT", failingInput)

	inputDir := filepath.Join(tempDir, "recordings")
	if err := os.Mkdir(inputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range inputNames {
		if err := os.WriteFile(filepath.Join(inputDir, name), []byte("media"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	modelPath := filepath.Join(tempDir, "model.bin")
	if err := os.WriteFile(modelPath, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "ggml-silero-v6.2.0.bin"), []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Execute(context.Background(), []string{
		"-model", modelPath,
		inputDir,
	}, &stdout, &stderr)

	return stdout.String(), stderr.String(), exitCode, inputDir, inputLog
}
