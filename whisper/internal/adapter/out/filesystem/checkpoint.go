package filesystem

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
	portout "github.com/wonjinsin/ai-toybox/whisper/internal/port/out"
)

const (
	incrementalCheckpointVersion = 1
	incrementalPipelineVersion   = 4
	transcriptionBatchSize       = portout.TranscriptionBatchSize
)

type incrementalCheckpoint struct {
	Version         int                    `json:"version"`
	Fingerprint     incrementalFingerprint `json:"fingerprint"`
	Stage           string                 `json:"stage"`
	BatchSize       int                    `json:"batch_size"`
	MediaDuration   time.Duration          `json:"media_duration"`
	CompletedChunks int                    `json:"completed_chunks"`
	RetryCursor     int                    `json:"retry_cursor"`
	Chunks          []checkpointAudioChunk `json:"chunks"`
	Cues            []checkpointCue        `json:"cues"`
}

type incrementalFingerprint struct {
	PipelineVersion int                    `json:"pipeline_version"`
	Language        string                 `json:"language"`
	Input           checkpointFileIdentity `json:"input"`
	Model           checkpointFileIdentity `json:"model"`
	VADModel        checkpointFileIdentity `json:"vad_model"`
	Whisper         checkpointFileIdentity `json:"whisper"`
	VADTool         checkpointFileIdentity `json:"vad_tool"`
	FFmpeg          checkpointFileIdentity `json:"ffmpeg"`
	FFprobe         checkpointFileIdentity `json:"ffprobe"`
}

type checkpointFileIdentity struct {
	Path             string `json:"path"`
	Size             int64  `json:"size"`
	ModifiedUnixNano int64  `json:"modified_unix_nano"`
}

type checkpointAudioChunk struct {
	Start time.Duration `json:"start"`
	End   time.Duration `json:"end"`
}

type checkpointCue struct {
	Start       time.Duration     `json:"start"`
	End         time.Duration     `json:"end"`
	Text        string            `json:"text"`
	Probability float64           `json:"probability"`
	Tokens      []checkpointToken `json:"tokens"`
	Origin      checkpointOrigin  `json:"origin"`
}

type checkpointToken struct {
	Start       time.Duration    `json:"start"`
	End         time.Duration    `json:"end"`
	Text        string           `json:"text"`
	Probability float64          `json:"probability"`
	Origin      checkpointOrigin `json:"origin"`
}

type checkpointOrigin struct {
	Index int           `json:"index"`
	Start time.Duration `json:"start"`
	End   time.Duration `json:"end"`
}

func incrementalArtifactPaths(outputBase string) (string, string) {
	directory := filepath.Dir(outputBase)
	baseName := filepath.Base(outputBase)
	partialPath := outputBase + ".partial.srt"
	checkpointPath := filepath.Join(directory, "."+baseName+".srt.whisper-local-checkpoint.json")
	return partialPath, checkpointPath
}

func newIncrementalCheckpointWithState(fingerprint incrementalFingerprint, stage string, retryCursor int, mediaDuration time.Duration, chunks []domain.Chunk, completedChunks int, cues []domain.Cue) incrementalCheckpoint {
	checkpointChunks := make([]checkpointAudioChunk, len(chunks))
	for index, chunk := range chunks {
		checkpointChunks[index] = checkpointAudioChunk{Start: chunk.Start, End: chunk.End}
	}
	return incrementalCheckpoint{
		Version:         incrementalCheckpointVersion,
		Fingerprint:     fingerprint,
		Stage:           stage,
		BatchSize:       transcriptionBatchSize,
		MediaDuration:   mediaDuration,
		CompletedChunks: completedChunks,
		RetryCursor:     retryCursor,
		Chunks:          checkpointChunks,
		Cues:            checkpointCuesFromSubtitleCues(cues),
	}
}

func newIncrementalFingerprint(inputPath, modelPath, vadModelPath, whisperPath, vadToolPath, ffmpegPath, ffprobePath, language string) (incrementalFingerprint, error) {
	paths := []struct {
		label string
		path  string
	}{
		{label: "input", path: inputPath},
		{label: "model", path: modelPath},
		{label: "VAD model", path: vadModelPath},
		{label: "whisper-cli", path: whisperPath},
		{label: "whisper-vad-speech-segments", path: vadToolPath},
		{label: "ffmpeg", path: ffmpegPath},
		{label: "ffprobe", path: ffprobePath},
	}
	identities := make([]checkpointFileIdentity, len(paths))
	for index, candidate := range paths {
		identity, err := checkpointIdentityForFile(candidate.path)
		if err != nil {
			return incrementalFingerprint{}, fmt.Errorf("identify %s for checkpoint: %w", candidate.label, err)
		}
		identities[index] = identity
	}
	return incrementalFingerprint{
		PipelineVersion: incrementalPipelineVersion,
		Language:        language,
		Input:           identities[0],
		Model:           identities[1],
		VADModel:        identities[2],
		Whisper:         identities[3],
		VADTool:         identities[4],
		FFmpeg:          identities[5],
		FFprobe:         identities[6],
	}, nil
}

