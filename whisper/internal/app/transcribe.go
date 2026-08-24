package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const audioNormalizationFilter = "dynaudnorm=f=500:g=31:p=0.95:m=10:b=1"

var (
	vadSegmentLogPattern   = regexp.MustCompile(`VAD segment \d+: start = ([0-9]+(?:\.[0-9]+)?), end = ([0-9]+(?:\.[0-9]+)?)`)
	vadSegmentCountPattern = regexp.MustCompile(`Final speech segments after filtering: ([0-9]+)`)
)

func Transcribe(ctx context.Context, options Options) (string, error) {
	return transcribeWithProgress(ctx, options, nil)
}

func transcribeWithProgress(ctx context.Context, options Options, progress progressReporter) (string, error) {
	return transcribeWithProgressUsingRunner(ctx, options, progress, newWhisperCommandRunner(1, runCommandCaptureStderr))
}

func transcribeWithProgressUsingRunner(ctx context.Context, options Options, progress progressReporter, whisperRunner *whisperCommandRunner) (string, error) {
	overallStartedAt := time.Now()
	if err := validateLanguage(options.Language); err != nil {
		return "", err
	}
	if err := validateFormat(options.Format); err != nil {
		return "", err
	}

	inputPath, err := requireRegularFile(options.InputPath, "input")
	if err != nil {
		return "", err
	}
	outputDir := options.OutputDir
	if outputDir == "" {
		outputDir = filepath.Dir(inputPath)
	}
	outputDir, err = filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("resolve output directory: %w", err)
	}
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputBase := filepath.Join(outputDir, baseName)
	outputPath := outputBase + "." + options.Format
	if !options.Force {
		if _, err := os.Stat(outputPath); err == nil {
			return "", newOutputExistsError(outputPath)
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("inspect output file %q: %w", outputPath, err)
		}
	}

	modelPath, err := resolveModelPath(options.ModelPath)
	if err != nil {
		return "", err
	}
	vadModelPath, err := resolveRequiredVADModelPath(options.VADModelPath, modelPath)
	if err != nil {
		return "", err
	}
	corrections, err := loadCorrections(options.CorrectionsPath)
	if err != nil {
		return "", err
	}

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", errors.New("ffmpeg not found: install it with 'brew install ffmpeg'")
	}
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		return "", errors.New("ffprobe not found: install it with 'brew install ffmpeg'")
	}
	whisperPath, err := exec.LookPath("whisper-cli")
	if err != nil {
		return "", errors.New("whisper-cli not found: install it with 'brew install whisper-cpp'")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", fmt.Errorf("create output directory %q: %w", outputDir, err)
	}
	var fingerprint incrementalFingerprint
	var restoredCheckpoint incrementalCheckpoint
	hasRestoredCheckpoint := false
	if options.Format == "srt" {
		fingerprint, err = newIncrementalFingerprint(inputPath, modelPath, vadModelPath, whisperPath, ffmpegPath, ffprobePath, options.Language)
		if err != nil {
			return "", err
		}
		_, checkpointPath := incrementalArtifactPaths(outputBase)
		restoredCheckpoint, hasRestoredCheckpoint, err = loadIncrementalCheckpoint(checkpointPath, fingerprint)
		if err != nil {
			return "", err
		}
	}
	outputReservation, err := os.CreateTemp(outputDir, ".whisper-local-output-*")
	if err != nil {
		return "", fmt.Errorf("reserve temporary transcript: %w", err)
	}
	temporaryOutputBase := outputReservation.Name()
	if err := outputReservation.Close(); err != nil {
		os.Remove(temporaryOutputBase)
		return "", fmt.Errorf("close temporary transcript reservation: %w", err)
	}
	defer os.Remove(temporaryOutputBase)
	temporaryOutputPath := temporaryOutputBase + "." + options.Format

	tempFile, err := os.CreateTemp("", "whisper-local-*.wav")
	if err != nil {
		return "", fmt.Errorf("create temporary audio file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)
	if err := tempFile.Close(); err != nil {
		return "", fmt.Errorf("close temporary audio file: %w", err)
	}

	ffmpegArgs := []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-i", inputPath,
		"-vn", "-af", audioNormalizationFilter,
		"-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le",
		tempPath,
	}
	if err := runWithProgress(progress, inputPath, "오디오 추출 중...", overallStartedAt, func() error {
		return runCommand(ctx, ffmpegPath, ffmpegArgs)
	}); err != nil {
		return "", fmt.Errorf("extract audio with ffmpeg: %w", err)
	}

	var mediaDuration time.Duration
	var chunks []audioChunk
	var rawCues []subtitleCue
	completedChunks := 0
	if hasRestoredCheckpoint {
		mediaDuration = restoredCheckpoint.MediaDuration
		chunks = audioChunksFromCheckpoint(restoredCheckpoint.Chunks)
		rawCues = subtitleCuesFromCheckpointCues(restoredCheckpoint.Cues)
		completedChunks = restoredCheckpoint.CompletedChunks
		reportProgress(progress, inputPath, fmt.Sprintf("체크포인트 재개: %d/%d개", completedChunks, len(chunks)))
		if completedChunks > 0 {
			if err := persistIncrementalProgress(outputBase, options.Language, corrections, fingerprint, mediaDuration, chunks, completedChunks, rawCues); err != nil {
				return "", err
			}
		}
	} else {
		mediaDuration, err = probeMediaDuration(ctx, ffprobePath, inputPath)
		if err != nil {
			return "", err
		}
		var speechSegments []speechSegment
		if err := runWithProgress(progress, inputPath, "음성 구간 탐지 중...", overallStartedAt, func() error {
			var operationErr error
			speechSegments, operationErr = detectSpeechSegments(ctx, whisperRunner, whisperPath, modelPath, tempPath, options.Language, vadModelPath)
			return operationErr
		}); err != nil {
			return "", err
		}
		chunks = buildSpeechChunks(speechSegments, mediaDuration)
		if options.Format == "srt" {
			_, checkpointPath := incrementalArtifactPaths(outputBase)
			checkpoint := newIncrementalCheckpointWithFingerprint(fingerprint, mediaDuration, chunks, 0, nil)
			if err := persistIncrementalCheckpoint(checkpointPath, checkpoint); err != nil {
				return "", err
			}
		}
	}
	chunkDirectory, err := os.MkdirTemp("", "whisper-local-chunks-*")
	if err != nil {
		return "", fmt.Errorf("create temporary chunk directory: %w", err)
	}
	defer os.RemoveAll(chunkDirectory)
	var transcriptionChunks []transcriptionChunk
	chunksToExtract := chunks
	firstChunkIndex := 0
	if options.Format == "srt" {
		chunksToExtract = chunks[completedChunks:]
		firstChunkIndex = completedChunks
	}
	if err := runWithProgress(progress, inputPath, fmt.Sprintf("음성 조각 생성: %d개", len(chunksToExtract)), overallStartedAt, func() error {
		var operationErr error
		transcriptionChunks, operationErr = extractTranscriptionChunksFromIndex(ctx, ffmpegPath, tempPath, chunkDirectory, chunksToExtract, firstChunkIndex)
		return operationErr
	}); err != nil {
		return "", err
	}
	if options.Format == "srt" {
		if err := runWithProgress(progress, inputPath, fmt.Sprintf("전사 중: %d개 (완료 %d/%d개)", len(transcriptionChunks), completedChunks, len(chunks)), overallStartedAt, func() error {
			for start := 0; start < len(transcriptionChunks); start += transcriptionBatchSize {
				end := min(start+transcriptionBatchSize, len(transcriptionChunks))
				if err := transcribeAudioChunks(ctx, whisperRunner, whisperPath, modelPath, options.Language, transcriptionChunks[start:end]); err != nil {
					return err
				}
				globalStart := completedChunks + start
				globalEnd := completedChunks + end
				batchCues, err := loadTranscriptionCuesFromIndex(transcriptionChunks[start:end], globalStart)
				if err != nil {
					return err
				}
				rawCues = append(rawCues, batchCues...)
				if err := persistIncrementalProgress(outputBase, options.Language, corrections, fingerprint, mediaDuration, chunks, globalEnd, rawCues); err != nil {
					return err
				}
				reportProgress(progress, inputPath, fmt.Sprintf("전사 진행: %d/%d개 (partial SRT 저장됨)", globalEnd, len(chunks)))
			}
			return nil
		}); err != nil {
			return "", err
		}
	} else {
		if err := runWithProgress(progress, inputPath, fmt.Sprintf("전사 중: %d개", len(transcriptionChunks)), overallStartedAt, func() error {
			return transcribeAudioChunks(ctx, whisperRunner, whisperPath, modelPath, options.Language, transcriptionChunks)
		}); err != nil {
			return "", err
		}
		rawCues, err = loadTranscriptionCues(transcriptionChunks)
		if err != nil {
			return "", err
		}
	}
	rawCues = reconcileChunkBoundaries(rawCues)
	retryCount := countLowConfidenceCues(rawCues)
	if err := runWithProgress(progress, inputPath, fmt.Sprintf("저신뢰 구간 재시도: %d개", retryCount), overallStartedAt, func() error {
		var operationErr error
		rawCues, operationErr = retryLowConfidenceCues(ctx, whisperRunner, ffmpegPath, whisperPath, modelPath, options.Language, tempPath, chunkDirectory, rawCues, mediaDuration)
		return operationErr
	}); err != nil {
		return "", err
	}
	reportProgress(progress, inputPath, "자막 정리 및 검증 중...")
	cues := cleanSubtitleCues(rawCues, options.Language, corrections, mediaDuration)
	if err := validateSubtitleCues(cues, mediaDuration); err != nil {
		return "", err
	}
	transcript, err := renderTranscript(cues, options.Format, options.Language)
	if err != nil {
		return "", err
	}
	reportProgress(progress, inputPath, "결과 저장 중...")
	if err := writeAndPublishTranscript(temporaryOutputPath, outputPath, []byte(transcript), options.Force, os.Link); err != nil {
		return "", err
	}

	reportProgress(progress, inputPath, "처리 완료")
	return outputPath, nil
}

