package app

import (
	"math"
	"testing"
	"time"
)

func TestParseWhisperJSONAddsOriginalChunkOffset(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
  "transcription": [
    {
      "offsets": {"from": 1000, "to": 2200},
      "text": " こんにちは？",
      "tokens": [
        {"text": " こんにちは", "offsets": {"from": 1000, "to": 2100}, "p": 0.9},
        {"text": "？", "offsets": {"from": 2100, "to": 2200}, "p": 0.1}
      ]
    }
  ]
}`)

	cues, err := parseWhisperJSON(payload, 30*time.Second)
	if err != nil {
		t.Fatalf("parseWhisperJSON() error = %v", err)
	}
	if len(cues) != 1 {
		t.Fatalf("len(cues) = %d, want 1", len(cues))
	}
	cue := cues[0]
	if cue.Start != 31*time.Second || cue.End != 32200*time.Millisecond {
		t.Errorf("cue timing = %s --> %s, want 31s --> 32.2s", cue.Start, cue.End)
	}
	if cue.Text != "こんにちは？" {
		t.Errorf("cue text = %q, want %q", cue.Text, "こんにちは？")
	}
	if math.Abs(cue.Probability-0.9) > 0.0001 {
		t.Errorf("cue probability = %f, want 0.9", cue.Probability)
	}
}
