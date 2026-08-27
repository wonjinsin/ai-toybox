package app

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	vadStateResetInterval             = 15 * time.Second
	vadWorkerCount                    = 4
	transcriptionChunkWorkerCount     = 4
	transcriptionProgressPollInterval = 250 * time.Millisecond
)

type vadAudioChunk struct {
	Path  string
	Start time.Duration
	End   time.Duration
}

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

func detectSpeechSegments(ctx context.Context, execute commandExecutor, ffmpegPath, vadToolPath, audioPath, vadModelPath string) ([]speechSegment, error) {
	directory, err := os.MkdirTemp("", "whisper-local-vad-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary VAD directory: %w", err)
	}
	defer os.RemoveAll(directory)

	chunks, err := segmentAudioForVAD(ctx, ffmpegPath, audioPath, directory)
	if err != nil {
		return nil, err
	}
	chunkResults := make([][]speechSegment, len(chunks))
	workContext, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan int)
	workerCount := min(vadWorkerCount, len(chunks))
	var workers sync.WaitGroup
	var firstError error
	var errorOnce sync.Once
	for range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				chunk := chunks[index]
				segments, detectErr := detectSpeechSegmentsInFile(workContext, execute, vadToolPath, chunk.Path, vadModelPath)
				if detectErr != nil {
					errorOnce.Do(func() {
						firstError = fmt.Errorf("detect speech in VAD chunk %d at %s: %w", index+1, formatSeconds(chunk.Start), detectErr)
						cancel()
					})
					continue
				}
				adjusted := make([]speechSegment, 0, len(segments))
				for _, segment := range segments {
					adjusted = append(adjusted, speechSegment{
						Start: chunk.Start + segment.Start,
						End:   min(chunk.Start+segment.End, chunk.End),
					})
				}
				chunkResults[index] = adjusted
			}
		}()
	}

enqueue:
	for index := range chunks {
		select {
		case jobs <- index:
		case <-workContext.Done():
			break enqueue
		}
	}
	close(jobs)
	workers.Wait()
	if firstError != nil {
		return nil, firstError
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	results := make([]speechSegment, 0)
	for _, segments := range chunkResults {
		results = append(results, segments...)
	}
	return results, nil
}

func segmentAudioForVAD(ctx context.Context, ffmpegPath, audioPath, directory string) ([]vadAudioChunk, error) {
	manifestPath := filepath.Join(directory, "vad-chunks.csv")
	outputPattern := filepath.Join(directory, "vad_%05d.wav")
	args := []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-i", audioPath,
		"-map", "0:a:0", "-c:a", "copy",
		"-f", "segment", "-segment_time", formatSeconds(vadStateResetInterval),
		"-reset_timestamps", "1",
		"-segment_list", manifestPath, "-segment_list_type", "csv",
		outputPattern,
	}
	if err := runCommand(ctx, ffmpegPath, args); err != nil {
		return nil, fmt.Errorf("split normalized audio for periodic VAD reset: %w", err)
	}

	manifest, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("open VAD chunk manifest: %w", err)
	}
	defer manifest.Close()

	reader := csv.NewReader(manifest)
	chunks := make([]vadAudioChunk, 0)
	for {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("read VAD chunk manifest: %w", readErr)
		}
		if len(record) != 3 {
			return nil, fmt.Errorf("read VAD chunk manifest: got %d fields, want 3", len(record))
		}
		name := filepath.Clean(record[0])
		if filepath.IsAbs(name) || name != filepath.Base(name) {
			return nil, fmt.Errorf("read VAD chunk manifest: unsafe chunk path %q", record[0])
		}
		startSeconds, startErr := strconv.ParseFloat(record[1], 64)
		endSeconds, endErr := strconv.ParseFloat(record[2], 64)
		if startErr != nil || endErr != nil || !isFinite(startSeconds) || !isFinite(endSeconds) || startSeconds < 0 || endSeconds <= startSeconds {
			return nil, fmt.Errorf("read VAD chunk manifest: invalid interval %q to %q", record[1], record[2])
		}
		path, pathErr := requireRegularFile(filepath.Join(directory, name), "VAD audio chunk")
		if pathErr != nil {
			return nil, pathErr
		}
		chunks = append(chunks, vadAudioChunk{
			Path:  path,
			Start: time.Duration(startSeconds * float64(time.Second)),
			End:   time.Duration(endSeconds * float64(time.Second)),
		})
	}
	if len(chunks) == 0 {
		return nil, errors.New("split normalized audio for periodic VAD reset: no audio chunks created")
	}
	return chunks, nil
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func detectSpeechSegmentsInFile(ctx context.Context, execute commandExecutor, vadToolPath, audioPath, vadModelPath string) ([]speechSegment, error) {
	// whisper.cpp 1.9.1 routes the dedicated tool's min-silence option into
	// min-speech. Omitting it preserves 100 ms speech detection; chunk grouping
	// already joins the resulting short gaps.
	args := []string{
		"-t", "1",
		"-f", audioPath,
		"-vm", vadModelPath,
		"-vt", "0.35",
		"--vad-min-speech-duration-ms", "100",
		"--vad-speech-pad-ms", "250",
		"--vad-samples-overlap", "0.20",
		"--vad-max-speech-duration-s", "20",
	}
	log, err := execute(ctx, vadToolPath, args)
	if err != nil {
		return nil, fmt.Errorf("detect speech with whisper.cpp VAD tool: %w", err)
	}
	segments, err := parseSpeechSegments(log)
	if err != nil {
		return nil, fmt.Errorf("parse whisper.cpp VAD output: %w", err)
	}
	return segments, nil
}

