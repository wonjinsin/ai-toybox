package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	incrementalCheckpointVersion = 1
	incrementalPipelineVersion   = 1
	transcriptionBatchSize       = 32
	checkpointStageTranscribing  = "transcribing"
	checkpointStageRetrying      = "retrying"
	checkpointStageRetryComplete = "retry_complete"
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

func newIncrementalCheckpoint(mediaDuration time.Duration, chunks []audioChunk, completedChunks int, cues []subtitleCue) incrementalCheckpoint {
	return newIncrementalCheckpointWithFingerprint(incrementalFingerprint{}, mediaDuration, chunks, completedChunks, cues)
}

func newIncrementalCheckpointWithFingerprint(fingerprint incrementalFingerprint, mediaDuration time.Duration, chunks []audioChunk, completedChunks int, cues []subtitleCue) incrementalCheckpoint {
	return newIncrementalCheckpointWithState(fingerprint, checkpointStageTranscribing, 0, mediaDuration, chunks, completedChunks, cues)
}

func newIncrementalCheckpointWithState(fingerprint incrementalFingerprint, stage string, retryCursor int, mediaDuration time.Duration, chunks []audioChunk, completedChunks int, cues []subtitleCue) incrementalCheckpoint {
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

func newIncrementalFingerprint(inputPath, modelPath, vadModelPath, whisperPath, ffmpegPath, ffprobePath, language string) (incrementalFingerprint, error) {
	paths := []struct {
		label string
		path  string
	}{
		{label: "input", path: inputPath},
		{label: "model", path: modelPath},
		{label: "VAD model", path: vadModelPath},
		{label: "whisper-cli", path: whisperPath},
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
		FFmpeg:          identities[4],
		FFprobe:         identities[5],
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

func checkpointCuesFromSubtitleCues(cues []subtitleCue) []checkpointCue {
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

func checkpointOriginFromTranscriptionOrigin(origin transcriptionOrigin) checkpointOrigin {
	return checkpointOrigin{Index: origin.Index, Start: origin.Start, End: origin.End}
}

func subtitleCuesFromCheckpointCues(cues []checkpointCue) []subtitleCue {
	result := make([]subtitleCue, len(cues))
	for cueIndex, cue := range cues {
		tokens := make([]subtitleToken, len(cue.Tokens))
		for tokenIndex, token := range cue.Tokens {
			tokens[tokenIndex] = subtitleToken{
				Start: token.Start, End: token.End, Text: token.Text, Probability: token.Probability,
				Origin: transcriptionOriginFromCheckpointOrigin(token.Origin),
			}
		}
		result[cueIndex] = subtitleCue{
			Start: cue.Start, End: cue.End, Text: cue.Text, Probability: cue.Probability,
			Tokens: tokens, Origin: transcriptionOriginFromCheckpointOrigin(cue.Origin),
		}
	}
	return result
}

func transcriptionOriginFromCheckpointOrigin(origin checkpointOrigin) transcriptionOrigin {
	return transcriptionOrigin{Index: origin.Index, Start: origin.Start, End: origin.End}
}

func audioChunksFromCheckpoint(chunks []checkpointAudioChunk) []audioChunk {
	result := make([]audioChunk, len(chunks))
	for index, chunk := range chunks {
		result[index] = audioChunk{Start: chunk.Start, End: chunk.End}
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
	if checkpoint.Stage != checkpointStageTranscribing && checkpoint.Stage != checkpointStageRetrying && checkpoint.Stage != checkpointStageRetryComplete {
		return incrementalCheckpoint{}, false, fmt.Errorf("validate transcription checkpoint %q: unsupported stage %q", path, checkpoint.Stage)
	}
	if checkpoint.MediaDuration <= 0 || checkpoint.CompletedChunks < 0 || checkpoint.CompletedChunks > len(checkpoint.Chunks) {
		return incrementalCheckpoint{}, false, fmt.Errorf("validate transcription checkpoint %q: invalid progress", path)
	}
	if checkpoint.RetryCursor < 0 || checkpoint.RetryCursor > len(checkpoint.Cues) {
		return incrementalCheckpoint{}, false, fmt.Errorf("validate transcription checkpoint %q: invalid retry cursor", path)
	}
	if checkpoint.Stage == checkpointStageTranscribing && checkpoint.RetryCursor != 0 {
		return incrementalCheckpoint{}, false, fmt.Errorf("validate transcription checkpoint %q: retry cursor set during main transcription", path)
	}
	if checkpoint.Stage != checkpointStageTranscribing && checkpoint.CompletedChunks != len(checkpoint.Chunks) {
		return incrementalCheckpoint{}, false, fmt.Errorf("validate transcription checkpoint %q: retry stage before main transcription completed", path)
	}
	if checkpoint.Stage == checkpointStageRetryComplete && checkpoint.RetryCursor != len(checkpoint.Cues) {
		return incrementalCheckpoint{}, false, fmt.Errorf("validate transcription checkpoint %q: incomplete retry cursor", path)
	}
	for index, chunk := range checkpoint.Chunks {
		if chunk.Start < 0 || chunk.End <= chunk.Start || chunk.End > checkpoint.MediaDuration {
			return incrementalCheckpoint{}, false, fmt.Errorf("validate transcription checkpoint %q: invalid chunk %d", path, index)
		}
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

func persistIncrementalProgress(outputBase, language string, corrections map[string]string, fingerprint incrementalFingerprint, mediaDuration time.Duration, chunks []audioChunk, completedChunks int, rawCues []subtitleCue) error {
	return persistIncrementalProgressWithState(outputBase, language, corrections, fingerprint, checkpointStageTranscribing, 0, mediaDuration, chunks, completedChunks, rawCues)
}

func persistIncrementalProgressWithState(outputBase, language string, corrections map[string]string, fingerprint incrementalFingerprint, stage string, retryCursor int, mediaDuration time.Duration, chunks []audioChunk, completedChunks int, rawCues []subtitleCue) error {
	partialCues := cleanSubtitleCues(rawCues, language, corrections, mediaDuration)
	if err := validateSubtitleCues(partialCues, mediaDuration); err != nil {
		return fmt.Errorf("validate partial transcript: %w", err)
	}
	partial, err := renderTranscript(partialCues, "srt", language)
	if err != nil {
		return fmt.Errorf("render partial transcript: %w", err)
	}
	partialPath, checkpointPath := incrementalArtifactPaths(outputBase)
	if err := replaceFileAtomically(partialPath, []byte(partial), 0o644); err != nil {
		return fmt.Errorf("write partial transcript %q: %w", partialPath, err)
	}
	checkpoint := newIncrementalCheckpointWithState(fingerprint, stage, retryCursor, mediaDuration, chunks, completedChunks, rawCues)
	if err := persistIncrementalCheckpoint(checkpointPath, checkpoint); err != nil {
		return err
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
