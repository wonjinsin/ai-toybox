package ffmpeg

import (
	"bytes"
	"context"
	"encoding/csv"
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

	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/out/process"
	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

const (
	audioNormalizationFilter      = "dynaudnorm=f=500:g=31:p=0.95:m=100:b=1"
	vadStateResetInterval         = 15 * time.Second
	transcriptionChunkWorkerCount = 4
)

type VADChunk struct {
	Path  string
	Start time.Duration
	End   time.Duration
}

type Processor struct {
	ffmpegPath  string
	ffprobePath string
	inputPath   string
	audioPath   string
	directory   string
}

func New(ffmpegPath, ffprobePath, inputPath, audioPath, directory string) *Processor {
	return &Processor{ffmpegPath: ffmpegPath, ffprobePath: ffprobePath, inputPath: inputPath, audioPath: audioPath, directory: directory}
}

func (processor *Processor) Normalize(ctx context.Context) error {
	args := []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-i", processor.inputPath,
		"-vn", "-af", audioNormalizationFilter,
		"-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le",
		processor.audioPath,
	}
	if err := process.Run(ctx, processor.ffmpegPath, args); err != nil {
		return fmt.Errorf("extract audio with ffmpeg: %w", err)
	}
	return nil
}

func (processor *Processor) Duration(ctx context.Context) (time.Duration, error) {
	return probeMediaDuration(ctx, processor.ffprobePath, processor.inputPath)
}

func (processor *Processor) Extract(ctx context.Context, chunks []domain.Chunk, firstIndex int) (portout.AudioBatch, error) {
	if len(chunks) == 0 {
		return portout.AudioBatch{}, ctx.Err()
	}
	directory, err := os.MkdirTemp(processor.directory, "batch-*")
	if err != nil {
		return portout.AudioBatch{}, fmt.Errorf("create transcription batch directory: %w", err)
	}
	release := func() error {
		if err := os.RemoveAll(directory); err != nil {
			return fmt.Errorf("remove transcription batch directory: %w", err)
		}
		return nil
	}
	audio, err := extractTranscriptionChunksFromIndexUsing(ctx, process.Run, processor.ffmpegPath, processor.audioPath, directory, chunks, firstIndex)
	if err != nil {
		return portout.AudioBatch{}, errors.Join(err, release())
	}
	return portout.AudioBatch{Chunks: audio, Release: release}, nil
}

var _ portout.AudioProcessor = (*Processor)(nil)

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

func SegmentForVAD(ctx context.Context, ffmpegPath, audioPath, directory string) ([]VADChunk, error) {
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
	if err := process.Run(ctx, ffmpegPath, args); err != nil {
		return nil, fmt.Errorf("split normalized audio for periodic VAD reset: %w", err)
	}

	manifest, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("open VAD chunk manifest: %w", err)
	}
	defer manifest.Close()

	reader := csv.NewReader(manifest)
	chunks := make([]VADChunk, 0)
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
		chunks = append(chunks, VADChunk{
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

func extractTranscriptionChunksFromIndexUsing(ctx context.Context, execute func(context.Context, string, []string) error, ffmpegPath, audioPath, directory string, chunks []domain.Chunk, firstIndex int) ([]portout.ChunkAudio, error) {
	results := make([]portout.ChunkAudio, len(chunks))
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
				results[index] = portout.ChunkAudio{Chunk: chunk, Index: firstIndex + index, Path: path}
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

func formatSeconds(value time.Duration) string {
	return strconv.FormatFloat(value.Seconds(), 'f', 3, 64)
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
