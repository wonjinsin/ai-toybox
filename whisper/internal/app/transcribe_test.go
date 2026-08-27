package app

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestTranscribePersistsPartialAfterCompletedBatchWhenNextBatchFails(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeFFmpeg(t, binDir, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\n: > \"$output\"\n")
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\nexit 0\n")
	writeFakeFFprobe(t, binDir, "700.0")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	inputPath, modelPath := writeInputAndModel(t, tempDir)

	var mainCalls int
	runner := newWhisperCommandRunner(1, func(_ context.Context, executable string, args []string) (string, error) {
		if filepath.Base(executable) == "whisper-vad-speech-segments" {
			return vadLogWithSegments(129), nil
		}
		mainCalls++
		if mainCalls == 2 {
			return "", errors.New("second batch failed")
		}
		for _, argument := range args {
			if !strings.Contains(filepath.Base(argument), "chunk_") || filepath.Ext(argument) != ".wav" {
				continue
			}
			payload := []byte(`{"transcription":[{"offsets":{"from":100,"to":500},"text":"발화","tokens":[{"text":"발화","offsets":{"from":100,"to":500},"p":0.95}]}]}`)
			if err := os.WriteFile(argument+".json", payload, 0o644); err != nil {
				return "", err
			}
		}
		return "", nil
	})

	_, err := transcribeWithProgressUsingRunner(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "srt",
		ModelPath: modelPath,
	}, nil, runner)
	if err == nil || !strings.Contains(err.Error(), "second batch failed") {
		t.Fatalf("transcribeWithProgressUsingRunner() error = %v, want second batch failure", err)
	}
	if mainCalls != 2 {
		t.Fatalf("main transcription calls = %d, want 2", mainCalls)
	}
	partialPath := filepath.Join(tempDir, "input.partial.srt")
	partial, readErr := os.ReadFile(partialPath)
	if readErr != nil {
		t.Fatalf("read partial transcript: %v", readErr)
	}
	if got := strings.Count(string(partial), " --> "); got != 128 {
		t.Errorf("partial cues = %d, want 128", got)
	}
	checkpointPath := filepath.Join(tempDir, ".input.srt.whisper-local-checkpoint.json")
	if _, statErr := os.Stat(checkpointPath); statErr != nil {
		t.Errorf("checkpoint stat error = %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(tempDir, "input.srt")); !os.IsNotExist(statErr) {
		t.Errorf("final transcript exists after batch failure; stat error = %v", statErr)
	}
}

func TestTranscribeReconcilesBoundaryAcrossTranscriptionBatches(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeFFmpeg(t, binDir, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\n: > \"$output\"\n")
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\nexit 0\n")
	writeFakeFFprobe(t, binDir, "2600.0")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	inputPath, modelPath := writeInputAndModel(t, tempDir)

	var mainCalls int
	runner := newWhisperCommandRunner(1, func(_ context.Context, executable string, args []string) (string, error) {
		if filepath.Base(executable) == "whisper-vad-speech-segments" {
			return "whisper_vad_segments_from_probs: Final speech segments after filtering: 1\n" +
				"whisper_vad_segments_from_probs: VAD segment 0: start = 1.00, end = 2541.00 (duration: 2540.00)\n", nil
		}
		mainCalls++
		for _, argument := range args {
			baseName := filepath.Base(argument)
			if !strings.HasPrefix(baseName, "chunk_") || filepath.Ext(argument) != ".wav" {
				continue
			}
			var chunkIndex int
			if _, err := fmt.Sscanf(baseName, "chunk_%04d_", &chunkIndex); err != nil {
				return "", err
			}
			from, to, text := 100, 500, "일반"
			switch chunkIndex {
			case 127:
				from, to, text = 19750, 19850, "경계"
			case 128:
				from, to, text = 200, 300, "경계"
			}
			payload := []byte(fmt.Sprintf(`{"transcription":[{"offsets":{"from":%d,"to":%d},"text":%q,"tokens":[{"text":%q,"offsets":{"from":%d,"to":%d},"p":0.95}]}]}`, from, to, text, text, from, to))
			if err := os.WriteFile(argument+".json", payload, 0o644); err != nil {
				return "", err
			}
		}
		return "", nil
	})

	outputPath, err := transcribeWithProgressUsingRunner(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "srt",
		ModelPath: modelPath,
	}, nil, runner)
	if err != nil {
		t.Fatalf("transcribeWithProgressUsingRunner() error = %v", err)
	}
	if mainCalls != 2 {
		t.Fatalf("main transcription calls = %d, want 2", mainCalls)
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if got := strings.Count(string(content), " --> "); got != 128 {
		t.Errorf("final cues = %d, want 128 after boundary reconciliation", got)
	}
	partialPath, checkpointPath := incrementalArtifactPaths(strings.TrimSuffix(outputPath, ".srt"))
	if _, statErr := os.Stat(partialPath); !os.IsNotExist(statErr) {
		t.Errorf("partial transcript remains after final success; stat error = %v", statErr)
	}
	if _, statErr := os.Stat(checkpointPath); !os.IsNotExist(statErr) {
		t.Errorf("checkpoint remains after final success; stat error = %v", statErr)
	}
}

