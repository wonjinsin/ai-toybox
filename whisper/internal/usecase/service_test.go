package usecase

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	portin "github.com/wonjinsin/ai-toybox/whisper/internal/port/in"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

type memorySession struct {
	state           portout.ResumeState
	restored        bool
	recognitions    int
	failRecognition int
	normalized      bool
	detected        bool
	closed          bool
	removed         bool
	partials        [][]domain.Cue
	published       []domain.Cue
	saved           []portout.ResumeState
	chunks          []portout.ChunkAudio
}

func (m *memorySession) Open(context.Context, portout.SessionRequest) (portout.Session, error) {
	return portout.Session{
		Job:   portout.Job{InputPath: "input.mp4", OutputPath: "input.srt", Language: "ko", Format: "srt"},
		Audio: m, Detector: m, Recognizer: m, Checkpoints: m, Transcripts: m,
		Close: func() error { m.closed = true; return nil },
	}, nil
}
func (m *memorySession) Normalize(context.Context) error { m.normalized = true; return nil }
func (m *memorySession) Duration(context.Context) (time.Duration, error) {
	return 1000 * time.Second, nil
}
func (m *memorySession) Detect(context.Context) ([]domain.SpeechSegment, error) {
	m.detected = true
	return []domain.SpeechSegment{{Start: time.Second, End: 2 * time.Second}}, nil
}
func (m *memorySession) Extract(_ context.Context, chunks []domain.Chunk, first int) (portout.AudioBatch, error) {
	result := make([]portout.ChunkAudio, len(chunks))
	for i, chunk := range chunks {
		result[i] = portout.ChunkAudio{Chunk: chunk, Index: first + i}
	}
	m.chunks = append(m.chunks, result...)
	return portout.AudioBatch{Chunks: result}, nil
}
func (m *memorySession) Recognize(_ context.Context, chunks []portout.ChunkAudio, _ portout.RecognitionMode, progress func(int)) ([][]domain.Cue, error) {
	m.recognitions++
	if m.recognitions == m.failRecognition {
		return nil, errors.New("inference failed")
	}
	result := make([][]domain.Cue, len(chunks))
	for i, chunk := range chunks {
		result[i] = []domain.Cue{{Start: chunk.Chunk.Start, End: chunk.Chunk.End, Text: "발화", Probability: .95,
			Origin: domain.Origin{Index: chunk.Index, Start: chunk.Chunk.Start, End: chunk.Chunk.End}}}
	}
	if progress != nil {
		progress(len(chunks))
	}
	return result, nil
}
func (m *memorySession) Load() (portout.ResumeState, bool, error) { return m.state, m.restored, nil }
func (m *memorySession) Save(state portout.ResumeState) error {
	m.saved = append(m.saved, state)
	return nil
}
func (m *memorySession) WritePartial(cues []domain.Cue) error {
	m.partials = append(m.partials, cues)
	return nil
}
func (m *memorySession) Publish(cues []domain.Cue) error { m.published = cues; return nil }
func (m *memorySession) RemoveProgress() error           { m.removed = true; return nil }

func TestTranscribeThroughPortsPublishesValidatedCues(t *testing.T) {
	memory := &memorySession{}
	var service portin.Transcriber = NewService(nil, memory, nil)
	path, err := service.Transcribe(context.Background(), portin.Request{InputPath: "input.mp4", Language: "ko", Format: "srt"})
	if err != nil {
		t.Fatal(err)
	}
	if path != "input.srt" || len(memory.published) != 1 || memory.published[0].Text != "발화" {
		t.Fatalf("published %q: %#v", path, memory.published)
	}
	if !memory.normalized || !memory.detected || !memory.closed || !memory.removed {
		t.Fatalf("incomplete processing lifecycle: %#v", memory)
	}
	if err := domain.ValidateSubtitleCues(memory.published, 1000*time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestResumeSkipsCompletedAudioAndKeepsCheckpointInputImmutable(t *testing.T) {
	original := portout.ResumeState{Stage: portout.StageTranscribing, MediaDuration: 10 * time.Second, CompletedChunks: 1,
		Chunks: []domain.Chunk{{Start: time.Second, End: 2 * time.Second}, {Start: 5 * time.Second, End: 6 * time.Second}},
		Cues:   []domain.Cue{{Start: time.Second, End: 2 * time.Second, Text: "완료", Probability: .95}}}
	memory := &memorySession{state: original, restored: true}
	_, err := NewService(nil, memory, nil).Transcribe(context.Background(), portin.Request{InputPath: "input.mp4", Language: "ko", Format: "srt"})
	if err != nil {
		t.Fatal(err)
	}
	if memory.detected || !memory.normalized || len(memory.chunks) != 1 || memory.chunks[0].Index != 1 {
		t.Fatalf("resume performed incorrect work: %#v", memory)
	}
	if len(memory.published) != 2 || memory.published[0].Text != "완료" {
		t.Fatalf("lost resumed cues: %#v", memory.published)
	}
	if !reflect.DeepEqual(original, memory.state) {
		t.Fatal("resume mutated stored checkpoint")
	}
}

func TestFailedBatchPreservesOnlyDurableCompletedProgress(t *testing.T) {
	chunks := make([]domain.Chunk, 129)
	for i := range chunks {
		chunks[i] = domain.Chunk{Start: time.Duration(i*5+1) * time.Second, End: time.Duration(i*5+2) * time.Second}
	}
	memory := &memorySession{restored: true, failRecognition: 2, state: portout.ResumeState{Stage: portout.StageTranscribing, MediaDuration: 1000 * time.Second, Chunks: chunks}}
	_, err := NewService(nil, memory, nil).Transcribe(context.Background(), portin.Request{InputPath: "input.mp4", Language: "ko", Format: "srt"})
	if err == nil || err.Error() != "inference failed" {
		t.Fatalf("error = %v", err)
	}
	if len(memory.saved) != 1 || memory.saved[0].CompletedChunks != 128 || len(memory.partials[0]) != 128 {
		t.Fatalf("incorrect durable progress: %#v", memory.saved)
	}
	if memory.published != nil || memory.removed || !memory.closed {
		t.Fatal("failure published or removed recovery artifacts")
	}
}
