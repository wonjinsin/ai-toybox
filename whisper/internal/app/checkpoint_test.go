package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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