func TestTranscribeResumesWithoutRetranscribingCheckpointedChunks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeFFmpeg(t, binDir, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\n: > \"$output\"\n")
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\nexit 0\n")
	writeFakeFFprobe(t, binDir, "700.0")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	inputPath, modelPath := writeInputAndModel(t, tempDir)
	partialPath := filepath.Join(tempDir, "input.partial.srt")

	var vadCalls int
	var mainCalls int
	var partialRebuilt bool
	chunkAttempts := make(map[int]int)
	runner := newWhisperCommandRunner(1, func(_ context.Context, executable string, args []string) (string, error) {
		if filepath.Base(executable) == "whisper-vad-speech-segments" {
			vadCalls++
			return vadLogWithSegments(129), nil
		}
		mainCalls++
		if mainCalls == 3 {
			partial, err := os.ReadFile(partialPath)
			if err != nil {
				return "", fmt.Errorf("read rebuilt partial: %w", err)
			}
			partialRebuilt = strings.Count(string(partial), " --> ") == 128 && strings.Contains(string(partial), "수정")
		}
		for _, argument := range args {
			baseName := filepath.Base(argument)
			if !strings.HasPrefix(baseName, "chunk_") || filepath.Ext(argument) != ".wav" {
				continue
			}
			var chunkIndex int
			if _, err := fmt.Sscanf(baseName, "chunk_%04d_", &chunkIndex); err != nil {
				return "", err
			}
			chunkAttempts[chunkIndex]++
		}
		if mainCalls == 2 {
			return "", errors.New("second batch failed")
		}
		for _, argument := range args {
			if !strings.HasPrefix(filepath.Base(argument), "chunk_") || filepath.Ext(argument) != ".wav" {
				continue
			}
			payload := []byte(`{"transcription":[{"offsets":{"from":100,"to":500},"text":"발화","tokens":[{"text":"발화","offsets":{"from":100,"to":500},"p":0.95}]}]}`)
			if err := os.WriteFile(argument+".json", payload, 0o644); err != nil {
				return "", err
			}
		}
		return "", nil
	})
	options := Options{InputPath: inputPath, Language: "ko", Format: "srt", ModelPath: modelPath}

	if _, err := transcribeWithProgressUsingRunner(context.Background(), options, nil, runner); err == nil {
		t.Fatal("first transcription error = nil, want second batch failure")
	}
	correctionsPath := filepath.Join(tempDir, "corrections.json")
	if err := os.WriteFile(correctionsPath, []byte(`{"발화":"수정"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	options.CorrectionsPath = correctionsPath
	if err := os.Remove(partialPath); err != nil {
		t.Fatal(err)
	}
	var progressMessages []string
	outputPath, err := transcribeWithProgressUsingRunner(context.Background(), options, func(_ string, message string) {
		progressMessages = append(progressMessages, message)
	}, runner)
	if err != nil {
		t.Fatalf("resumed transcription error = %v", err)
	}
	if vadCalls != 1 {
		t.Errorf("VAD calls = %d, want 1", vadCalls)
	}
	if mainCalls != 3 {
		t.Errorf("main transcription calls = %d, want 3", mainCalls)
	}
	if !partialRebuilt {
		t.Error("partial transcript was not rebuilt from checkpoint before resumed batch")
	}
	if !containsArgument(progressMessages, "체크포인트 재개: 128/129개") {
		t.Errorf("progress messages = %q, want checkpoint resume count", progressMessages)
	}
	for index := range 128 {
		if chunkAttempts[index] != 1 {
			t.Errorf("chunk %d attempts = %d, want 1", index, chunkAttempts[index])
		}
	}
	if chunkAttempts[128] != 2 {
		t.Errorf("chunk 128 attempts = %d, want 2", chunkAttempts[128])
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if got := strings.Count(string(content), " --> "); got != 129 {
		t.Errorf("final cues = %d, want 129", got)
	}
	if !strings.Contains(string(content), "수정") {
		t.Errorf("final transcript = %q, want current corrections", content)
	}
}

func TestTranscribeResumesAfterCompletedLowConfidenceRetryBatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeFFmpeg(t, binDir, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\n: > \"$output\"\n")
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\nexit 0\n")
	writeFakeFFprobe(t, binDir, "700.0")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	inputPath, modelPath := writeInputAndModel(t, tempDir)
	partialPath := filepath.Join(tempDir, "input.partial.srt")

	var vadCalls int
	var mainCalls int
	var retryCalls int
	var improvedPartialRestored bool
	runner := newWhisperCommandRunner(1, func(_ context.Context, executable string, args []string) (string, error) {
		if filepath.Base(executable) == "whisper-vad-speech-segments" {
			vadCalls++
			return vadLogWithSegments(129), nil
		}
		isRetry := containsConsecutiveArguments(args, "-bs", "8")
		if !isRetry {
			mainCalls++
			for _, argument := range args {
				if !strings.HasPrefix(filepath.Base(argument), "chunk_") || filepath.Ext(argument) != ".wav" {
					continue
				}
				payload := []byte(`{"transcription":[{"offsets":{"from":100,"to":500},"text":"초기","tokens":[{"text":"초기","offsets":{"from":100,"to":500},"p":0.40}]}]}`)
				if err := os.WriteFile(argument+".json", payload, 0o644); err != nil {
					return "", err
				}
			}
			return "", nil
		}

		retryCalls++
		if retryCalls == 3 {
			partial, err := os.ReadFile(partialPath)
			if err != nil {
				return "", fmt.Errorf("read retry partial: %w", err)
			}
			improvedPartialRestored = strings.Contains(string(partial), "개선1")
		}
		if retryCalls == 2 {
			return "", errors.New("second retry batch failed")
		}
		for _, argument := range args {
			if !strings.HasPrefix(filepath.Base(argument), "chunk_") || filepath.Ext(argument) != ".wav" {
				continue
			}
			text := fmt.Sprintf("개선%d", retryCalls)
			payload := []byte(fmt.Sprintf(`{"transcription":[{"offsets":{"from":1000,"to":1200},"text":%q,"tokens":[{"text":%q,"offsets":{"from":1000,"to":1200},"p":0.95}]}]}`, text, text))
			if err := os.WriteFile(argument+".json", payload, 0o644); err != nil {
				return "", err
			}
		}
		return "", nil
	})
	options := Options{InputPath: inputPath, Language: "ko", Format: "srt", ModelPath: modelPath}

	if _, err := transcribeWithProgressUsingRunner(context.Background(), options, nil, runner); err == nil || !strings.Contains(err.Error(), "second retry batch failed") {
		t.Fatalf("first transcription error = %v, want second retry batch failure", err)
	}
	partial, readErr := os.ReadFile(partialPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !strings.Contains(string(partial), "개선1") {
		t.Errorf("partial transcript = %q, want completed retry result", partial)
	}

	outputPath, err := transcribeWithProgressUsingRunner(context.Background(), options, nil, runner)
	if err != nil {
		t.Fatalf("resumed transcription error = %v", err)
	}
	if vadCalls != 1 {
		t.Errorf("VAD calls = %d, want 1", vadCalls)
	}
	if mainCalls != 2 {
		t.Errorf("main transcription calls = %d, want 2", mainCalls)
	}
	if retryCalls != 3 {
		t.Errorf("retry calls = %d, want 3", retryCalls)
	}
	if !improvedPartialRestored {
		t.Error("completed retry result was not restored before remaining retry")
	}
	content, readErr := os.ReadFile(outputPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !strings.Contains(string(content), "개선1") || !strings.Contains(string(content), "개선3") {
		t.Errorf("final transcript = %q, want preserved and resumed retry results", content)
	}
}

func TestTranscribePreservesRetryBatchCursorWhenResumedRetryFailsAgain(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeFFmpeg(t, binDir, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\n: > \"$output\"\n")
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\nexit 0\n")
	writeFakeFFprobe(t, binDir, "700.0")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	inputPath, modelPath := writeInputAndModel(t, tempDir)

	var retryCalls int
	runner := newWhisperCommandRunner(1, func(_ context.Context, executable string, args []string) (string, error) {
		if filepath.Base(executable) == "whisper-vad-speech-segments" {
			return vadLogWithSegments(129), nil
		}
		if !containsConsecutiveArguments(args, "-bs", "8") {
			for _, argument := range args {
				if !strings.HasPrefix(filepath.Base(argument), "chunk_") || filepath.Ext(argument) != ".wav" {
					continue
				}
				payload := []byte(`{"transcription":[{"offsets":{"from":100,"to":500},"text":"초기","tokens":[{"text":"초기","offsets":{"from":100,"to":500},"p":0.40}]}]}`)
				if err := os.WriteFile(argument+".json", payload, 0o644); err != nil {
					return "", err
				}
			}
			return "", nil
		}

		retryCalls++
		if retryCalls == 2 {
			return "", errors.New("first resumed cue failed")
		}
		if retryCalls == 3 {
			return "", errors.New("second resumed cue failed")
		}
		probability := 0.40
		text := "변화없음"
		if retryCalls >= 4 {
			probability = 0.95
			text = "개선"
		}
		for _, argument := range args {
			if !strings.HasPrefix(filepath.Base(argument), "chunk_") || filepath.Ext(argument) != ".wav" {
				continue
			}
			payload := []byte(fmt.Sprintf(`{"transcription":[{"offsets":{"from":1000,"to":1200},"text":%q,"tokens":[{"text":%q,"offsets":{"from":1000,"to":1200},"p":%.2f}]}]}`, text, text, probability))
			if err := os.WriteFile(argument+".json", payload, 0o644); err != nil {
				return "", err
			}
		}
		return "", nil
	})
	options := Options{InputPath: inputPath, Language: "ko", Format: "srt", ModelPath: modelPath}

	if _, err := transcribeWithProgressUsingRunner(context.Background(), options, nil, runner); err == nil || !strings.Contains(err.Error(), "first resumed cue failed") {
		t.Fatalf("first transcription error = %v, want first retry failure", err)
	}
	if _, err := transcribeWithProgressUsingRunner(context.Background(), options, nil, runner); err == nil || !strings.Contains(err.Error(), "second resumed cue failed") {
		t.Fatalf("second transcription error = %v, want repeated retry failure", err)
	}

	checkpointPath := filepath.Join(tempDir, ".input.srt.whisper-local-checkpoint.json")
	checkpointContent, err := os.ReadFile(checkpointPath)
	if err != nil {
		t.Fatal(err)
	}
	var checkpoint incrementalCheckpoint
	if err := json.Unmarshal(checkpointContent, &checkpoint); err != nil {
		t.Fatal(err)
	}
	if checkpoint.Stage != checkpointStageRetrying || checkpoint.RetryCursor != 128 {
		t.Errorf("checkpoint state = %s/%d, want retrying/128", checkpoint.Stage, checkpoint.RetryCursor)
	}

	if _, err := transcribeWithProgressUsingRunner(context.Background(), options, nil, runner); err != nil {
		t.Fatalf("third transcription error = %v", err)
	}
	if retryCalls != 4 {
		t.Errorf("retry calls = %d, want 4 without repeating completed retry", retryCalls)
	}
}

func TestTranscribeKeepsIncrementalArtifactsWhenFinalOutputAppearsDuringRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeFFmpeg(t, binDir, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\n: > \"$output\"\n")
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\nexit 0\n")
	writeFakeFFprobe(t, binDir, "10.0")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	inputPath, modelPath := writeInputAndModel(t, tempDir)
	finalPath := filepath.Join(tempDir, "input.srt")

	runner := newWhisperCommandRunner(1, func(_ context.Context, executable string, args []string) (string, error) {
		if filepath.Base(executable) == "whisper-vad-speech-segments" {
			return vadLogWithSegments(1), nil
		}
		for _, argument := range args {
			if !strings.HasPrefix(filepath.Base(argument), "chunk_") || filepath.Ext(argument) != ".wav" {
				continue
			}
			payload := []byte(`{"transcription":[{"offsets":{"from":100,"to":500},"text":"발화","tokens":[{"text":"발화","offsets":{"from":100,"to":500},"p":0.95}]}]}`)
			if err := os.WriteFile(argument+".json", payload, 0o644); err != nil {
				return "", err
			}
		}
		if err := os.WriteFile(finalPath, []byte("external\n"), 0o644); err != nil {
			return "", err
		}
		return "", nil
	})

	_, err := transcribeWithProgressUsingRunner(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "srt",
		ModelPath: modelPath,
	}, nil, runner)
	if !isOutputExistsError(err) {
		t.Fatalf("transcription error = %v, want output-exists error", err)
	}
	content, readErr := os.ReadFile(finalPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != "external\n" {
		t.Errorf("final transcript = %q, want external content preserved", content)
	}
	partialPath, checkpointPath := incrementalArtifactPaths(strings.TrimSuffix(finalPath, ".srt"))
	if _, statErr := os.Stat(partialPath); statErr != nil {
		t.Errorf("partial transcript stat error = %v", statErr)
	}
	if _, statErr := os.Stat(checkpointPath); statErr != nil {
		t.Errorf("checkpoint stat error = %v", statErr)
	}
}

func vadLogWithSegments(count int) string {
	var log strings.Builder
	fmt.Fprintf(&log, "whisper_vad_segments_from_probs: Final speech segments after filtering: %d\n", count)
	for index := range count {
		start := float64(index*5 + 1)
		end := start + 1
		fmt.Fprintf(&log, "whisper_vad_segments_from_probs: VAD segment %d: start = %.2f, end = %.2f (duration: 1.00)\n", index, start, end)
	}
	return log.String()
}

func TestTranscribeCreatesTranscriptNextToInput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

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
	writeSuccessfulWhisperPipeline(t, binDir, "안녕하세요")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	inputPath := filepath.Join(tempDir, "회의 영상.mp4")
	modelPath := filepath.Join(tempDir, "ggml-small.bin")
	if err := os.WriteFile(inputPath, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(modelPath, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "ggml-silero-v6.2.0.bin"), []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeFakeFFprobe(t, binDir, "60.0")

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

func TestTranscribeNormalizesAudioForQuietSpeech(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeFakeFFmpeg(t, binDir, `#!/bin/sh
set -eu
printf '%s\n' "$@" > "$FFMPEG_ARGS"
for argument do output="$argument"; done
: > "$output"
`)
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\nprintf '%s\\n' 'whisper_vad_segments_from_probs: Final speech segments after filtering: 0' >&2\n")
	writeFakeVADTool(t, binDir, vadLogWithSegments(0))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	argsPath := filepath.Join(tempDir, "ffmpeg-args")
	t.Setenv("FFMPEG_ARGS", argsPath)

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	if _, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
	}); err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}

	captured, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	arguments := strings.Fields(string(captured))
	if !containsConsecutiveArguments(arguments, "-af", "dynaudnorm=f=500:g=31:p=0.95:m=100:b=1") {
		t.Errorf("ffmpeg arguments = %q, want quiet-speech normalization filter", arguments)
	}
}

func TestTranscribeAmplifiesQuietAudioAfterSilence(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executable uses POSIX shell")
	}
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg is not installed")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(ffmpegPath, filepath.Join(binDir, "ffmpeg")); err != nil {
		t.Fatal(err)
	}
	writeFakeFFprobe(t, binDir, "18.0")
	writeExecutable(t, filepath.Join(binDir, "whisper-vad-speech-segments"), `#!/bin/sh
set -eu
while [ "$#" -gt 0 ]; do
  case "$1" in
    -f) input_audio="$2"; shift 2 ;;
    *) shift ;;
  esac