func extractTranscriptionChunks(ctx context.Context, ffmpegPath, audioPath, directory string, chunks []audioChunk) ([]transcriptionChunk, error) {
	return extractTranscriptionChunksFromIndex(ctx, ffmpegPath, audioPath, directory, chunks, 0)
}

func extractTranscriptionChunksFromIndex(ctx context.Context, ffmpegPath, audioPath, directory string, chunks []audioChunk, firstIndex int) ([]transcriptionChunk, error) {
	return extractTranscriptionChunksFromIndexUsing(ctx, runCommand, ffmpegPath, audioPath, directory, chunks, firstIndex)
}

func extractTranscriptionChunksFromIndexUsing(ctx context.Context, execute func(context.Context, string, []string) error, ffmpegPath, audioPath, directory string, chunks []audioChunk, firstIndex int) ([]transcriptionChunk, error) {
	results := make([]transcriptionChunk, len(chunks))
	workContext, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan int)
	workerCount := min(transcriptionChunkWorkerCount, len(chunks))
	var workers sync.WaitGroup
	var firstError error
	var errorOnce sync.Once
	for range workerCount {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				chunk := chunks[index]
				path := filepath.Join(directory, fmt.Sprintf("chunk_%04d_%012d.wav", firstIndex+index, chunk.Start.Milliseconds()))
				args := []string{
					"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
					"-ss", formatSeconds(chunk.Start),
					"-t", formatSeconds(chunk.End - chunk.Start),
					"-i", audioPath,
					"-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le",
					path,
				}
				if err := execute(workContext, ffmpegPath, args); err != nil {
					errorOnce.Do(func() {
						firstError = fmt.Errorf("extract speech chunk %d: %w", firstIndex+index, err)
						cancel()
					})
					continue
				}
				results[index] = transcriptionChunk{audioChunk: chunk, Path: path}
			}
		}()
	}

enqueue:
	for index := range chunks {
		select {
		case jobs <- index:
		case <-workContext.Done():
			break enqueue
		}
	}
	close(jobs)
	workers.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if firstError != nil {
		return nil, firstError
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
		"-bs", "1",
	}
	for _, chunk := range chunks {
		args = append(args, chunk.Path)
	}
	if _, err := whisperRunner.Run(ctx, whisperPath, args); err != nil {
		return fmt.Errorf("transcribe speech chunks with whisper.cpp: %w", err)
	}
	return nil
}

func transcribeAudioChunksWithProgress(ctx context.Context, whisperRunner *whisperCommandRunner, whisperPath, modelPath, language string, chunks []transcriptionChunk, reportCompleted func(int)) error {
	return transcribeAudioChunksWithProgressInterval(ctx, whisperRunner, whisperPath, modelPath, language, chunks, transcriptionProgressPollInterval, reportCompleted)
}

func transcribeAudioChunksWithProgressInterval(ctx context.Context, whisperRunner *whisperCommandRunner, whisperPath, modelPath, language string, chunks []transcriptionChunk, interval time.Duration, reportCompleted func(int)) error {
	if len(chunks) == 0 {
		return nil
	}
	if reportCompleted == nil || interval <= 0 {
		return transcribeAudioChunks(ctx, whisperRunner, whisperPath, modelPath, language, chunks)
	}

	transcriptionDone := make(chan error, 1)
	go func() {
		transcriptionDone <- transcribeAudioChunks(ctx, whisperRunner, whisperPath, modelPath, language, chunks)
	}()

	completed := 0
	reportAvailableOutputs := func() {
		previousCompleted := completed
		for completed < len(chunks) {
			content, err := os.ReadFile(chunks[completed].Path + ".json")
			if err != nil || !json.Valid(content) {
				break
			}
			completed++
		}
		if completed > previousCompleted {
			reportCompleted(completed)
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case err := <-transcriptionDone:
			reportAvailableOutputs()
			return err
		case <-ticker.C:
			reportAvailableOutputs()
		}
	}
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
	return loadTranscriptionCuesFromIndex(chunks, 0)
}

func loadTranscriptionCuesFromIndex(chunks []transcriptionChunk, firstIndex int) ([]subtitleCue, error) {
	cues := make([]subtitleCue, 0, len(chunks))
	for index, chunk := range chunks {
		jsonPath := chunk.Path + ".json"
		content, err := os.ReadFile(jsonPath)
		if err != nil {
			return nil, fmt.Errorf("read whisper JSON %q: %w", jsonPath, err)
		}
		chunkCues, err := parseWhisperJSONForOrigin(content, transcriptionOrigin{
			Index: firstIndex + index,
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