func parseSpeechSegments(log string) ([]speechSegment, error) {
	countMatch := vadSegmentCountPattern.FindStringSubmatch(log)
	if countMatch == nil {
		return nil, errors.New("VAD segment summary not found in whisper.cpp output")
	}
	expectedCount, err := strconv.Atoi(countMatch[1])
	if err != nil {
		return nil, fmt.Errorf("parse VAD segment count %q: %w", countMatch[1], err)
	}

	matches := vadSegmentLogPattern.FindAllStringSubmatch(log, -1)
	segments := make([]speechSegment, 0, len(matches))
	for _, match := range matches {
		startSeconds, startErr := strconv.ParseFloat(match[1], 64)
		endSeconds, endErr := strconv.ParseFloat(match[2], 64)
		if startErr != nil || endErr != nil || endSeconds <= startSeconds {
			continue
		}
		segments = append(segments, speechSegment{
			Start: time.Duration(startSeconds * float64(time.Second)),
			End:   time.Duration(endSeconds * float64(time.Second)),
		})
	}
	if len(segments) != expectedCount {
		return nil, fmt.Errorf("parsed %d VAD segments, want %d from whisper.cpp summary", len(segments), expectedCount)
	}
	return segments, nil
}

func resolveModelPath(explicitPath string) (string, error) {
	if explicitPath != "" {
		return requireRegularFile(explicitPath, "model")
	}
	if environmentPath := os.Getenv("WHISPER_MODEL"); environmentPath != "" {
		return requireRegularFile(environmentPath, "model from WHISPER_MODEL")
	}

	candidates := make([]string, 0, 3)
	if workingDirectory, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(workingDirectory, "models", "ggml-large-v3.bin"))
	}
	if executablePath, err := os.Executable(); err == nil {
		if resolvedPath, resolveErr := filepath.EvalSymlinks(executablePath); resolveErr == nil {
			executablePath = resolvedPath
		}
		candidates = append(candidates, filepath.Join(filepath.Dir(executablePath), "..", "models", "ggml-large-v3.bin"))
	}
	if cacheDirectory, err := os.UserCacheDir(); err == nil {
		candidates = append(candidates, filepath.Join(cacheDirectory, "whisper-local", "models", "ggml-large-v3.bin"))
	}

	for _, candidate := range candidates {
		if modelPath, err := requireRegularFile(candidate, "model"); err == nil {
			return modelPath, nil
		}
	}
	return "", fmt.Errorf("Whisper model not found; tried: %s; use --model or WHISPER_MODEL", strings.Join(candidates, ", "))
}

