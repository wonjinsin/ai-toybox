package out

import (
	"context"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
)

type ChunkAudio struct {
	Chunk domain.Chunk
	Index int
	Path  string
}

// AudioBatch groups extracted audio with its optional resource cleanup.
type AudioBatch struct {
	Chunks []ChunkAudio
	// Release frees batch resources; callers must invoke it after consuming decoded results.
	// It may be nil when the batch owns no resources.
	Release func() error
}

type AudioProcessor interface {
	Normalize(context.Context) error
	Duration(context.Context) (time.Duration, error)
	Extract(context.Context, []domain.Chunk, int) (AudioBatch, error)
}

type SpeechDetector interface {
	Detect(context.Context) ([]domain.SpeechSegment, error)
}

type RecognitionMode int

const (
	RecognitionStandard RecognitionMode = iota
	RecognitionRetry
)

type Recognizer interface {
	Recognize(context.Context, []ChunkAudio, RecognitionMode, func(int)) ([][]domain.Cue, error)
}