done
cp "$input_audio" "$NORMALIZED_AUDIO_DIR/${input_audio##*/}"
printf '%s\n' 'whisper_vad_segments_from_probs: Final speech segments after filtering: 0' >&2
`)
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	inputPath := filepath.Join(tempDir, "quiet.wav")
	inputSamples := quietToneAfterSilence(16000)
	writePCM16WAV(t, inputPath, 16000, inputSamples)
	modelPath := filepath.Join(tempDir, "model.bin")
	if err := os.WriteFile(modelPath, []byte("model"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "ggml-silero-v6.2.0.bin"), []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}
	normalizedDirectory := filepath.Join(tempDir, "normalized")
	if err := os.Mkdir(normalizedDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NORMALIZED_AUDIO_DIR", normalizedDirectory)

	if _, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
	}); err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}

	normalizedPaths, err := filepath.Glob(filepath.Join(normalizedDirectory, "vad_*.wav"))
	if err != nil {
		t.Fatal(err)
	}
	var normalizedSamples []int16
	for _, path := range normalizedPaths {
		normalizedSamples = append(normalizedSamples, readPCM16WAV(t, path)...)
	}
	if len(normalizedSamples) != len(inputSamples) {
		t.Fatalf("normalized samples = %d, want %d", len(normalizedSamples), len(inputSamples))
	}
	inputPeak := pcmPeak(inputSamples)
	normalizedPeak := pcmPeak(normalizedSamples)
	if normalizedPeak < inputPeak*50 || normalizedPeak > inputPeak*110 {
		t.Errorf("normalized peak = %d, want 50x to 110x input peak %d", normalizedPeak, inputPeak)
	}
	leadingSilenceSamples := 8 * 16000
	if peak := pcmPeak(normalizedSamples[:leadingSilenceSamples]); peak != 0 {
		t.Errorf("normalized leading silence peak = %d, want 0", peak)
	}
	inputOnset := firstSignalSample(inputSamples, 10)
	normalizedOnset := firstSignalSample(normalizedSamples, 10)
	if normalizedOnset < inputOnset-1 || normalizedOnset > inputOnset+1 {
		t.Errorf("normalized onset = %d, want within one sample of input onset %d", normalizedOnset, inputOnset)
	}
}

func TestTranscribePreservesSilenceGapFromVADLog(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

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
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), `#!/bin/sh
set -eu
index=0
for argument do
  case "$argument" in
    *.wav)
      index=$((index + 1))
      printf '{"transcription":[{"offsets":{"from":150,"to":1150},"text":"발화%d","tokens":[{"text":"발화%d","offsets":{"from":150,"to":1150},"p":0.95}]}]}\n' "$index" "$index" > "$argument.json"
      ;;
  esac
