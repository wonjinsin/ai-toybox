package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type transcriptionChunk struct {
	audioChunk
	Path string
}

func probeMediaDuration(ctx context.Context, ffprobePath, inputPath string) (time.Duration, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := exec.CommandContext(ctx, ffprobePath, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return 0, fmt.Errorf("probe media duration: %w", err)
		}
		return 0, fmt.Errorf("probe media duration: %w: %s", err, message)
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(stdout.String()), 64)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("parse media duration %q", strings.TrimSpace(stdout.String()))
	}
	return time.Duration(seconds * float64(time.Second)), nil
}

func detectSpeechSegments(ctx context.Context, whisperRunner *whisperCommandRunner, whisperPath, modelPath, audioPath, language, vadModelPath string) ([]speechSegment, error) {
	args := []string{
		"-m", modelPath,
		"-f", audioPath,
		"-l", language,
		"-mc", "0",
		"-sns",
		"--vad", "-vm", vadModelPath,
		"-vt", "0.35",
		"-vspd", "100",
		"-vsd", "500",
		"-vp", "250",
		"-vo", "0.20",
		"-vmsd", "20",
	}
	log, err := whisperRunner.Run(ctx, whisperPath, args)
	if err != nil {
		return nil, fmt.Errorf("detect speech with whisper.cpp: %w", err)
	}
	segments, err := parseSpeechSegments(log)
	if err != nil {
		return nil, fmt.Errorf("parse whisper.cpp VAD output: %w", err)
	}
	return segments, nil
}

func extractTranscriptionChunks(ctx context.Context, ffmpegPath, audioPath, directory string, chunks []audioChunk) ([]transcriptionChunk, error) {
	results := make([]transcriptionChunk, 0, len(chunks))
	for index, chunk := range chunks {
		path := filepath.Join(directory, fmt.Sprintf("chunk_%04d_%012d.wav", index, chunk.Start.Milliseconds()))
		args := []string{
			"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
			"-ss", formatSeconds(chunk.Start),
			"-t", formatSeconds(chunk.End - chunk.Start),
			"-i", audioPath,
			"-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le",
			path,
		}
		if err := runCommand(ctx, ffmpegPath, args); err != nil {
			return nil, fmt.Errorf("extract speech chunk %d: %w", index, err)
		}
		results = append(results, transcriptionChunk{audioChunk: chunk, Path: path})
	}
	return results, nil
}

func transcribeAudioChunks(ctx context.Context, whisperRunner *whisperCommandRunner, whisperPath, modelPath, language string, chunks []transcriptionChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	args := []string{
		"-m", modelPath,
		"-l", language,
		"-ojf",
		"-mc", "0",
		"-sns",
		"-ml", "30",
		"-sow",
	}
	for _, chunk := range chunks {
		args = append(args, chunk.Path)
	}
	if _, err := whisperRunner.Run(ctx, whisperPath, args); err != nil {
		return fmt.Errorf("transcribe speech chunks with whisper.cpp: %w", err)
	}
	return nil
}

func transcribeRetryAudioChunks(ctx context.Context, whisperRunner *whisperCommandRunner, whisperPath, modelPath, language string, chunks []transcriptionChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	args := []string{
		"-m", modelPath,
		"-l", language,
		"-ojf",
		"-mc", "0",
		"-sns",
		"-ml", "30",
		"-sow",
		"-tp", "0",
		"-bs", "8",
	}
	for _, chunk := range chunks {
		args = append(args, chunk.Path)
	}
	if _, err := whisperRunner.Run(ctx, whisperPath, args); err != nil {
		return fmt.Errorf("retry low-confidence speech chunks with whisper.cpp: %w", err)
	}
	return nil
}

func loadTranscriptionCues(chunks []transcriptionChunk) ([]subtitleCue, error) {
	cues := make([]subtitleCue, 0, len(chunks))
	for index, chunk := range chunks {
		jsonPath := chunk.Path + ".json"
		content, err := os.ReadFile(jsonPath)
		if err != nil {
			return nil, fmt.Errorf("read whisper JSON %q: %w", jsonPath, err)
		}
		chunkCues, err := parseWhisperJSONForOrigin(content, transcriptionOrigin{
			Index: index,
			Start: chunk.Start,
			End:   chunk.End,
		})
		if err != nil {
			return nil, fmt.Errorf("parse whisper JSON %q: %w", jsonPath, err)
		}
		cues = append(cues, chunkCues...)
	}
	return cues, nil
}

func validateSubtitleCues(cues []subtitleCue, mediaDuration time.Duration) error {
	for index, cue := range cues {
		if strings.TrimSpace(cue.Text) == "" {
			return fmt.Errorf("subtitle cue %d has empty text", index+1)
		}
		if cue.Start < 0 || cue.End <= cue.Start {
			return fmt.Errorf("subtitle cue %d has invalid timing %s --> %s", index+1, cue.Start, cue.End)
		}
		if mediaDuration > 0 && cue.End > mediaDuration {
			return fmt.Errorf("subtitle cue %d exceeds media duration", index+1)
		}
		if index > 0 && cue.Start < cues[index-1].End {
			return fmt.Errorf("subtitle cues %d and %d overlap", index, index+1)
		}
	}
	return nil
}

func formatSeconds(value time.Duration) string {
	return strconv.FormatFloat(value.Seconds(), 'f', 3, 64)
}
