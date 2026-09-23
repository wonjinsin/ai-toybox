package out

import (
	"errors"
	"fmt"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
)

const (
	TranscriptionBatchSize = 128
	StageTranscribing      = "transcribing"
	StageRetrying          = "retrying"
	StageRetryComplete     = "retry_complete"
)

type ResumeState struct {
	Stage           string
	MediaDuration   time.Duration
	CompletedChunks int
	RetryCursor     int
	Chunks          []domain.Chunk
	Cues            []domain.Cue
}

type CheckpointStore interface {
	Load() (ResumeState, bool, error)
	Save(ResumeState) error
}

type TranscriptStore interface {
	WritePartial([]domain.Cue) error
	Publish([]domain.Cue) error
	RemoveProgress() error
}

func (state ResumeState) Validate() error {
	if state.Stage != StageTranscribing && state.Stage != StageRetrying && state.Stage != StageRetryComplete {
		return fmt.Errorf("unsupported stage %q", state.Stage)
	}
	if state.MediaDuration <= 0 || state.CompletedChunks < 0 || state.CompletedChunks > len(state.Chunks) {
		return errors.New("invalid progress")
	}
	if state.RetryCursor < 0 || state.RetryCursor > len(state.Cues) {
		return errors.New("invalid retry cursor")
	}
	if state.Stage == StageTranscribing && state.RetryCursor != 0 {
		return errors.New("retry cursor set during main transcription")
	}
	if state.Stage != StageTranscribing && state.CompletedChunks != len(state.Chunks) {
		return errors.New("retry stage before main transcription completed")
	}
	if state.Stage == StageRetryComplete && state.RetryCursor != len(state.Cues) {
		return errors.New("incomplete retry cursor")
	}
	for index, chunk := range state.Chunks {
		if chunk.Start < 0 || chunk.End <= chunk.Start || chunk.End > state.MediaDuration {
			return fmt.Errorf("invalid chunk %d", index)
		}
	}
	return nil
}