done
`)
	writeFakeVADTool(t, binDir, "whisper_vad_segments_from_probs: Final speech segments after filtering: 2\n"+
		"whisper_vad_segments_from_probs: VAD segment 0: start = 9.89, end = 12.67 (duration: 2.78)\n"+
		"whisper_vad_segments_from_probs: VAD segment 1: start = 32.65, end = 35.16 (duration: 2.51)\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	if err := os.WriteFile(filepath.Join(tempDir, "ggml-silero-v6.2.0.bin"), []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}
	outputPath, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "srt",
		ModelPath: modelPath,
	})
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "00:00:32,650 --> 00:00:33,650") {
		t.Errorf("transcript = %q, want second cue aligned to VAD segment", content)
	}
}

func TestTranscribeCreatesEmptySubtitleWhenNoSpeechIsDetected(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

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
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), `#!/bin/sh
printf '%s\n' 'whisper_vad_segments_from_probs: Final speech segments after filtering: 0' >&2
`)
	writeFakeVADTool(t, binDir, vadLogWithSegments(0))
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	if err := os.WriteFile(filepath.Join(tempDir, "ggml-silero-v6.2.0.bin"), []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}
	outputPath, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "srt",
		ModelPath: modelPath,
	})
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "" {
		t.Errorf("empty transcript = %q, want empty", content)
	}
}

func TestParseSpeechSegmentsRejectsMissingSummary(t *testing.T) {
	t.Parallel()

	_, err := parseSpeechSegments("whisper changed its VAD log format")
	if err == nil || !strings.Contains(err.Error(), "VAD segment summary") {
		t.Fatalf("parseSpeechSegments() error = %v, want VAD summary error", err)
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

func TestResolveVADModelPathDoesNotUseLegacyModel(t *testing.T) {
	modelDir := t.TempDir()
	vadPath := filepath.Join(modelDir, "ggml-silero-v5.1.2.bin")
	if err := os.WriteFile(vadPath, []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}

	if resolved := resolveVADModelPath(filepath.Join(modelDir, "ggml-large-v3.bin")); resolved != "" {
		t.Errorf("resolveVADModelPath() = %q, want empty", resolved)
	}
}

func TestResolveVADModelPathPrefersSileroV62(t *testing.T) {
	modelDir := t.TempDir()
	legacyPath := filepath.Join(modelDir, "ggml-silero-v5.1.2.bin")
	if err := os.WriteFile(legacyPath, []byte("legacy vad"), 0o644); err != nil {
		t.Fatal(err)
	}
	v62Path := filepath.Join(modelDir, "ggml-silero-v6.2.0.bin")
	if err := os.WriteFile(v62Path, []byte("v6.2 vad"), 0o644); err != nil {
		t.Fatal(err)
	}

	if resolved := resolveVADModelPath(filepath.Join(modelDir, "ggml-large-v3.bin")); resolved != v62Path {
		t.Errorf("resolveVADModelPath() = %q, want %q", resolved, v62Path)
	}
}

func TestResolveVADModelPathMissingReturnsEmpty(t *testing.T) {
	if resolved := resolveVADModelPath(filepath.Join(t.TempDir(), "ggml-large-v3.bin")); resolved != "" {
		t.Errorf("resolveVADModelPath() = %q, want empty", resolved)
	}
}

func TestResolveRequiredVADModelPathUsesExplicitPath(t *testing.T) {
	t.Parallel()

	modelDir := t.TempDir()
	modelPath := filepath.Join(modelDir, "model.bin")
	explicitPath := filepath.Join(modelDir, "custom-vad.bin")
	if err := os.WriteFile(explicitPath, []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := resolveRequiredVADModelPath(explicitPath, modelPath)
	if err != nil {
		t.Fatalf("resolveRequiredVADModelPath() error = %v", err)
	}
	if got != explicitPath {
		t.Errorf("resolveRequiredVADModelPath() = %q, want %q", got, explicitPath)
	}
}

func TestLoadCorrectionsReadsJSONMap(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "corrections.json")
	if err := os.WriteFile(path, []byte(`{"レッサン":"レッスン"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := loadCorrections(path)
	if err != nil {
		t.Fatalf("loadCorrections() error = %v", err)
	}
	if got["レッサン"] != "レッスン" {
		t.Errorf("loadCorrections() = %#v, want correction map", got)
	}
}

func TestLoadCorrectionsRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "corrections.json")
	if err := os.WriteFile(path, []byte(`{"broken":`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := loadCorrections(path)
	if err == nil || !strings.Contains(err.Error(), "decode corrections file") {
		t.Fatalf("loadCorrections() error = %v, want JSON decode context", err)
	}
}

func TestValidateSubtitleCuesRejectsMediaOverflow(t *testing.T) {
	t.Parallel()

	err := validateSubtitleCues([]subtitleCue{{
		Start: 9 * time.Second,
		End:   11 * time.Second,
		Text:  "overflow",
	}}, 10*time.Second)
	if err == nil || !strings.Contains(err.Error(), "exceeds media duration") {
		t.Fatalf("validateSubtitleCues() error = %v, want media duration error", err)
	}
}

func TestTranscribeUsesOriginalTimelineChunksAndSingleMultiFileInference(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeExecutable(t, filepath.Join(binDir, "ffprobe"), "#!/bin/sh\nprintf '60.0\\n'\n")
	writeFakeFFmpeg(t, binDir, `#!/bin/sh
set -eu
for argument do output="$argument"; done
: > "$output"
`)
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), `#!/bin/sh
set -eu
printf 'call\n' >> "$WHISPER_CALLS"
for argument do
  case "$argument" in
    *.wav)
      printf '%s\n' '{"transcription":[{"offsets":{"from":150,"to":1150},"text":"こんにちは","tokens":[{"text":"こんにちは","offsets":{"from":150,"to":1150},"p":0.95}]}]}' > "$argument.json"
      ;;
  esac
done
`)
	writeFakeVADTool(t, binDir, "whisper_vad_segments_from_probs: Final speech segments after filtering: 3\n"+
		"whisper_vad_segments_from_probs: VAD segment 0: start = 10.00, end = 12.00 (duration: 2.00)\n"+
		"whisper_vad_segments_from_probs: VAD segment 1: start = 12.50, end = 15.00 (duration: 2.50)\n"+
		"whisper_vad_segments_from_probs: VAD segment 2: start = 30.00, end = 55.00 (duration: 25.00)\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	callLog := filepath.Join(tempDir, "whisper-calls")
	t.Setenv("WHISPER_CALLS", callLog)

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	if err := os.WriteFile(filepath.Join(tempDir, "ggml-silero-v6.2.0.bin"), []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}

	outputPath, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ja",
		Format:    "srt",
		ModelPath: modelPath,
	})
	if err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, timing := range []string{
		"00:00:10,000 --> 00:00:11,000",
		"00:00:30,000 --> 00:00:31,000",
		"00:00:49,700 --> 00:00:50,700",
	} {
		if !strings.Contains(string(content), timing) {
			t.Errorf("transcript = %q, want original timing %q", content, timing)
		}
	}
	calls, err := os.ReadFile(callLog)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(strings.Fields(string(calls))); got != 1 {
		t.Errorf("whisper-cli calls = %d, want one multi-file inference", got)
	}
}

