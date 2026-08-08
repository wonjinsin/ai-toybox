package app

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestTranscribeCreatesTranscriptNextToInput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

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
format=txt
while [ "$#" -gt 0 ]; do
  case "$1" in
    -of) output_base="$2"; shift 2 ;;
    -otxt) format=txt; shift ;;
    -osrt) format=srt; shift ;;
    -ovtt) format=vtt; shift ;;
    *) shift ;;
  esac
done
printf '안녕하세요\n' > "${output_base}.${format}"
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	inputPath := filepath.Join(tempDir, "회의 영상.mp4")
	modelPath := filepath.Join(tempDir, "ggml-small.bin")
	if err := os.WriteFile(inputPath, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(modelPath, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}

	outputPath, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
	})
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}

	wantOutput := filepath.Join(tempDir, "회의 영상.txt")
	if outputPath != wantOutput {
		t.Fatalf("outputPath = %q, want %q", outputPath, wantOutput)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", outputPath, err)
	}
	if string(content) != "안녕하세요\n" {
		t.Errorf("transcript = %q, want %q", content, "안녕하세요\n")
	}
}

func TestResolveModelPathUsesEnvironmentVariable(t *testing.T) {
	modelPath := filepath.Join(t.TempDir(), "custom-model.bin")
	if err := os.WriteFile(modelPath, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WHISPER_MODEL", modelPath)

	resolved, err := resolveModelPath("")
	if err != nil {
		t.Fatalf("resolveModelPath() error = %v", err)
	}
	if resolved != modelPath {
		t.Errorf("resolveModelPath() = %q, want %q", resolved, modelPath)
	}
}

func TestResolveModelPathUsesWorkingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	modelDir := filepath.Join(tempDir, "models")
	if err := os.Mkdir(modelDir, 0o755); err != nil {
		t.Fatal(err)
	}
	modelPath := filepath.Join(modelDir, "ggml-large-v3.bin")
	if err := os.WriteFile(modelPath, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WHISPER_MODEL", "")
	t.Chdir(tempDir)

	resolved, err := resolveModelPath("")
	if err != nil {
		t.Fatalf("resolveModelPath() error = %v", err)
	}
	if resolved != modelPath {
		t.Errorf("resolveModelPath() = %q, want %q", resolved, modelPath)
	}
}

func TestRunCommandIncludesProcessError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executable uses POSIX shell")
	}

	executable := filepath.Join(t.TempDir(), "fails")
	writeExecutable(t, executable, "#!/bin/sh\nprintf 'decoder failed' >&2\nexit 7\n")

	err := runCommand(context.Background(), executable, nil)
	if err == nil {
		t.Fatal("runCommand() error = nil, want process error")
	}
	if !strings.Contains(err.Error(), "decoder failed") {
		t.Errorf("runCommand() error = %q, want process stderr", err)
	}
}

func TestTranscribeStopsWhenFFmpegFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(binDir, "ffmpeg"), "#!/bin/sh\nprintf 'bad media' >&2\nexit 3\n")
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\ntouch \"$WHISPER_SENTINEL\"\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	sentinelPath := filepath.Join(tempDir, "whisper-ran")
	t.Setenv("WHISPER_SENTINEL", sentinelPath)

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	_, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want ffmpeg error")
	}
	if !strings.Contains(err.Error(), "bad media") {
		t.Errorf("Transcribe() error = %q, want ffmpeg stderr", err)
	}
	if _, statErr := os.Stat(sentinelPath); !os.IsNotExist(statErr) {
		t.Errorf("whisper-cli ran after ffmpeg failure; stat error = %v", statErr)
	}
}

func TestTranscribeReportsMissingOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

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
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	_, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want missing output error")
	}
	if !strings.Contains(err.Error(), "did not create expected output") {
		t.Errorf("Transcribe() error = %q, want missing output context", err)
	}
}

func TestTranscribeRefusesToOverwriteTranscript(t *testing.T) {
	tempDir := t.TempDir()
	inputPath, modelPath := writeInputAndModel(t, tempDir)
	outputPath := filepath.Join(tempDir, "input.txt")
	if err := os.WriteFile(outputPath, []byte("keep me"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want overwrite protection error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Transcribe() error = %q, want existing output context", err)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "keep me" {
		t.Errorf("existing transcript = %q, want unchanged", content)
	}
}

func TestTranscribeDoesNotOverwriteOutputCreatedDuringProcessing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

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
printf 'other process\n' > "$WHISPER_FINAL_OUTPUT"
printf 'new transcript\n' > "${output_base}.txt"
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	outputPath := filepath.Join(tempDir, "input.txt")
	t.Setenv("WHISPER_FINAL_OUTPUT", outputPath)

	_, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want concurrent output error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Transcribe() error = %q, want existing output context", err)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "other process\n" {
		t.Errorf("concurrent output = %q, want unchanged", content)
	}
}

func TestTranscribeRemovesPartialOutputWhenWhisperFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

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
printf 'partial\n' > "${output_base}.txt"
printf 'inference failed\n' >&2
exit 4
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	_, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want whisper error")
	}
	if !strings.Contains(err.Error(), "inference failed") {
		t.Errorf("Transcribe() error = %q, want whisper stderr", err)
	}
	if _, statErr := os.Stat(filepath.Join(tempDir, "input.txt")); !os.IsNotExist(statErr) {
		t.Errorf("partial final output remains; stat error = %v", statErr)
	}
	temporaryOutputs, globErr := filepath.Glob(filepath.Join(tempDir, ".whisper-local-output-*"))
	if globErr != nil {
		t.Fatal(globErr)
	}
	if len(temporaryOutputs) != 0 {
		t.Errorf("temporary outputs remain: %v", temporaryOutputs)
	}
}

func TestTranscribeForceReplacesExistingTranscript(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

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
printf 'replacement\n' > "${output_base}.txt"
`)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	outputPath := filepath.Join(tempDir, "input.txt")
	if err := os.WriteFile(outputPath, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	resultPath, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
		Force:     true,
	})
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}
	if resultPath != outputPath {
		t.Errorf("Transcribe() = %q, want %q", resultPath, outputPath)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "replacement\n" {
		t.Errorf("transcript = %q, want replacement", content)
	}
}

func TestTranscribeRejectsInvalidOptionsAtExecutionBoundary(t *testing.T) {
	tempDir := t.TempDir()
	inputPath, modelPath := writeInputAndModel(t, tempDir)

	_, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "docx",
		ModelPath: modelPath,
	})
	if err == nil {
		t.Fatal("Transcribe() error = nil, want unsupported format error")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Transcribe() error = %q, want option validation context", err)
	}
}

func writeInputAndModel(t *testing.T, directory string) (string, string) {
	t.Helper()
	inputPath := filepath.Join(directory, "input.mp4")
	modelPath := filepath.Join(directory, "model.bin")
	if err := os.WriteFile(inputPath, []byte("media"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(modelPath, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	return inputPath, modelPath
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
