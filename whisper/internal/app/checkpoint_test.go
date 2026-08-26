package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestIncrementalCheckpointJSONPreservesCueDetails(t *testing.T) {
	origin := transcriptionOrigin{Index: 37, Start: 10 * time.Second, End: 25 * time.Second}
	checkpoint := newIncrementalCheckpoint(90*time.Second, []audioChunk{{Start: 10 * time.Second, End: 25 * time.Second}}, 1, []subtitleCue{{
		Start:       11 * time.Second,
		End:         12 * time.Second,
		Text:        "자막",
		Probability: 0.87,
		Tokens: []subtitleToken{{
			Start:       11 * time.Second,
			End:         12 * time.Second,
			Text:        "자막",
			Probability: 0.91,
			Origin:      origin,
		}},
		Origin: origin,
	}})

	content, err := json.Marshal(checkpoint)
	if err != nil {
		t.Fatal(err)
	}
	var decoded incrementalCheckpoint
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, checkpoint) {
		t.Errorf("decoded checkpoint = %#v, want %#v", decoded, checkpoint)
	}
}

func TestLoadIncrementalCheckpointRejectsDifferentFingerprint(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, ".input.srt.whisper-local-checkpoint.json")
	fingerprint := incrementalFingerprint{PipelineVersion: incrementalPipelineVersion, Language: "ja"}
	checkpoint := newIncrementalCheckpointWithFingerprint(fingerprint, 60*time.Second, []audioChunk{{Start: time.Second, End: 2 * time.Second}}, 1, nil)
	if err := persistIncrementalCheckpoint(path, checkpoint); err != nil {
		t.Fatal(err)
	}

	changed := fingerprint
	changed.Language = "ko"
	_, found, err := loadIncrementalCheckpoint(path, changed)
	if err != nil {
		t.Fatalf("loadIncrementalCheckpoint() error = %v", err)
	}
	if found {
		t.Fatal("loadIncrementalCheckpoint() found = true, want false")
	}
}

func TestLoadIncrementalCheckpointRejectsVersionBeforeQuietSpeechNormalization(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, ".input.srt.whisper-local-checkpoint.json")
	legacyFingerprint := incrementalFingerprint{PipelineVersion: 1, Language: "ko"}
	checkpoint := newIncrementalCheckpointWithFingerprint(legacyFingerprint, 60*time.Second, []audioChunk{{Start: time.Second, End: 2 * time.Second}}, 1, nil)
	if err := persistIncrementalCheckpoint(path, checkpoint); err != nil {
		t.Fatal(err)
	}

	currentFingerprint := incrementalFingerprint{PipelineVersion: incrementalPipelineVersion, Language: "ko"}
	_, found, err := loadIncrementalCheckpoint(path, currentFingerprint)
	if err != nil {
		t.Fatalf("loadIncrementalCheckpoint() error = %v", err)
	}
	if found {
		t.Fatal("loadIncrementalCheckpoint() found = true, want false for checkpoint created before quiet-speech normalization")
	}
}

func TestLoadIncrementalCheckpointReportsCorruptFileWithoutRemovingIt(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, ".input.srt.whisper-local-checkpoint.json")
	content := []byte(`{"version":`)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}

	_, found, err := loadIncrementalCheckpoint(path, incrementalFingerprint{})
	if err == nil {
		t.Fatal("loadIncrementalCheckpoint() error = nil, want decode error")
	}
	if found {
		t.Fatal("loadIncrementalCheckpoint() found = true, want false")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error = %q, want checkpoint path %q", err, path)
	}
	stored, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(stored) != string(content) {
		t.Errorf("corrupt checkpoint = %q, want preserved content %q", stored, content)
	}
}

func TestLoadIncrementalCheckpointRejectsRetryCursorDuringTranscription(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, ".input.srt.whisper-local-checkpoint.json")
	fingerprint := incrementalFingerprint{PipelineVersion: incrementalPipelineVersion, Language: "ja"}
	checkpoint := newIncrementalCheckpointWithState(fingerprint, checkpointStageTranscribing, 1, 60*time.Second, []audioChunk{{Start: time.Second, End: 2 * time.Second}}, 1, []subtitleCue{{
		Start: time.Second, End: 2 * time.Second, Text: "invalid cursor", Probability: 0.9,
	}})
	if err := persistIncrementalCheckpoint(path, checkpoint); err != nil {
		t.Fatal(err)
	}

	_, found, err := loadIncrementalCheckpoint(path, fingerprint)
	if err == nil {
		t.Fatal("loadIncrementalCheckpoint() error = nil, want stage validation error")
	}
	if found {
		t.Fatal("loadIncrementalCheckpoint() found = true, want false")
	}
}

func TestReplaceFileAtomicallyReplacesContentWithoutTemporaryFile(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "input.partial.srt")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := replaceFileAtomically(path, []byte("new"), 0o644); err != nil {
		t.Fatalf("replaceFileAtomically() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "new" {
		t.Errorf("content = %q, want new", content)
	}
	temporaryFiles, err := filepath.Glob(filepath.Join(directory, ".input.partial.srt.tmp-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(temporaryFiles) != 0 {
		t.Errorf("temporary files remain: %v", temporaryFiles)
	}
}