func TestTranscribeRetriesLowConfidenceCueWithDeterministicDecoding(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeFFprobe(t, binDir, "10.0")
	writeFakeFFmpeg(t, binDir, `#!/bin/sh
set -eu
for argument do output="$argument"; done
: > "$output"
`)
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), `#!/bin/sh
set -eu
case " $* " in
  *" -bs 8 "*)
    printf '%s\n' "$@" > "$RETRY_ARGS"
    payload='{"transcription":[{"offsets":{"from":1000,"to":2000},"text":"재시도","tokens":[{"text":" 재시도","offsets":{"from":1000,"to":2000},"p":0.9}]}]}'
    ;;
  *)
    payload='{"transcription":[{"offsets":{"from":150,"to":1150},"text":"초기","tokens":[{"text":" 초기","offsets":{"from":150,"to":1150},"p":0.31}]}]}'
    ;;
esac
for argument do
  case "$argument" in
    *.wav) printf '%s\n' "$payload" > "$argument.json" ;;
  esac
done
`)
	writeFakeVADTool(t, binDir, "whisper_vad_segments_from_probs: Final speech segments after filtering: 1\n"+
		"whisper_vad_segments_from_probs: VAD segment 0: start = 1.00, end = 2.00 (duration: 1.00)\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	retryArgsPath := filepath.Join(tempDir, "retry-args")
	t.Setenv("RETRY_ARGS", retryArgsPath)

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	if err := os.WriteFile(filepath.Join(tempDir, "ggml-silero-v6.2.0.bin"), []byte("vad"), 0o644); err != nil {
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
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "재시도\n" {
		t.Errorf("transcript = %q, want retry result", content)
	}
	arguments, err := os.ReadFile(retryArgsPath)
	if err != nil {
		t.Fatalf("read retry arguments: %v", err)
	}
	if !containsConsecutiveArguments(strings.Fields(string(arguments)), "-tp", "0") || !containsConsecutiveArguments(strings.Fields(string(arguments)), "-bs", "8") {
		t.Errorf("retry arguments = %q, want temperature 0 and beam size 8", arguments)
	}
}

func TestTranscribeUsesSpeechPreservingVADSettings(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

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
	writeExecutable(t, filepath.Join(binDir, "whisper-vad-speech-segments"), `#!/bin/sh
set -eu
printf '%s\n' "$@" > "$VAD_ARGS"
printf '%s\n' 'whisper_vad_segments_from_probs: Final speech segments after filtering: 0' >&2
`)
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	argsPath := filepath.Join(tempDir, "vad-args")
	t.Setenv("VAD_ARGS", argsPath)

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	vadPath := filepath.Join(tempDir, "ggml-silero-v6.2.0.bin")
	if err := os.WriteFile(vadPath, []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
	}); err != nil {
		t.Fatalf("Transcribe() error = %v", err)
	}

	captured, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatal(err)
	}
	arguments := strings.Fields(string(captured))
	if containsArgument(arguments, "-d") {
		t.Errorf("VAD arguments = %q, duration limit must not be set during VAD scan", arguments)
	}
	for _, expected := range [][2]string{
		{"-vm", vadPath},
		{"-vt", "0.35"},
		{"--vad-min-speech-duration-ms", "100"},
		{"--vad-speech-pad-ms", "250"},
		{"--vad-samples-overlap", "0.20"},
		{"--vad-max-speech-duration-s", "20"},
	} {
		if !containsConsecutiveArguments(arguments, expected[0], expected[1]) {
			t.Errorf("VAD arguments = %q, want %q followed by %q", arguments, expected[0], expected[1])
		}
	}
	for _, expected := range []string{"-m", "-l", "-mc", "-sns", "--vad", "--vad-min-silence-duration-ms"} {
		if !containsArgument(arguments, expected) {
			continue
		}
		t.Errorf("VAD arguments = %q, must not include transcription option %q", arguments, expected)
	}
}

func TestTranscribeRequiresVADModel(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeFFmpeg(t, binDir, "#!/bin/sh\ntouch \"$COMMAND_SENTINEL\"\n")
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\ntouch \"$COMMAND_SENTINEL\"\n")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	sentinelPath := filepath.Join(tempDir, "command-ran")
	t.Setenv("COMMAND_SENTINEL", sentinelPath)

	inputPath, modelPath := writeInputAndModel(t, tempDir)
	if err := os.Remove(filepath.Join(tempDir, "ggml-silero-v6.2.0.bin")); err != nil {
		t.Fatal(err)
	}
	_, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
	})
	if err == nil || !strings.Contains(err.Error(), "VAD model not found") {
		t.Fatalf("Transcribe() error = %v, want missing VAD model error", err)
	}
	if _, statErr := os.Stat(sentinelPath); !os.IsNotExist(statErr) {
		t.Errorf("external command ran without VAD model; stat error = %v", statErr)
	}
}

