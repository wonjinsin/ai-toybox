package whisper

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/out/ffmpeg"
	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/out/process"
	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

const vadWorkerCount = 4

var (
	vadSegmentLogPattern   = regexp.MustCompile(`VAD segment \d+: start = ([0-9]+(?:\.[0-9]+)?), end = ([0-9]+(?:\.[0-9]+)?)`)
	vadSegmentCountPattern = regexp.MustCompile(`Final speech segments after filtering: ([0-9]+)`)
)

type Detector struct {
	ffmpegPath   string
	vadToolPath  string
	audioPath    string
	vadModelPath string
	execute      CommandExecutor
}

func NewDetector(ffmpegPath, vadToolPath, audioPath, vadModelPath string, execute CommandExecutor) *Detector {
	if execute == nil {
		execute = process.RunCaptureStderr
	}
	return &Detector{ffmpegPath: ffmpegPath, vadToolPath: vadToolPath, audioPath: audioPath, vadModelPath: vadModelPath, execute: execute}
}

func (detector *Detector) Detect(ctx context.Context) ([]domain.SpeechSegment, error) {
	return detectSpeechSegments(ctx, detector.execute, detector.ffmpegPath, detector.vadToolPath, detector.audioPath, detector.vadModelPath)
}

var _ portout.SpeechDetector = (*Detector)(nil)

func detectSpeechSegments(ctx context.Context, execute CommandExecutor, ffmpegPath, vadToolPath, audioPath, vadModelPath string) ([]domain.SpeechSegment, error) {
	directory, err := os.MkdirTemp("", "whisper-local-vad-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary VAD directory: %w", err)
	}
	defer os.RemoveAll(directory)

	chunks, err := ffmpeg.SegmentForVAD(ctx, ffmpegPath, audioPath, directory)
	if err != nil {
		return nil, err
	}
	chunkResults := make([][]domain.SpeechSegment, len(chunks))
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
				adjusted := make([]domain.SpeechSegment, 0, len(segments))
				for _, segment := range segments {
					adjusted = append(adjusted, domain.SpeechSegment{
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

	results := make([]domain.SpeechSegment, 0)
	for _, segments := range chunkResults {
		results = append(results, segments...)
	}
	return results, nil
}

func detectSpeechSegmentsInFile(ctx context.Context, execute CommandExecutor, vadToolPath, audioPath, vadModelPath string) ([]domain.SpeechSegment, error) {
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

func parseSpeechSegments(log string) ([]domain.SpeechSegment, error) {
	countMatch := vadSegmentCountPattern.FindStringSubmatch(log)
	if countMatch == nil {
		return nil, errors.New("VAD segment summary not found in whisper.cpp output")
	}
	expectedCount, err := strconv.Atoi(countMatch[1])
	if err != nil {
		return nil, fmt.Errorf("parse VAD segment count %q: %w", countMatch[1], err)
	}

	matches := vadSegmentLogPattern.FindAllStringSubmatch(log, -1)
	segments := make([]domain.SpeechSegment, 0, len(matches))
	for _, match := range matches {
		startSeconds, startErr := strconv.ParseFloat(match[1], 64)
		endSeconds, endErr := strconv.ParseFloat(match[2], 64)
		if startErr != nil || endErr != nil || endSeconds <= startSeconds {
			continue
		}
		segments = append(segments, domain.SpeechSegment{
			Start: time.Duration(startSeconds * float64(time.Second)),
			End:   time.Duration(endSeconds * float64(time.Second)),
		})
	}
	if len(segments) != expectedCount {
		return nil, fmt.Errorf("parsed %d VAD segments, want %d from whisper.cpp summary", len(segments), expectedCount)
	}
	return segments, nil
}
