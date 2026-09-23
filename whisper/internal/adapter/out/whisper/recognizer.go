package whisper

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

const transcriptionProgressPollInterval = 250 * time.Millisecond

type Recognizer struct {
	whisperPath string
	modelPath   string
	language    string
	runner      *CommandRunner
}

func NewRecognizer(whisperPath, modelPath, language string, runner *CommandRunner) *Recognizer {
	return &Recognizer{whisperPath: whisperPath, modelPath: modelPath, language: language, runner: runner}
}

func (recognizer *Recognizer) Recognize(ctx context.Context, chunks []portout.ChunkAudio, mode portout.RecognitionMode, reportCompleted func(int)) ([][]domain.Cue, error) {
	var err error
	switch mode {
	case portout.RecognitionStandard:
		err = transcribeAudioChunksWithProgress(ctx, recognizer.runner, recognizer.whisperPath, recognizer.modelPath, recognizer.language, chunks, reportCompleted)
	case portout.RecognitionRetry:
		err = transcribeRetryAudioChunks(ctx, recognizer.runner, recognizer.whisperPath, recognizer.modelPath, recognizer.language, chunks)
	default:
		return nil, fmt.Errorf("unsupported recognition mode %d", mode)
	}
	if err != nil {
		return nil, err
	}
	return loadTranscriptionCues(chunks)
}

var _ portout.Recognizer = (*Recognizer)(nil)

func transcribeAudioChunks(ctx context.Context, whisperRunner *CommandRunner, whisperPath, modelPath, language string, chunks []portout.ChunkAudio) error {
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

func transcribeAudioChunksWithProgress(ctx context.Context, whisperRunner *CommandRunner, whisperPath, modelPath, language string, chunks []portout.ChunkAudio, reportCompleted func(int)) error {
	return transcribeAudioChunksWithProgressInterval(ctx, whisperRunner, whisperPath, modelPath, language, chunks, transcriptionProgressPollInterval, reportCompleted)
}

func transcribeAudioChunksWithProgressInterval(ctx context.Context, whisperRunner *CommandRunner, whisperPath, modelPath, language string, chunks []portout.ChunkAudio, interval time.Duration, reportCompleted func(int)) error {
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

func transcribeRetryAudioChunks(ctx context.Context, whisperRunner *CommandRunner, whisperPath, modelPath, language string, chunks []portout.ChunkAudio) error {
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

func formatSeconds(value time.Duration) string {
	return strconv.FormatFloat(value.Seconds(), 'f', 3, 64)
}

func loadTranscriptionCues(chunks []portout.ChunkAudio) ([][]domain.Cue, error) {
	results := make([][]domain.Cue, len(chunks))
	for index, chunk := range chunks {
		jsonPath := chunk.Path + ".json"
		content, err := os.ReadFile(jsonPath)
		if err != nil {
			return nil, fmt.Errorf("read whisper JSON %q: %w", jsonPath, err)
		}
		cues, err := parseWhisperJSONForOrigin(content, domain.Origin{Index: chunk.Index, Start: chunk.Chunk.Start, End: chunk.Chunk.End})
		if err != nil {
			return nil, fmt.Errorf("parse whisper JSON %q: %w", jsonPath, err)
		}
		results[index] = cues
	}
	return results, nil
}
