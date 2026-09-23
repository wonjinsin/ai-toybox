package usecase

import (
	"context"
	"fmt"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

func (service *Service) retry(ctx context.Context, session portout.Session, state portout.ResumeState) (portout.ResumeState, error) {
	original := state.Cues
	lastSavedCursor := state.RetryCursor
	for cursor := state.RetryCursor; cursor < len(original); {
		if err := ctx.Err(); err != nil {
			return portout.ResumeState{}, err
		}
		indices := make([]int, 0, portout.TranscriptionBatchSize)
		windows := make([]domain.Chunk, 0, portout.TranscriptionBatchSize)
		for cursor < len(original) && len(indices) < portout.TranscriptionBatchSize {
			cue := original[cursor]
			if cue.Probability < domain.RetrySubtitleProbability {
				window, err := domain.RetryWindowForCue(cue, state.MediaDuration)
				if err != nil {
					return portout.ResumeState{}, err
				}
				indices = append(indices, cursor)
				windows = append(windows, window)
			}
			cursor++
		}
		if len(indices) == 0 {
			continue
		}
		results, err := service.recognizeRetryBatch(ctx, session, windows)
		if err != nil {
			return portout.ResumeState{}, err
		}
		if len(results) != len(indices) {
			return portout.ResumeState{}, fmt.Errorf("recognition returned %d retry results, want %d", len(results), len(indices))
		}
		updated := append([]domain.Cue(nil), state.Cues...)
		for index, cueIndex := range indices {
			candidate, err := domain.RetryCandidateForCue(original[cueIndex], results[index])
			if err != nil {
				return portout.ResumeState{}, err
			}
			updated[cueIndex] = domain.SelectRetryCue(original[cueIndex], candidate)
		}
		state = portout.ResumeState{Stage: portout.StageRetrying, MediaDuration: state.MediaDuration, Chunks: state.Chunks,
			CompletedChunks: state.CompletedChunks, RetryCursor: cursor, Cues: updated}
		if session.Job.Format == "srt" {
			if err := service.saveRetryProgress(session, state); err != nil {
				return portout.ResumeState{}, err
			}
			lastSavedCursor = cursor
		}
	}
	if session.Job.Format == "srt" && lastSavedCursor < len(original) {
		state.RetryCursor = len(original)
		if err := service.saveRetryProgress(session, state); err != nil {
			return portout.ResumeState{}, err
		}
	}
	return state, nil
}

func (service *Service) recognizeRetryBatch(ctx context.Context, session portout.Session, windows []domain.Chunk) ([][]domain.Cue, error) {
	batch, err := session.Audio.Extract(ctx, windows, 0)
	if err != nil {
		return nil, fmt.Errorf("extract retry chunks: %w", err)
	}
	defer service.releaseAudio(session.Job.InputPath, batch)
	return session.Recognizer.Recognize(ctx, batch.Chunks, portout.RecognitionRetry, nil)
}

func (service *Service) saveRetryProgress(session portout.Session, state portout.ResumeState) error {
	if err := persistProgress(session, state); err != nil {
		return err
	}
	service.report(portin.ProgressEvent{InputPath: session.Job.InputPath, Kind: portin.EventSaved, Stage: portin.StepRetry,
		Completed: state.RetryCursor, Total: len(state.Cues)})
	return nil
}
