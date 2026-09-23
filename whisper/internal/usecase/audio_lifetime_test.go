package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

type boundedAudio struct {
	portout.AudioProcessor
	active       int
	released     int
	releaseError error
}

func (audio *boundedAudio) Extract(ctx context.Context, chunks []domain.Chunk, first int) (portout.AudioBatch, error) {
	if len(chunks) == 0 {
		return portout.AudioBatch{}, nil
	}
	if audio.active != 0 {
		return portout.AudioBatch{}, errors.New("previous audio batch still occupies temporary storage")
	}
	batch, err := audio.AudioProcessor.Extract(ctx, chunks, first)
	if err != nil {
		return portout.AudioBatch{}, err
	}
	audio.active++
	return portout.AudioBatch{Chunks: batch.Chunks, Release: func() error {
		audio.active--
		audio.released++
		return audio.releaseError
	}}, nil
}

func TestRetryReleasesEachAudioBatchBeforeExtractingTheNext(t *testing.T) {
	memory := &memorySession{state: retryContractState(129), restored: true}
	audio := &boundedAudio{AudioProcessor: memory}
	var service portin.Transcriber = contractService(memory, func(session portout.Session) portout.Session {
		session.Audio = audio
		session.Recognizer = contractRecognizer(func(_ context.Context, chunks []portout.ChunkAudio, _ portout.RecognitionMode, _ func(int)) ([][]domain.Cue, error) {
			return improvedRetryCues(chunks), nil
		})
		return session
	}, nil)
	if _, err := service.Transcribe(context.Background(), contractRequest()); err != nil {
		t.Fatal(err)
	}
	if audio.active != 0 || audio.released != 2 || len(memory.published) != 129 {
		t.Fatalf("active = %d, released = %d, published = %d", audio.active, audio.released, len(memory.published))
	}
}

func TestRecognitionFailureReleasesItsAudioBatch(t *testing.T) {
	for _, retry := range []bool{false, true} {
		name := "main"
		if retry {
			name = "retry"
		}
		t.Run(name, func(t *testing.T) {
			memory := &memorySession{failRecognition: 1}
			if retry {
				memory.state, memory.restored = retryContractState(1), true
			}
			audio := &boundedAudio{AudioProcessor: memory}
			service := contractService(memory, func(session portout.Session) portout.Session {
				session.Audio = audio
				return session
			}, nil)
			_, err := service.Transcribe(context.Background(), contractRequest())
			if err == nil || err.Error() != "inference failed" {
				t.Fatalf("error = %v", err)
			}
			if audio.active != 0 || audio.released != 1 || memory.published != nil {
				t.Fatalf("active = %d, released = %d, published = %#v", audio.active, audio.released, memory.published)
			}
		})
	}
}

func TestAudioReleaseFailureReportsWarningWithoutFailingTranscription(t *testing.T) {
	memory := &memorySession{}
	releaseError := errors.New("temporary batch cleanup failed")
	audio := &boundedAudio{AudioProcessor: memory, releaseError: releaseError}
	var warnings []error
	service := contractService(memory, func(session portout.Session) portout.Session {
		session.Audio = audio
		return session
	}, func(event portin.ProgressEvent) {
		if event.Kind == portin.EventWarning {
			warnings = append(warnings, event.Err)
		}
	})
	path, err := service.Transcribe(context.Background(), contractRequest())
	if err != nil || path != "input.srt" || len(memory.published) != 1 {
		t.Fatalf("path = %q, error = %v, published = %#v", path, err, memory.published)
	}
	if len(warnings) != 1 || !errors.Is(warnings[0], releaseError) {
		t.Fatalf("warnings = %v", warnings)
	}
}
