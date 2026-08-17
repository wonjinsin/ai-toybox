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

func TestParseWhisperJSONIgnoresInvalidTokenTimingForCueConfidence(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
  "transcription": [
    {
      "offsets": {"from": 1000, "to": 3000},
      "text": "정상 토큰과 손상 토큰",
      "tokens": [
        {"text": " 정상", "offsets": {"from": 1000, "to": 2000}, "p": 0.9},
        {"text": " 손상", "offsets": {"from": 2500, "to": 2200}, "p": 0.1}
      ]
    }
  ]
}`)

	cues, err := parseWhisperJSON(payload, 0)
	if err != nil {
		t.Fatalf("parseWhisperJSON() error = %v", err)
	}
	if len(cues) != 1 {
		t.Fatalf("len(cues) = %d, want 1", len(cues))
	}
	if math.Abs(cues[0].Probability-0.9) > 0.0001 {
		t.Errorf("cue probability = %f, want 0.9", cues[0].Probability)
	}
}

func TestParseWhisperJSONPreservesAbsoluteLexicalAndPunctuationTokens(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
  "transcription": [
    {
      "offsets": {"from": 1000, "to": 2200},
      "text": "안녕하세요!",
      "tokens": [
        {"text": "[_BEG_]", "offsets": {"from": 1000, "to": 1000}, "p": 1.0},
        {"text": " 안녕하세요", "offsets": {"from": 1000, "to": 2100}, "p": 0.9},
        {"text": "!", "offsets": {"from": 2100, "to": 2200}, "p": 0.1}
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

	tokens := cues[0].Tokens
	if len(tokens) != 2 {
		t.Fatalf("token count = %d, want 2", len(tokens))
	}
	first := tokens[0]
	if got := first.Text; got != " 안녕하세요" {
		t.Errorf("first token text = %q, want %q", got, " 안녕하세요")
	}
	if got := first.Start; got != 31*time.Second {
		t.Errorf("first token start = %s, want 31s", got)
	}
	if got := first.End; got != 32100*time.Millisecond {
		t.Errorf("first token end = %s, want 32.1s", got)
	}
}

func TestParseWhisperJSONPreservesChunkOriginOnTokens(t *testing.T) {
	t.Parallel()

	payload := []byte(`{
  "transcription": [
    {
      "offsets": {"from": 100, "to": 900},
      "text": "발화",
      "tokens": [
        {"text": " 발화", "offsets": {"from": 100, "to": 900}, "p": 0.9}
      ]
    }
  ]
}`)

	cues, err := parseWhisperJSON(payload, 12*time.Second)
	if err != nil {
		t.Fatalf("parseWhisperJSON() error = %v", err)
	}
	origin := cues[0].Tokens[0].Origin
	if got := origin.Start; got != 12*time.Second {
		t.Errorf("origin start = %s, want 12s", got)
	}
}