func checkpointIdentityForFile(path string) (checkpointFileIdentity, error) {
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return checkpointFileIdentity{}, err
	}
	absolutePath, err := filepath.Abs(resolvedPath)
	if err != nil {
		return checkpointFileIdentity{}, err
	}
	info, err := os.Stat(absolutePath)
	if err != nil {
		return checkpointFileIdentity{}, err
	}
	if !info.Mode().IsRegular() {
		return checkpointFileIdentity{}, fmt.Errorf("%q is not a regular file", absolutePath)
	}
	return checkpointFileIdentity{Path: absolutePath, Size: info.Size(), ModifiedUnixNano: info.ModTime().UnixNano()}, nil
}

func checkpointCuesFromSubtitleCues(cues []domain.Cue) []checkpointCue {
	result := make([]checkpointCue, len(cues))
	for cueIndex, cue := range cues {
		tokens := make([]checkpointToken, len(cue.Tokens))
		for tokenIndex, token := range cue.Tokens {
			tokens[tokenIndex] = checkpointToken{
				Start: token.Start, End: token.End, Text: token.Text, Probability: token.Probability,
				Origin: checkpointOriginFromTranscriptionOrigin(token.Origin),
			}
		}
		result[cueIndex] = checkpointCue{
			Start: cue.Start, End: cue.End, Text: cue.Text, Probability: cue.Probability,
			Tokens: tokens, Origin: checkpointOriginFromTranscriptionOrigin(cue.Origin),
		}
	}
	return result
}

func checkpointOriginFromTranscriptionOrigin(origin domain.Origin) checkpointOrigin {
	return checkpointOrigin{Index: origin.Index, Start: origin.Start, End: origin.End}
}

func subtitleCuesFromCheckpointCues(cues []checkpointCue) []domain.Cue {
	result := make([]domain.Cue, len(cues))
	for cueIndex, cue := range cues {
		tokens := make([]domain.Token, len(cue.Tokens))
		for tokenIndex, token := range cue.Tokens {
			tokens[tokenIndex] = domain.Token{
				Start: token.Start, End: token.End, Text: token.Text, Probability: token.Probability,
				Origin: transcriptionOriginFromCheckpointOrigin(token.Origin),
			}
		}
		result[cueIndex] = domain.Cue{
			Start: cue.Start, End: cue.End, Text: cue.Text, Probability: cue.Probability,
			Tokens: tokens, Origin: transcriptionOriginFromCheckpointOrigin(cue.Origin),
		}
	}
	return result
}

func transcriptionOriginFromCheckpointOrigin(origin checkpointOrigin) domain.Origin {
	return domain.Origin{Index: origin.Index, Start: origin.Start, End: origin.End}
}

func audioChunksFromCheckpoint(chunks []checkpointAudioChunk) []domain.Chunk {
	result := make([]domain.Chunk, len(chunks))
	for index, chunk := range chunks {
		result[index] = domain.Chunk{Start: chunk.Start, End: chunk.End}
	}
	return result
}

func loadIncrementalCheckpoint(path string, fingerprint incrementalFingerprint) (incrementalCheckpoint, bool, error) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return incrementalCheckpoint{}, false, nil
	}
	if err != nil {
		return incrementalCheckpoint{}, false, fmt.Errorf("read transcription checkpoint %q: %w", path, err)
	}
	var checkpoint incrementalCheckpoint
	if err := json.Unmarshal(content, &checkpoint); err != nil {
		return incrementalCheckpoint{}, false, fmt.Errorf("decode transcription checkpoint %q: %w", path, err)
	}
	if checkpoint.Version != incrementalCheckpointVersion || checkpoint.BatchSize != transcriptionBatchSize || checkpoint.Fingerprint != fingerprint {
		return incrementalCheckpoint{}, false, nil
	}
	if err := checkpoint.resumeState().Validate(); err != nil {
		return incrementalCheckpoint{}, false, fmt.Errorf("validate transcription checkpoint %q: %w", path, err)
	}
	return checkpoint, true, nil
}

func persistIncrementalCheckpoint(path string, checkpoint incrementalCheckpoint) error {
	content, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("encode transcription checkpoint: %w", err)
	}
	if err := replaceFileAtomically(path, content, 0o600); err != nil {
		return fmt.Errorf("write transcription checkpoint %q: %w", path, err)
	}
	return nil
}

func removeIncrementalArtifacts(outputBase string) error {
	partialPath, checkpointPath := incrementalArtifactPaths(outputBase)
	var removalErrors []error
	for _, path := range []string{checkpointPath, partialPath} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			removalErrors = append(removalErrors, fmt.Errorf("remove incremental artifact %q: %w", path, err))
		}
	}
	return errors.Join(removalErrors...)
}

func replaceFileAtomically(path string, content []byte, mode os.FileMode) (returnErr error) {
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		if returnErr != nil {
			_ = os.Remove(temporaryPath)
		}
	}()

	_, writeErr := temporary.Write(content)
	var chmodErr error
	if writeErr == nil {
		chmodErr = temporary.Chmod(mode)
	}
	var syncErr error
	if writeErr == nil && chmodErr == nil {
		syncErr = temporary.Sync()
	}
	closeErr := temporary.Close()
	if operationErr := errors.Join(writeErr, chmodErr, syncErr, closeErr); operationErr != nil {
		return fmt.Errorf("write temporary file %q: %w", temporaryPath, operationErr)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace file %q: %w", path, err)
	}
	return nil
}
