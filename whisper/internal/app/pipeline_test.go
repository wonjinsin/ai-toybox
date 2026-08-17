package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadTranscriptionCuesPreservesChunkOrigin(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	firstPath := filepath.Join(directory, "first.wav")
	secondPath := filepath.Join(directory, "second.wav")
	payload := []byte(`{"transcription":[{"offsets":{"from":100,"to":900},"text":"발화","tokens":[{"text":" 발화","offsets":{"from":100,"to":900},"p":0.9}]}]}`)
	for _, path := range []string{firstPath, secondPath} {
		if err := os.WriteFile(path+".json", payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cues, err := loadTranscriptionCues([]transcriptionChunk{
		{audioChunk: audioChunk{Start: 10 * time.Second, End: 12 * time.Second}, Path: firstPath},
		{audioChunk: audioChunk{Start: 20 * time.Second, End: 23 * time.Second}, Path: secondPath},
	})
	if err != nil {
		t.Fatalf("loadTranscriptionCues() error = %v", err)
	}
	if len(cues) != 2 {
		t.Fatalf("len(cues) = %d, want 2", len(cues))
	}
	if got := cues[1].Origin.Index; got != 1 {
		t.Errorf("second origin index = %d, want 1", got)
	}
	if got := cues[1].Origin.End; got != 23*time.Second {
		t.Errorf("second origin end = %s, want 23s", got)
	}
	if got := cues[1].Tokens[0].Origin; got != cues[1].Origin {
		t.Errorf("token origin = %#v, want %#v", got, cues[1].Origin)
	}
}
