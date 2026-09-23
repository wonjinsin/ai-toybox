package domain

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCleanSubtitleCuesPreservesInput(t *testing.T) {
	t.Parallel()

	first := strings.Repeat("前", 20)
	second := strings.Repeat("後", 20)
	cues := []Cue{
		{Start: 8 * time.Second, End: 20 * time.Second, Text: " 終了? ", Probability: 0.9},
		{
			Start: time.Second, End: 7 * time.Second, Text: first + second, Probability: 0.9,
			Tokens: []Token{
				{Start: time.Second, End: 3 * time.Second, Text: first},
				{Start: 4 * time.Second, End: 7 * time.Second, Text: second},
			},
		},
	}
	before := append([]Cue(nil), cues...)
	before[1].Tokens = append([]Token(nil), cues[1].Tokens...)
	corrections := map[string]string{"終了": "完了"}

	got := CleanSubtitleCues(cues, "ja", corrections, 12*time.Second)
	if len(got) != 3 || got[2].Text != "完了？" || got[2].End != 12*time.Second {
		t.Fatalf("CleanSubtitleCues() = %#v, want split, corrected, and clamped cues", got)
	}
	if !reflect.DeepEqual(cues, before) {
		t.Fatalf("input cues changed: got %#v, want %#v", cues, before)
	}
	if !reflect.DeepEqual(corrections, map[string]string{"終了": "完了"}) {
		t.Fatalf("input corrections changed: %#v", corrections)
	}
}

func TestRetryCandidateForCuePreservesInputTokenOrigin(t *testing.T) {
	t.Parallel()

	original := Cue{Start: time.Second, End: 2 * time.Second, Origin: Origin{Index: 4}}
	retryOrigin := Origin{Index: 7, Start: time.Second, End: 3 * time.Second}
	retryCues := []Cue{{Tokens: []Token{{
		Start: 1100 * time.Millisecond, End: 1900 * time.Millisecond,
		Text: " speech", Probability: 0.9, Origin: retryOrigin,
	}}}}

	candidate, err := RetryCandidateForCue(original, retryCues)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidate.Tokens) != 1 || candidate.Tokens[0].Origin != original.Origin {
		t.Fatalf("candidate = %#v, want original cue origin", candidate)
	}
	if retryCues[0].Tokens[0].Origin != retryOrigin {
		t.Fatalf("retry token origin changed: %#v", retryCues[0].Tokens[0].Origin)
	}
}
