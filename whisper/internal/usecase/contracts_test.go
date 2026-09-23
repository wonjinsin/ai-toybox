package usecase

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

type contractFactory func(context.Context, portout.SessionRequest) (portout.Session, error)

func (factory contractFactory) Open(ctx context.Context, request portout.SessionRequest) (portout.Session, error) {
	return factory(ctx, request)
}

type contractCatalog func(string) (portout.InputSet, error)

func (catalog contractCatalog) Discover(path string) (portout.InputSet, error) { return catalog(path) }

type contractRecognizer func(context.Context, []portout.ChunkAudio, portout.RecognitionMode, func(int)) ([][]domain.Cue, error)

func (recognizer contractRecognizer) Recognize(ctx context.Context, chunks []portout.ChunkAudio, mode portout.RecognitionMode, progress func(int)) ([][]domain.Cue, error) {
	return recognizer(ctx, chunks, mode, progress)
}

type contractTranscript struct {
	portout.TranscriptStore
	partialError error
	publishError error
}

func (store contractTranscript) WritePartial(cues []domain.Cue) error {
	if store.partialError != nil {
		return store.partialError
	}
	return store.TranscriptStore.WritePartial(cues)
}
func (store contractTranscript) Publish(cues []domain.Cue) error {
	if store.publishError != nil {
		return store.publishError
	}
	return store.TranscriptStore.Publish(cues)
}

func contractService(memory *memorySession, configure func(portout.Session) portout.Session, observer portin.ProgressObserver) *Service {
	return NewService(nil, contractFactory(func(ctx context.Context, request portout.SessionRequest) (portout.Session, error) {
		session, err := memory.Open(ctx, request)
		if err != nil {
			return portout.Session{}, err
		}
		return configure(session), nil
	}), observer)
}

func contractRequest() portin.Request {
	return portin.Request{InputPath: "input.mp4", Language: "ko", Format: "srt", Parallel: 1}
}

func retryContractState(count int) portout.ResumeState {
	chunks := make([]domain.Chunk, count)
	cues := make([]domain.Cue, count)
	for index := range cues {
		start := time.Duration(index*3+2) * time.Second
		chunks[index] = domain.Chunk{Start: start, End: start + time.Second}
		cues[index] = domain.Cue{Start: start, End: start + time.Second, Text: "원본", Probability: .4,
			Origin: domain.Origin{Index: index, Start: start, End: start + time.Second}}
	}
	return portout.ResumeState{Stage: portout.StageRetrying, MediaDuration: 1000 * time.Second, CompletedChunks: count, Chunks: chunks, Cues: cues}
}

func improvedRetryCues(chunks []portout.ChunkAudio) [][]domain.Cue {
	results := make([][]domain.Cue, len(chunks))
	for index, chunk := range chunks {
		start, end := chunk.Chunk.Start+1100*time.Millisecond, chunk.Chunk.Start+1900*time.Millisecond
		results[index] = []domain.Cue{{Start: start, End: end, Text: "개선", Probability: .9,
			Tokens: []domain.Token{{Start: start, End: end, Text: " 개선", Probability: .9}}}}
	}
	return results
}

