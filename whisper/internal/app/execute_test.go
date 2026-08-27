package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
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
	writeFakeFFmpeg(t, binDir, `#!/bin/sh
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

func TestExecuteReportsTranscriptionProgress(t *testing.T) {
	stdout, stderr, exitCode, _, _ := executeDirectory(t, []string{"meeting.mp3"}, "")

	if exitCode != 0 {
		t.Fatalf("Execute() exit code = %d, want 0; stderr = %q", exitCode, stderr)
	}
	for _, message := range []string{
		"[meeting.mp3] 오디오 추출 중...",
		"[meeting.mp3] 음성 구간 탐지 중...",
		"[meeting.mp3] 음성 조각 생성: 1개",
		"[meeting.mp3] 전사 중: 1개",
		"[meeting.mp3] 자막 정리 및 검증 중...",
		"[meeting.mp3] 결과 저장 중...",
	} {
		if !strings.Contains(stderr, message) {
			t.Errorf("stderr = %q, want progress message %q", stderr, message)
		}
	}
	if !strings.Contains(stdout, "완료:") {
		t.Errorf("stdout = %q, want completion output", stdout)
	}
	if strings.Contains(stdout, " 중...") {
		t.Errorf("stdout = %q, want progress only on stderr", stdout)
	}
}

func TestExecuteReportsLowConfidenceRetryProgress(t *testing.T) {
	t.Setenv("WHISPER_TEST_PROBABILITY", "0.31")
	_, stderr, exitCode, _, _ := executeDirectory(t, []string{"meeting.mp3"}, "")

	if exitCode != 0 {
		t.Fatalf("Execute() exit code = %d, want 0; stderr = %q", exitCode, stderr)
	}
	if !strings.Contains(stderr, "[meeting.mp3] 저신뢰 구간 재시도: 1개") {
		t.Errorf("stderr = %q, want low-confidence retry progress", stderr)
	}
}

func TestExecuteReportsWorkerCompletionBeforeStartingNextFile(t *testing.T) {
	nextInputMarker := filepath.Join(t.TempDir(), "c-started")
	t.Setenv("BLOCKED_INPUT", "a.mp3")
	t.Setenv("SIGNAL_INPUT", "c.mp3")
	t.Setenv("NEXT_INPUT_MARKER", nextInputMarker)
	_, stderr, exitCode, _, _ := executeDirectoryWithOptions(
		t,
		[]string{"a.mp3", "b.mp3", "c.mp3"},
		"",
		[]string{"-parallel", "2"},
		nil,
	)

	if exitCode != 0 {
		t.Fatalf("Execute() exit code = %d, want 0; stderr = %q", exitCode, stderr)
	}
	nextStart := strings.Index(stderr, "[c.mp3] 오디오 추출 중...")
	if nextStart < 0 {
		t.Fatalf("stderr = %q, want third-file start", stderr)
	}
	firstCompletion := strings.Index(stderr, "[a.mp3] 처리 완료")
	secondCompletion := strings.Index(stderr, "[b.mp3] 처리 완료")
	if (firstCompletion < 0 || firstCompletion > nextStart) && (secondCompletion < 0 || secondCompletion > nextStart) {
		t.Errorf("stderr = %q, want a worker completion before third-file start", stderr)
	}
}

func TestExecuteLimitsWhisperCommandsWhileKeepingThreeFileWorkers(t *testing.T) {
	concurrencyDir := t.TempDir()
	t.Setenv("WHISPER_CONCURRENCY_DIR", concurrencyDir)

	_, stderr, exitCode, _, _ := executeDirectoryWithOptions(
		t,
		[]string{"a.mp3", "b.mp3", "c.mp3"},
		"",
		[]string{"-parallel", "3"},
		nil,
	)

	if exitCode != 0 {
		t.Fatalf("Execute() exit code = %d, want 0; stderr = %q", exitCode, stderr)
	}
	content, err := os.ReadFile(filepath.Join(concurrencyDir, "max"))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(content)); got != "2" {
		t.Errorf("maximum concurrent whisper-cli commands = %s, want 2", got)
	}
}

func TestExecuteRecoversWhenWhisperCommandRunsOutOfMemoryOnce(t *testing.T) {
	oomMarker := filepath.Join(t.TempDir(), "oom-once")
	t.Setenv("WHISPER_OOM_ONCE_FILE", oomMarker)

	stdout, stderr, exitCode, _, _ := executeDirectory(t, []string{"meeting.mp3"}, "")

	if exitCode != 0 {
		t.Fatalf("Execute() exit code = %d, want 0; stderr = %q", exitCode, stderr)
	}
	if !strings.Contains(stdout, "요약: 성공 1개, 건너뜀 0개, 실패 0개") {
		t.Errorf("stdout = %q, want successful summary after OOM retry", stdout)
	}
	if _, err := os.Stat(oomMarker); err != nil {
		t.Fatalf("OOM marker was not created: %v", err)
	}
}

func TestExecuteReportsWorkerFailureBeforeStartingNextFile(t *testing.T) {
	nextInputMarker := filepath.Join(t.TempDir(), "c-started")
	t.Setenv("BLOCKED_INPUT", "a.mp3")
	t.Setenv("SIGNAL_INPUT", "c.mp3")
	t.Setenv("NEXT_INPUT_MARKER", nextInputMarker)
	stdout, stderr, exitCode, _, _ := executeDirectoryWithOptions(
		t,
		[]string{"a.mp3", "b.mp3", "c.mp3"},
		"b.mp3",
		[]string{"-parallel", "2"},
		nil,
	)

	if exitCode != 1 {
		t.Fatalf("Execute() exit code = %d, want 1; stderr = %q", exitCode, stderr)
	}
	failureMessage := "[b.mp3] 처리 실패: extract audio with ffmpeg: exit status 3: bad media"
	failure := strings.Index(stderr, failureMessage)
	if failure < 0 {
		t.Fatalf("stderr = %q, want detailed failed-worker state %q", stderr, failureMessage)
	}
	nextStart := strings.Index(stderr, "[c.mp3] 오디오 추출 중...")
	if nextStart < 0 {
		t.Fatalf("stderr = %q, want third-file start", stderr)
	}
	if failure > nextStart {
		t.Errorf("stderr = %q, want worker failure before third-file start", stderr)
	}
	if strings.Contains(stderr, "next input did not start") {
		t.Errorf("stderr = %q, want blocked worker released by third-file start", stderr)
	}
	if !strings.Contains(stdout, "요약: 성공 2개, 건너뜀 0개, 실패 1개") {
		t.Errorf("stdout = %q, want only configured input to fail", stdout)
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
	if !strings.Contains(stdout, "요약: 성공 1개, 건너뜀 0개, 실패 0개") {
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
	if !strings.Contains(stdout, "요약: 성공 2개, 건너뜀 0개, 실패 1개") {
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

func TestExecuteSkipsExistingDirectoryOutput(t *testing.T) {
	stdout, stderr, exitCode, inputDir, inputLog := executeDirectoryWithOptions(
		t,
		[]string{"a.mp3", "b.mp3"},
		"",
		[]string{"-format", "srt", "-parallel", "2"},
		[]string{"a.srt"},
	)

	if exitCode != 0 {
		t.Fatalf("Execute() exit code = %d, want 0; stderr = %q", exitCode, stderr)
	}
	if !strings.Contains(stdout, "건너뜀: "+filepath.Join(inputDir, "a.mp3")) {
		t.Errorf("stdout = %q, want skipped input", stdout)
	}
	if !strings.Contains(stdout, "완료: "+filepath.Join(inputDir, "b.srt")) {
		t.Errorf("stdout = %q, want completed output", stdout)
	}
	if !strings.Contains(stdout, "요약: 성공 1개, 건너뜀 1개, 실패 0개") {
		t.Errorf("stdout = %q, want skipped-file summary", stdout)
	}
	if strings.Contains(stderr, "[a.mp3] 처리 실패") {
		t.Errorf("stderr = %q, want existing output skipped without failure state", stderr)
	}
	loggedInputs, err := os.ReadFile(inputLog)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Fields(string(loggedInputs)), []string{filepath.Join(inputDir, "b.mp3")}; !slices.Equal(got, want) {
		t.Errorf("processed inputs = %q, want %q", got, want)
	}
	existingContent, err := os.ReadFile(filepath.Join(inputDir, "a.srt"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(existingContent), "existing transcript"; got != want {
		t.Errorf("existing transcript = %q, want %q", got, want)
	}
}

func TestExecuteSkipsExistingOutputBeforeLoadingDependencies(t *testing.T) {
	inputDir := t.TempDir()
	inputPath := filepath.Join(inputDir, "meeting.mp3")
	if err := os.WriteFile(inputPath, []byte("media"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inputDir, "meeting.srt"), []byte("existing transcript"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Execute(context.Background(), []string{
		"-format", "srt",
		"-model", filepath.Join(inputDir, "missing-model.bin"),
		inputDir,
	}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("Execute() exit code = %d, want 0; stderr = %q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "건너뜀: "+inputPath) {
		t.Errorf("stdout = %q, want skipped input", stdout.String())
	}
	if !strings.Contains(stdout.String(), "요약: 성공 0개, 건너뜀 1개, 실패 0개") {
		t.Errorf("stdout = %q, want skipped-file summary", stdout.String())
	}
}

func executeDirectory(t *testing.T, inputNames []string, failingInput string) (string, string, int, string, string) {
	t.Helper()

	return executeDirectoryWithOptions(t, inputNames, failingInput, nil, nil)
}

func executeDirectoryWithOptions(
	t *testing.T,
	inputNames []string,
	failingInput string,
	extraArgs []string,
	existingOutputs []string,
) (string, string, int, string, string) {
	t.Helper()

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeFFmpeg(t, binDir, `#!/bin/sh
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
    input_name=$(basename "$input")
    if [ "$input_name" = "$FAILING_INPUT" ]; then
      printf 'bad media' >&2
      exit 3
    fi
    if [ "$input_name" = "${SIGNAL_INPUT:-}" ]; then
      : > "${NEXT_INPUT_MARKER:?}"
    fi
    if [ "$input_name" = "${BLOCKED_INPUT:-}" ]; then
      attempts=0
      while [ ! -e "${NEXT_INPUT_MARKER:?}" ]; do
        attempts=$((attempts + 1))
        if [ "$attempts" -ge 500 ]; then
          printf 'next input did not start' >&2
          exit 4
        fi
        sleep 0.01
      done
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
	for _, name := range existingOutputs {
		if err := os.WriteFile(filepath.Join(inputDir, name), []byte("existing transcript"), 0o644); err != nil {
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
	args := make([]string, 0, 3+len(extraArgs))
	args = append(args, "-model", modelPath)
	args = append(args, extraArgs...)
	args = append(args, inputDir)
	exitCode := Execute(context.Background(), args, &stdout, &stderr)

	return stdout.String(), stderr.String(), exitCode, inputDir, inputLog
}
