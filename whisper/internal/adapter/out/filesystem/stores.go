package filesystem

import (
	"fmt"
	"os"

	"github.com/wonjinsin/ai-toybox/whisper/internal/adapter/out/subtitle"
	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

// FingerprintConfig holds engine and input identities for checkpoint compatibility.
type FingerprintConfig struct {
	InputPath    string
	ModelPath    string
	VADModelPath string
	WhisperPath  string
	VADToolPath  string
	FFmpegPath   string
	FFprobePath  string
	Language     string
}

type CheckpointStore struct {
	path        string
	fingerprint incrementalFingerprint
}

func NewCheckpointStore(outputBase string, config FingerprintConfig) (*CheckpointStore, error) {
	fingerprint, err := newIncrementalFingerprint(config.InputPath, config.ModelPath, config.VADModelPath,
		config.WhisperPath, config.VADToolPath, config.FFmpegPath, config.FFprobePath, config.Language)
	if err != nil {
		return nil, err
	}
	_, path := incrementalArtifactPaths(outputBase)
	return &CheckpointStore{path: path, fingerprint: fingerprint}, nil
}

func (store *CheckpointStore) Load() (portout.ResumeState, bool, error) {
	checkpoint, found, err := loadIncrementalCheckpoint(store.path, store.fingerprint)
	if err != nil || !found {
		return portout.ResumeState{}, found, err
	}
	return checkpoint.resumeState(), true, nil
}

func (store *CheckpointStore) Save(state portout.ResumeState) error {
	checkpoint := newIncrementalCheckpointWithState(store.fingerprint, state.Stage, state.RetryCursor,
		state.MediaDuration, state.Chunks, state.CompletedChunks, state.Cues)
	return persistIncrementalCheckpoint(store.path, checkpoint)
}

func (checkpoint incrementalCheckpoint) resumeState() portout.ResumeState {
	return portout.ResumeState{
		Stage: checkpoint.Stage, MediaDuration: checkpoint.MediaDuration,
		CompletedChunks: checkpoint.CompletedChunks, RetryCursor: checkpoint.RetryCursor,
		Chunks: audioChunksFromCheckpoint(checkpoint.Chunks), Cues: subtitleCuesFromCheckpointCues(checkpoint.Cues),
	}
}

type TranscriptStore struct {
	outputBase    string
	outputPath    string
	temporaryPath string
	language      string
	format        string
	force         bool
}

func NewTranscriptStore(prepared Prepared, resources *Resources) *TranscriptStore {
	return &TranscriptStore{
		outputBase: prepared.OutputBase, outputPath: prepared.Job.OutputPath,
		temporaryPath: resources.TemporaryOutputPath, language: prepared.Job.Language,
		format: prepared.Job.Format, force: prepared.Force,
	}
}

func (store *TranscriptStore) WritePartial(cues []domain.Cue) error {
	partial, err := subtitle.Render(cues, "srt", store.language)
	if err != nil {
		return fmt.Errorf("render partial transcript: %w", err)
	}
	partialPath, _ := incrementalArtifactPaths(store.outputBase)
	if err := replaceFileAtomically(partialPath, []byte(partial), 0o644); err != nil {
		return fmt.Errorf("write partial transcript %q: %w", partialPath, err)
	}
	return nil
}

func (store *TranscriptStore) Publish(cues []domain.Cue) error {
	transcript, err := subtitle.Render(cues, store.format, store.language)
	if err != nil {
		return err
	}
	return writeAndPublishTranscript(store.temporaryPath, store.outputPath, []byte(transcript), store.force, os.Link)
}

func (store *TranscriptStore) RemoveProgress() error {
	return removeIncrementalArtifacts(store.outputBase)
}
