package domain

import (
	"testing"
	"time"
)

func TestSelectRetryCueUsesSpecifiedConfidenceThresholds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		original  float64
		candidate float64
		wantRetry bool
	}{
		{name: "rescues dropped cue", original: 0.31, candidate: 0.32, wantRetry: true},
		{name: "requires five-point improvement", original: 0.70, candidate: 0.74, wantRetry: false},
		{name: "accepts five-point improvement", original: 0.70, candidate: 0.75, wantRetry: true},
		{name: "keeps candidate below final threshold", original: 0.31, candidate: 0.31, wantRetry: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := Cue{Text: "original", Probability: test.original}
			candidate := Cue{Text: "retry", Probability: test.candidate}
			got := SelectRetryCue(original, candidate)
			if (got.Text == "retry") != test.wantRetry {
				t.Errorf("SelectRetryCue() = %q, want retry selected=%t", got.Text, test.wantRetry)
			}
		})
	}
}

func TestRetryWindowForCueClampsToMediaBounds(t *testing.T) {
	t.Parallel()

	window, err := RetryWindowForCue(Cue{Start: 500 * time.Millisecond, End: 9500 * time.Millisecond}, 10*time.Second)
	if err != nil {
		t.Fatalf("RetryWindowForCue() error = %v", err)
	}
	if window.Start != 0 || window.End != 10*time.Second {
		t.Errorf("retry window = %#v, want 0s --> 10s", window)
	}
}

func TestRetryCandidateForCueKeepsOriginalWhenRetryHasNoValidTokens(t *testing.T) {
	t.Parallel()

	original := Cue{
		Start:       time.Second,
		End:         2 * time.Second,
		Text:        "원본",
		Probability: 0.40,
	}
	candidate, err := RetryCandidateForCue(original, []Cue{{Text: "무효 재시도"}})
	if err != nil {
		t.Fatalf("RetryCandidateForCue() error = %v", err)
	}
	if candidate.Text != original.Text || candidate.Start != original.Start || candidate.End != original.End {
		t.Errorf("RetryCandidateForCue() = %#v, want original %#v", candidate, original)
	}
}

func TestRetryCandidateForCueKeepsOriginalWhenTokensAreOutsideOriginalRange(t *testing.T) {
	t.Parallel()

	original := Cue{Start: time.Second, End: 2 * time.Second, Text: "원본", Probability: 0.40}
	retryCues := []Cue{{
		Text: "범위 밖",
		Tokens: []Token{{
			Start:       3 * time.Second,
			End:         4 * time.Second,
			Text:        " 범위 밖",
			Probability: 0.90,
		}},
	}}
	candidate, err := RetryCandidateForCue(original, retryCues)
	if err != nil {
		t.Fatalf("RetryCandidateForCue() error = %v", err)
	}
	if candidate.Text != original.Text {
		t.Errorf("RetryCandidateForCue() text = %q, want original %q", candidate.Text, original.Text)
	}
}

func TestRetryCandidateForCueUsesValidCueAfterCueWithoutTokens(t *testing.T) {
	t.Parallel()

	original := Cue{Start: time.Second, End: 2 * time.Second, Text: "원본", Probability: 0.40}
	retryCues := []Cue{
		{Text: "무효 재시도"},
		{
			Text: "유효 재시도",
			Tokens: []Token{{
				Start:       1100 * time.Millisecond,
				End:         1900 * time.Millisecond,
				Text:        " 유효 재시도",
				Probability: 0.90,
			}},
		},
	}
	candidate, err := RetryCandidateForCue(original, retryCues)
	if err != nil {
		t.Fatalf("RetryCandidateForCue() error = %v", err)
	}
	if candidate.Text != "유효 재시도" {
		t.Errorf("RetryCandidateForCue() text = %q, want valid retry", candidate.Text)
	}
}
