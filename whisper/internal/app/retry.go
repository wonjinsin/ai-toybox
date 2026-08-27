package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	retrySubtitleProbability    = 0.50
	retryContextDuration        = time.Second
	minimumRetryConfidenceGain  = 0.05
	retryTranscriptionBatchSize = 128
)

type retryProgressFunc func(nextCursor int, cues []subtitleCue) error

type retryBatchItem struct {
	index  int
	cue    subtitleCue
	window audioChunk
}

func countLowConfidenceCues(cues []subtitleCue) int {
	count := 0
	for _, cue := range cues {
		if cue.Probability < retrySubtitleProbability {
			count++
		}
	}
	return count
}

func retryLowConfidenceCues(ctx context.Context, whisperRunner *whisperCommandRunner, ffmpegPath, whisperPath, modelPath, language, audioPath, directory string, cues []subtitleCue, mediaDuration time.Duration) ([]subtitleCue, error) {
	return retryLowConfidenceCuesFromCursor(ctx, whisperRunner, ffmpegPath, whisperPath, modelPath, language, audioPath, directory, cues, mediaDuration, 0, nil)
}

func retryLowConfidenceCuesFromCursor(ctx context.Context, whisperRunner *whisperCommandRunner, ffmpegPath, whisperPath, modelPath, language, audioPath, directory string, cues []subtitleCue, mediaDuration time.Duration, startCursor int, onProgress retryProgressFunc) ([]subtitleCue, error) {
	if startCursor < 0 || startCursor > len(cues) {
		return nil, fmt.Errorf("invalid retry cursor %d for %d cues", startCursor, len(cues))
	}
	result := append([]subtitleCue(nil), cues...)
	lastSavedCursor := startCursor
	for cursor := startCursor; cursor < len(cues); {
		batch, nextCursor, err := collectRetryBatch(cues, cursor, mediaDuration)
		if err != nil {
			return nil, err
		}
		cursor = nextCursor
		if len(batch) == 0 {
			continue
		}
		selectedCues, err := transcribeRetryBatch(ctx, whisperRunner, ffmpegPath, whisperPath, modelPath, language, audioPath, directory, batch)
		if err != nil {
			return nil, err
		}
		updated := append([]subtitleCue(nil), result...)
		for index, item := range batch {
			updated[item.index] = selectedCues[index]
		}
		result = updated
		if onProgress != nil {
			if err := onProgress(nextCursor, result); err != nil {
				return nil, err
			}
			lastSavedCursor = nextCursor
		}
	}
	if onProgress != nil && lastSavedCursor < len(cues) {
		if err := onProgress(len(cues), result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func collectRetryBatch(cues []subtitleCue, startCursor int, mediaDuration time.Duration) ([]retryBatchItem, int, error) {
	batch := make([]retryBatchItem, 0, retryTranscriptionBatchSize)
	cursor := startCursor
	for cursor < len(cues) && len(batch) < retryTranscriptionBatchSize {
		cue := cues[cursor]
		if cue.Probability < retrySubtitleProbability {
			window, err := retryWindowForCue(cue, mediaDuration)
			if err != nil {
				return nil, startCursor, err
			}
			batch = append(batch, retryBatchItem{index: cursor, cue: cue, window: window})
		}
		cursor++
	}
	return batch, cursor, nil
}

func transcribeRetryBatch(ctx context.Context, whisperRunner *whisperCommandRunner, ffmpegPath, whisperPath, modelPath, language, audioPath, directory string, batch []retryBatchItem) ([]subtitleCue, error) {
	retryDirectory, err := os.MkdirTemp(directory, "retry-*")
	if err != nil {
		return nil, fmt.Errorf("create retry chunk directory: %w", err)
	}
	defer os.RemoveAll(retryDirectory)

	windows := make([]audioChunk, len(batch))
	for index, item := range batch {
		windows[index] = item.window
	}
	chunks, err := extractTranscriptionChunks(ctx, ffmpegPath, audioPath, retryDirectory, windows)
	if err != nil {
		return nil, fmt.Errorf("extract retry chunks: %w", err)
	}
	if err := transcribeRetryAudioChunks(ctx, whisperRunner, whisperPath, modelPath, language, chunks); err != nil {
		return nil, err
	}

	result := make([]subtitleCue, len(batch))
	for index, item := range batch {
		retryCues, err := loadTranscriptionCues(chunks[index : index+1])
		if err != nil {
			return nil, fmt.Errorf("load retry cues: %w", err)
		}
		candidate, err := retryCandidateForCue(item.cue, retryCues)
		if err != nil {
			return nil, err
		}
		result[index] = selectRetryCue(item.cue, candidate)
	}
	return result, nil
}

func retryWindowForCue(cue subtitleCue, mediaDuration time.Duration) (audioChunk, error) {
	start := max(time.Duration(0), cue.Start-retryContextDuration)
	end := cue.End + retryContextDuration
	if mediaDuration > 0 {
		end = min(end, mediaDuration)
	}
	if end <= start {
		return audioChunk{}, fmt.Errorf("invalid retry window %s --> %s", start, end)
	}
	return audioChunk{Start: start, End: end}, nil
}

func retryCandidateForCue(original subtitleCue, retryCues []subtitleCue) (subtitleCue, error) {
	tokens := make([]subtitleToken, 0)
	for _, retryCue := range retryCues {
		if len(retryCue.Tokens) == 0 {
			continue
		}
		for _, token := range retryCue.Tokens {
			midpoint := token.Start + (token.End-token.Start)/2
			if midpoint < original.Start || midpoint >= original.End {
				continue
			}
			token.Origin = original.Origin
			tokens = append(tokens, token)
		}
	}
	if len(tokens) == 0 {
		return original, nil
	}
	text := strings.TrimSpace(joinSubtitleTokenText(tokens))
	if text == "" {
		return original, nil
	}
	return subtitleCue{
		Start:       tokens[0].Start,
		End:         tokens[len(tokens)-1].End,
		Text:        text,
		Probability: averageSubtitleTokenProbability(tokens, original.Probability),
		Tokens:      tokens,
		Origin:      original.Origin,
	}, nil
}

func selectRetryCue(original, candidate subtitleCue) subtitleCue {
	if candidate.Probability < minimumSubtitleProbability {
		return original
	}
	if original.Probability < minimumSubtitleProbability || candidate.Probability >= original.Probability+minimumRetryConfidenceGain {
		return candidate
	}
	return original
}