func TestRetryThroughPortsBatches129CuesAndPublishesImprovedTokens(t *testing.T) {
	memory := &memorySession{state: retryContractState(129), restored: true}
	var batchSizes []int
	service := contractService(memory, func(session portout.Session) portout.Session {
		session.Recognizer = contractRecognizer(func(_ context.Context, chunks []portout.ChunkAudio, mode portout.RecognitionMode, _ func(int)) ([][]domain.Cue, error) {
			if mode != portout.RecognitionRetry {
				t.Fatalf("resumed retry requested mode %v", mode)
			}
			batchSizes = append(batchSizes, len(chunks))
			return improvedRetryCues(chunks), nil
		})
		return session
	}, nil)
	if _, err := service.Transcribe(context.Background(), contractRequest()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(batchSizes, []int{128, 1}) {
		t.Fatalf("retry batches = %v", batchSizes)
	}
	if len(memory.published) != 129 {
		t.Fatalf("published cues = %d", len(memory.published))
	}
	for index, cue := range memory.published {
		if cue.Text != "개선" || cue.Probability != .9 || cue.Origin != memory.state.Cues[index].Origin {
			t.Fatalf("published cue %d = %#v", index, cue)
		}
	}
	last := memory.saved[len(memory.saved)-1]
	if last.Stage != portout.StageRetryComplete || last.RetryCursor != 129 || !memory.removed {
		t.Fatalf("final durable state = %#v", last)
	}
	if memory.state.Cues[0].Text != "원본" {
		t.Fatal("retry modified the restored checkpoint in place")
	}
}

func TestFailedRetryBatchKeepsDurableCursorAndResumeSkipsProcessedLowConfidenceCues(t *testing.T) {
	initial := retryContractState(129)
	memory := &memorySession{state: initial, restored: true}
	wantError := errors.New("retry interrupted")
	batches := 0
	service := contractService(memory, func(session portout.Session) portout.Session {
		session.Recognizer = contractRecognizer(func(_ context.Context, chunks []portout.ChunkAudio, _ portout.RecognitionMode, _ func(int)) ([][]domain.Cue, error) {
			batches++
			if batches == 2 {
				return nil, wantError
			}
			return make([][]domain.Cue, len(chunks)), nil
		})
		return session
	}, nil)
	if _, err := service.Transcribe(context.Background(), contractRequest()); !errors.Is(err, wantError) {
		t.Fatalf("error = %v", err)
	}
	saved := memory.saved[len(memory.saved)-1]
	if saved.Stage != portout.StageRetrying || saved.RetryCursor != 128 || saved.Cues[0].Probability != .4 {
		t.Fatalf("durable state = %#v", saved)
	}
	if memory.removed || memory.published != nil || !memory.closed {
		t.Fatal("failed retry discarded recovery state or leaked the session")
	}

	resumed := &memorySession{state: saved, restored: true}
	var windows []domain.Chunk
	resumeService := contractService(resumed, func(session portout.Session) portout.Session {
		session.Recognizer = contractRecognizer(func(_ context.Context, chunks []portout.ChunkAudio, mode portout.RecognitionMode, _ func(int)) ([][]domain.Cue, error) {
			if mode != portout.RecognitionRetry {
				t.Fatalf("mode = %v", mode)
			}
			for _, chunk := range chunks {
				windows = append(windows, chunk.Chunk)
			}
			return improvedRetryCues(chunks), nil
		})
		return session
	}, nil)
	if _, err := resumeService.Transcribe(context.Background(), contractRequest()); err != nil {
		t.Fatal(err)
	}
	if len(windows) != 1 || windows[0].Start != initial.Cues[128].Start-time.Second {
		t.Fatalf("resumed windows = %#v", windows)
	}
	if len(resumed.published) != 129 || resumed.published[0].Text != "원본" || resumed.published[128].Text != "개선" {
		t.Fatalf("resume lost or replaced previously completed cues: %#v", resumed.published)
	}
}

func TestPartialWriteFailureDoesNotAdvanceDurableCheckpoint(t *testing.T) {
	state := portout.ResumeState{Stage: portout.StageTranscribing, MediaDuration: 10 * time.Second, Chunks: []domain.Chunk{{Start: time.Second, End: 2 * time.Second}}}
	memory := &memorySession{state: state, restored: true}
	wantError := errors.New("partial transcript disk full")
	service := contractService(memory, func(session portout.Session) portout.Session {
		session.Transcripts = contractTranscript{TranscriptStore: memory, partialError: wantError}
		return session
	}, nil)
	_, err := service.Transcribe(context.Background(), contractRequest())
	if !errors.Is(err, wantError) {
		t.Fatalf("error = %v", err)
	}
	if len(memory.saved) != 0 || !reflect.DeepEqual(memory.state, state) {
		t.Fatalf("checkpoint advanced despite partial write failure: %#v", memory.saved)
	}
	if memory.published != nil || memory.removed || !memory.closed {
		t.Fatal("failed partial write published output, removed progress, or leaked resources")
	}
}

func TestPublishFailureRetainsCompletedCheckpointAndPartialTranscript(t *testing.T) {
	memory := &memorySession{}
	wantError := errors.New("output became unavailable")
	service := contractService(memory, func(session portout.Session) portout.Session {
		session.Transcripts = contractTranscript{TranscriptStore: memory, publishError: wantError}
		return session
	}, nil)
	_, err := service.Transcribe(context.Background(), contractRequest())
	if !errors.Is(err, wantError) {
		t.Fatalf("error = %v", err)
	}
	if len(memory.saved) == 0 || len(memory.partials) == 0 {
		t.Fatal("failed publication lost recoverable results")
	}
	last := memory.saved[len(memory.saved)-1]
	if last.Stage != portout.StageRetryComplete || last.CompletedChunks != 1 || memory.removed || !memory.closed {
		t.Fatalf("publication failure left wrong recovery state: %#v", last)
	}
}

func TestCorruptCheckpointIsRejectedBeforeAudioNormalization(t *testing.T) {
	memory := &memorySession{restored: true, state: portout.ResumeState{Stage: portout.StageTranscribing, MediaDuration: time.Second, CompletedChunks: 2}}
	_, err := NewService(nil, memory, nil).Transcribe(context.Background(), contractRequest())
	if err == nil || !strings.Contains(err.Error(), "validate transcription checkpoint") {
		t.Fatalf("error = %v", err)
	}
	if memory.normalized || memory.detected || len(memory.saved) != 0 || !memory.closed {
		t.Fatal("corrupt checkpoint started work, changed durable state, or leaked resources")
	}
}

func TestRecognitionResultSlotMismatchDoesNotAdvanceCheckpoint(t *testing.T) {
	for _, mode := range []portout.RecognitionMode{portout.RecognitionStandard, portout.RecognitionRetry} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			state := retryContractState(1)
			if mode == portout.RecognitionStandard {
				state = portout.ResumeState{Stage: portout.StageTranscribing, MediaDuration: state.MediaDuration, Chunks: state.Chunks}
			}
			memory := &memorySession{state: state, restored: true}
			service := contractService(memory, func(session portout.Session) portout.Session {
				session.Recognizer = contractRecognizer(func(context.Context, []portout.ChunkAudio, portout.RecognitionMode, func(int)) ([][]domain.Cue, error) {
					return nil, nil
				})
				return session
			}, nil)
			_, err := service.Transcribe(context.Background(), contractRequest())
			if err == nil || !strings.Contains(err.Error(), "recognition returned 0") {
				t.Fatalf("error = %v", err)
			}
			for _, saved := range memory.saved {
				if saved.CompletedChunks != state.CompletedChunks || saved.RetryCursor != state.RetryCursor {
					t.Fatalf("malformed result advanced checkpoint: %#v", saved)
				}
			}
			if memory.published != nil || memory.removed || !memory.closed {
				t.Fatal("malformed result was published or recovery lifecycle was broken")
			}
		})
	}
}