func TestTranscribeRequiresDedicatedVADTool(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executables use POSIX shell")
	}

	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, "bin")
	if err := os.Mkdir(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFakeFFmpeg(t, binDir, "#!/bin/sh\nexit 0\n")
	writeExecutable(t, filepath.Join(binDir, "ffprobe"), "#!/bin/sh\nexit 0\n")
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", binDir)

	inputPath := filepath.Join(tempDir, "input.mp4")
	modelPath := filepath.Join(tempDir, "model.bin")
	for path, content := range map[string]string{
		inputPath: "media",
		modelPath: "model",
		filepath.Join(tempDir, "ggml-silero-v6.2.0.bin"): "vad",
	} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	_, err := Transcribe(context.Background(), Options{
		InputPath: inputPath,
		Language:  "ko",
		Format:    "txt",
		ModelPath: modelPath,
	})
	if err == nil || !strings.Contains(err.Error(), "whisper-vad-speech-segments not found") {
		t.Fatalf("Transcribe() error = %v, want missing dedicated VAD tool error", err)
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

func containsArgument(arguments []string, expected string) bool {
	for _, argument := range arguments {
		if argument == expected {
			return true
		}
	}
	return false
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
	writeFakeFFmpeg(t, binDir, "#!/bin/sh\nprintf 'bad media' >&2\nexit 3\n")
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
	writeFakeFFmpeg(t, binDir, `#!/bin/sh
set -eu
for argument do output="$argument"; done
: > "$output"
`)
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), `#!/bin/sh
set -eu
case " $* " in
  *" --vad "*)
    printf '%s\n' \
      'whisper_vad_segments_from_probs: Final speech segments after filtering: 1' \
      'whisper_vad_segments_from_probs: VAD segment 0: start = 1.00, end = 2.00 (duration: 1.00)' >&2
    ;;
esac
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
		t.Fatal("Transcribe() error = nil, want missing output error")
	}
	if !strings.Contains(err.Error(), "read whisper JSON") {
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
	if !isOutputExistsError(err) {
		t.Errorf("Transcribe() error = %v, want output-exists classification", err)
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
	writeFakeFFmpeg(t, binDir, `#!/bin/sh
set -eu
for argument do output="$argument"; done
: > "$output"
`)
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), `#!/bin/sh
set -eu
case " $* " in
  *" --vad "*)
    printf '%s\n' \
      'whisper_vad_segments_from_probs: Final speech segments after filtering: 1' \
      'whisper_vad_segments_from_probs: VAD segment 0: start = 1.00, end = 2.00 (duration: 1.00)' >&2
    exit 0
    ;;
esac
printf 'other process\n' > "$WHISPER_FINAL_OUTPUT"
for argument do
  case "$argument" in
    *.wav)
      printf '%s\n' '{"transcription":[{"offsets":{"from":150,"to":1150},"text":"new transcript","tokens":[{"text":"new transcript","offsets":{"from":150,"to":1150},"p":0.95}]}]}' > "$argument.json"
      ;;
  esac
done
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
	if !isOutputExistsError(err) {
		t.Errorf("Transcribe() error = %v, want output-exists classification", err)
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
	writeFakeFFmpeg(t, binDir, `#!/bin/sh
set -eu
for argument do output="$argument"; done
: > "$output"
`)
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), `#!/bin/sh
set -eu
case " $* " in
  *" --vad "*)
    printf '%s\n' \
      'whisper_vad_segments_from_probs: Final speech segments after filtering: 1' \
      'whisper_vad_segments_from_probs: VAD segment 0: start = 1.00, end = 2.00 (duration: 1.00)' >&2
    exit 0
    ;;
esac
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
	writeFakeFFmpeg(t, binDir, `#!/bin/sh
set -eu
for argument do output="$argument"; done
: > "$output"
`)
	writeSuccessfulWhisperPipeline(t, binDir, "replacement")
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
	if err := os.WriteFile(filepath.Join(directory, "ggml-silero-v6.2.0.bin"), []byte("vad"), 0o644); err != nil {
		t.Fatal(err)
	}
	binDir := filepath.Join(directory, "bin")
	if info, err := os.Stat(binDir); err == nil && info.IsDir() {
		ffprobePath := filepath.Join(binDir, "ffprobe")
		if _, err := os.Stat(ffprobePath); os.IsNotExist(err) {
			writeExecutable(t, ffprobePath, "#!/bin/sh\nprintf '60.0\\n'\n")
		}
		vadToolPath := filepath.Join(binDir, "whisper-vad-speech-segments")
		if _, err := os.Stat(vadToolPath); os.IsNotExist(err) {
			writeFakeVADTool(t, binDir, vadLogWithSegments(1))
		}
	}
	return inputPath, modelPath
}

func writeFakeFFprobe(t *testing.T, binDir, duration string) {
	t.Helper()
	writeExecutable(t, filepath.Join(binDir, "ffprobe"), "#!/bin/sh\nprintf '"+duration+"\\n'\n")
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

func writeSuccessfulWhisperPipeline(t *testing.T, binDir, transcript string) {
	t.Helper()
	t.Setenv("WHISPER_TEST_TEXT", transcript)
	writeFakeVADTool(t, binDir, vadLogWithSegments(1))
	writeExecutable(t, filepath.Join(binDir, "whisper-cli"), `#!/bin/sh
set -eu
if [ -n "${WHISPER_OOM_ONCE_FILE:-}" ] && [ ! -e "$WHISPER_OOM_ONCE_FILE" ]; then
  : > "$WHISPER_OOM_ONCE_FILE"
  printf '%s\n' 'error: Insufficient Memory (00000008:kIOGPUCommandBufferCallbackErrorOutOfMemory)' >&2
  exit 1