func requireRegularFile(path, label string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%s file path is required", label)
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve %s file %q: %w", label, path, err)
	}
	info, err := os.Stat(absolutePath)
	if err != nil {
		return "", fmt.Errorf("open %s file %q: %w", label, absolutePath, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s path %q is not a regular file", label, absolutePath)
	}
	return absolutePath, nil
}

// resolveVADModelPath looks for the required Silero VAD model next to the Whisper model.
func resolveVADModelPath(modelPath string) string {
	candidate := filepath.Join(filepath.Dir(modelPath), "ggml-silero-v6.2.0.bin")
	if vadPath, err := requireRegularFile(candidate, "VAD model"); err == nil {
		return vadPath
	}
	return ""
}

func resolveRequiredVADModelPath(explicitPath, modelPath string) (string, error) {
	if explicitPath != "" {
		return requireRegularFile(explicitPath, "VAD model")
	}
	if environmentPath := os.Getenv("WHISPER_VAD_MODEL"); environmentPath != "" {
		return requireRegularFile(environmentPath, "VAD model from WHISPER_VAD_MODEL")
	}
	if resolved := resolveVADModelPath(modelPath); resolved != "" {
		return resolved, nil
	}
	return "", fmt.Errorf("VAD model not found next to Whisper model %q; use --vad-model or WHISPER_VAD_MODEL", modelPath)
}

func runCommand(ctx context.Context, executable string, args []string) error {
	_, err := runCommandCaptureStderr(ctx, executable, args)
	return err
}

func runCommandCaptureStderr(ctx context.Context, executable string, args []string) (string, error) {
	var stderr bytes.Buffer
	command := exec.CommandContext(ctx, executable, args...)
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return "", err
		}
		return "", fmt.Errorf("%w: %s", err, message)
	}
	return stderr.String(), nil
}