func TestCancellationDuringRecognitionPreservesProgressAndClosesSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	memory := &memorySession{}
	var events []portin.ProgressEvent
	service := contractService(memory, func(session portout.Session) portout.Session {
		session.Recognizer = contractRecognizer(func(ctx context.Context, _ []portout.ChunkAudio, _ portout.RecognitionMode, _ func(int)) ([][]domain.Cue, error) {
			cancel()
			<-ctx.Done()
			return nil, ctx.Err()
		})
		return session
	}, func(event portin.ProgressEvent) { events = append(events, event) })
	_, err := service.Transcribe(ctx, contractRequest())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
	if len(memory.saved) != 1 || memory.saved[0].CompletedChunks != 0 || memory.removed || !memory.closed {
		t.Fatal("cancellation lost progress or leaked resources")
	}
	last := events[len(events)-1]
	if last.Kind != portin.EventFailed || !errors.Is(last.Err, context.Canceled) {
		t.Fatalf("final event = %#v", last)
	}
}

func TestCanceledRequestDoesNotOpenSession(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	factory := contractFactory(func(context.Context, portout.SessionRequest) (portout.Session, error) {
		t.Fatal("canceled request opened a session")
		return portout.Session{}, nil
	})
	_, err := NewService(nil, factory, nil).Transcribe(ctx, contractRequest())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
}

func TestServiceRunPreservesDiscoveryDirectoryAndSortedPerFileResults(t *testing.T) {
	request := contractRequest()
	request.InputPath = "recordings"
	request.Parallel = 2
	wantError := errors.New("unreadable recording")
	catalog := contractCatalog(func(path string) (portout.InputSet, error) {
		if path != request.InputPath {
			t.Fatalf("catalog path = %q", path)
		}
		return portout.InputSet{Paths: []string{"z.mp4", "a.mp4"}, Directory: true}, nil
	})
	factory := contractFactory(func(ctx context.Context, request portout.SessionRequest) (portout.Session, error) {
		if request.InputPath == "z.mp4" {
			return portout.Session{}, wantError
		}
		memory := &memorySession{}
		session, err := memory.Open(ctx, request)
		session.Job = portout.Job{InputPath: request.InputPath, OutputPath: "a.srt", Language: request.Language, Format: request.Format}
		return session, err
	})
	got, err := NewService(catalog, factory, nil).Run(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Directory || len(got.Results) != 2 || got.Results[0].InputPath != "a.mp4" || got.Results[0].OutputPath != "a.srt" || got.Results[0].Err != nil || !errors.Is(got.Results[1].Err, wantError) {
		t.Fatalf("batch result = %#v", got)
	}
}

func TestServiceRunValidatesBeforeDiscoveryAndPreservesCatalogErrors(t *testing.T) {
	for _, request := range []portin.Request{
		{Language: "unsupported", Format: "srt", Parallel: 1},
		{Language: "ko", Format: "unsupported", Parallel: 1},
		{Language: "ko", Format: "srt", Parallel: 0},
	} {
		catalog := contractCatalog(func(string) (portout.InputSet, error) {
			t.Fatal("invalid request accessed catalog")
			return portout.InputSet{}, nil
		})
		if _, err := NewService(catalog, nil, nil).Run(context.Background(), request); err == nil {
			t.Fatalf("invalid request accepted: %#v", request)
		}
	}
	wantError := errors.New("catalog unavailable")
	catalog := contractCatalog(func(string) (portout.InputSet, error) { return portout.InputSet{}, wantError })
	if _, err := NewService(catalog, nil, nil).Run(context.Background(), contractRequest()); !errors.Is(err, wantError) {
		t.Fatalf("catalog error = %v", err)
	}
}
