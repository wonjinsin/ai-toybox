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
	transcriptionBatchSize       = 32
)

type incrementalCheckpoint struct {
	Version         int                    `json:"version"`
	Stage           string                 `json:"stage"`
	BatchSize       int                    `json:"batch_size"`
	MediaDuration   time.Duration          `json:"media_duration"`
	CompletedChunks int                    `json:"completed_chunks"`
	RetryCursor     int                    `json:"retry_cursor"`
	Chunks          []checkpointAudioChunk `json:"chunks"`
	Cues            []checkpointCue        `json:"cues"`
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
	checkpointChunks := make([]checkpointAudioChunk, len(chunks))
	for index, chunk := range chunks {
		checkpointChunks[index] = checkpointAudioChunk{Start: chunk.Start, End: chunk.End}
	}
	return incrementalCheckpoint{
		Version:         incrementalCheckpointVersion,
		Stage:           "transcribing",
		BatchSize:       transcriptionBatchSize,
		MediaDuration:   mediaDuration,
		CompletedChunks: completedChunks,
		Chunks:          checkpointChunks,
		Cues:            checkpointCuesFromSubtitleCues(cues),
	}
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

func persistIncrementalProgress(outputBase, language string, corrections map[string]string, mediaDuration time.Duration, chunks []audioChunk, completedChunks int, rawCues []subtitleCue) error {
	partialCues := cleanSubtitleCues(rawCues, language, corrections, mediaDuration)
	if err := validateSubtitleCues(partialCues, mediaDuration); err != nil {
		return fmt.Errorf("validate partial transcript: %w", err)
	}
	partial, err := renderTranscript(partialCues, "srt", language)
	if err != nil {
		return fmt.Errorf("render partial transcript: %w", err)
	}
	checkpointContent, err := json.Marshal(newIncrementalCheckpoint(mediaDuration, chunks, completedChunks, rawCues))
	if err != nil {
		return fmt.Errorf("encode transcription checkpoint: %w", err)
	}
	partialPath, checkpointPath := incrementalArtifactPaths(outputBase)
	if err := replaceFileAtomically(partialPath, []byte(partial), 0o644); err != nil {
		return fmt.Errorf("write partial transcript %q: %w", partialPath, err)
	}
	if err := replaceFileAtomically(checkpointPath, checkpointContent, 0o600); err != nil {
		return fmt.Errorf("write transcription checkpoint %q: %w", checkpointPath, err)
	}
	return nil
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