fi
if [ -n "${WHISPER_CONCURRENCY_DIR:-}" ]; then
  while ! mkdir "$WHISPER_CONCURRENCY_DIR/lock" 2>/dev/null; do sleep 0.01; done
  active=0
  if [ -f "$WHISPER_CONCURRENCY_DIR/active" ]; then active=$(cat "$WHISPER_CONCURRENCY_DIR/active"); fi
  active=$((active + 1))
  printf '%s\n' "$active" > "$WHISPER_CONCURRENCY_DIR/active"
  maximum=0
  if [ -f "$WHISPER_CONCURRENCY_DIR/max" ]; then maximum=$(cat "$WHISPER_CONCURRENCY_DIR/max"); fi
  if [ "$active" -gt "$maximum" ]; then printf '%s\n' "$active" > "$WHISPER_CONCURRENCY_DIR/max"; fi
  rmdir "$WHISPER_CONCURRENCY_DIR/lock"
  cleanup_concurrency() {
    while ! mkdir "$WHISPER_CONCURRENCY_DIR/lock" 2>/dev/null; do sleep 0.01; done
    active=$(cat "$WHISPER_CONCURRENCY_DIR/active")
    printf '%s\n' "$((active - 1))" > "$WHISPER_CONCURRENCY_DIR/active"
    rmdir "$WHISPER_CONCURRENCY_DIR/lock"
  }
  trap cleanup_concurrency EXIT
  sleep 0.20
fi
case " $* " in
  *" --vad "*)
    printf '%s\n' \
      'whisper_vad_segments_from_probs: Final speech segments after filtering: 1' \
      'whisper_vad_segments_from_probs: VAD segment 0: start = 1.00, end = 2.00 (duration: 1.00)' >&2
    exit 0
    ;;
esac
probability="${WHISPER_TEST_PROBABILITY:-0.95}"
from=150
to=1150
case " $* " in
  *" -bs 8 "*) probability=0.95; from=1000; to=2000 ;;
esac
for argument do
  case "$argument" in
    *.wav)
      printf '{"transcription":[{"offsets":{"from":%s,"to":%s},"text":"%s","tokens":[{"text":"%s","offsets":{"from":%s,"to":%s},"p":%s}]}]}\n' "$from" "$to" "$WHISPER_TEST_TEXT" "$WHISPER_TEST_TEXT" "$from" "$to" "$probability" > "$argument.json"
      ;;
  esac
done
`)
}

func writeFakeVADTool(t *testing.T, binDir, log string) {
	t.Helper()
	content := "#!/bin/sh\nset -eu\nprintf '%s' " + shellSingleQuote(log) + " >&2\n"
	writeExecutable(t, filepath.Join(binDir, "whisper-vad-speech-segments"), content)
}

func shellSingleQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func quietToneAfterSilence(sampleRate int) []int16 {
	const (
		silenceSeconds = 8
		speechSeconds  = 2
		frequency      = 220
		amplitude      = 80
	)
	totalSamples := sampleRate * (silenceSeconds*2 + speechSeconds)
	samples := make([]int16, totalSamples)
	speechStart := sampleRate * silenceSeconds
	for index := 0; index < sampleRate*speechSeconds; index++ {
		phase := 2 * math.Pi * frequency * float64(index) / float64(sampleRate)
		samples[speechStart+index] = int16(amplitude * math.Sin(phase))
	}
	return samples
}

func writePCM16WAV(t *testing.T, path string, sampleRate int, samples []int16) {
	t.Helper()
	dataSize := len(samples) * 2
	content := make([]byte, 44+dataSize)
	copy(content[0:4], "RIFF")
	binary.LittleEndian.PutUint32(content[4:8], uint32(36+dataSize))
	copy(content[8:12], "WAVE")
	copy(content[12:16], "fmt ")
	binary.LittleEndian.PutUint32(content[16:20], 16)
	binary.LittleEndian.PutUint16(content[20:22], 1)
	binary.LittleEndian.PutUint16(content[22:24], 1)
	binary.LittleEndian.PutUint32(content[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(content[28:32], uint32(sampleRate*2))
	binary.LittleEndian.PutUint16(content[32:34], 2)
	binary.LittleEndian.PutUint16(content[34:36], 16)
	copy(content[36:40], "data")
	binary.LittleEndian.PutUint32(content[40:44], uint32(dataSize))
	for index, sample := range samples {
		binary.LittleEndian.PutUint16(content[44+index*2:], uint16(sample))
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readPCM16WAV(t *testing.T, path string) []int16 {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) < 12 || string(content[0:4]) != "RIFF" || string(content[8:12]) != "WAVE" {
		t.Fatalf("%q is not a RIFF/WAVE file", path)
	}
	for offset := 12; offset+8 <= len(content); {
		chunkSize := int(binary.LittleEndian.Uint32(content[offset+4 : offset+8]))
		chunkStart := offset + 8
		chunkEnd := chunkStart + chunkSize
		if chunkEnd > len(content) {
			t.Fatalf("%q has a truncated WAV chunk", path)
		}
		if string(content[offset:offset+4]) == "data" {
			if chunkSize%2 != 0 {
				t.Fatalf("%q has odd-sized PCM16 data", path)
			}
			samples := make([]int16, chunkSize/2)
			for index := range samples {
				samples[index] = int16(binary.LittleEndian.Uint16(content[chunkStart+index*2:]))
			}
			return samples
		}
		offset = chunkEnd + chunkSize%2
	}
	t.Fatalf("%q has no WAV data chunk", path)
	return nil
}

func pcmPeak(samples []int16) int {
	peak := 0
	for _, sample := range samples {
		value := int(sample)
		if value < 0 {
			value = -value
		}
		if value > peak {
			peak = value
		}
	}
	return peak
}

func firstSignalSample(samples []int16, threshold int) int {
	for index, sample := range samples {
		value := int(sample)
		if value < 0 {
			value = -value
		}
		if value > threshold {
			return index
		}
	}
	return -1
}
