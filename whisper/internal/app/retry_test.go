package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRetryLowConfidenceCuesBatchesModelInference(t *testing.T) {
	if testing.Short() {
		t.Skip("test uses helper processes")
	}

	directory := t.TempDir()
	ffmpegPath := filepath.Join(directory, "ffmpeg")
	writeExecutable(t, ffmpegPath, "#!/bin/sh\nset -eu\nfor argument do output=\"$argument\"; done\n: > \"$output\"\n")

	cues := make([]subtitleCue, 129)
	for index := range cues {
		start := time.Duration(index*3+2) * time.Second
		cues[index] = subtitleCue{
			Start:       start,
			End:         start + time.Second,
			Text:        "원본",
			Probability: 0.40,
		}
	}
	var inferenceBatchSizes []int
	runner := newWhisperCommandRunner(1, func(_ context.Context, _ string, arguments []string) (string, error) {
		batchSize := 0
		for _, argument := range arguments {
			if filepath.Ext(argument) != ".wav" {
				continue
			}
			batchSize++
			payload := []byte(`{"transcription":[{"offsets":{"from":1100,"to":1900},"text":"개선","tokens":[{"text":" 개선","offsets":{"from":1100,"to":1900},"p":0.90}]}]}`)
			if err := os.WriteFile(argument+".json", payload, 0o644); err != nil {
				return "", err
			}
		}
		inferenceBatchSizes = append(inferenceBatchSizes, batchSize)
		return "", nil
	})

	got, err := retryLowConfidenceCues(
		context.Background(), runner, ffmpegPath, "whisper-cli", "model.bin", "ko", "audio.wav", directory, cues, 390*time.Second,
	)
	if err != nil {
		t.Fatalf("retryLowConfidenceCues() error = %v", err)
	}
	wantBatchSizes := []int{128, 1}
	if len(inferenceBatchSizes) != len(wantBatchSizes) {
		t.Fatalf("inference batch sizes = %v, want %v", inferenceBatchSizes, wantBatchSizes)
	}
	for index := range wantBatchSizes {
		if inferenceBatchSizes[index] != wantBatchSizes[index] {
			t.Errorf("inference batch sizes = %v, want %v", inferenceBatchSizes, wantBatchSizes)
		}
	}
	for index, cue := range got {
		if cue.Text != "개선" {
			t.Errorf("cue %d text = %q, want retry result", index, cue.Text)
		}
	}
}

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
			original := subtitleCue{Text: "original", Probability: test.original}
			candidate := subtitleCue{Text: "retry", Probability: test.candidate}
			got := selectRetryCue(original, candidate)
			if (got.Text == "retry") != test.wantRetry {
				t.Errorf("selectRetryCue() = %q, want retry selected=%t", got.Text, test.wantRetry)
			}
		})
	}
}

func TestRetryWindowForCueClampsToMediaBounds(t *testing.T) {
	t.Parallel()

	window, err := retryWindowForCue(subtitleCue{Start: 500 * time.Millisecond, End: 9500 * time.Millisecond}, 10*time.Second)
	if err != nil {
		t.Fatalf("retryWindowForCue() error = %v", err)
	}
	if window.Start != 0 || window.End != 10*time.Second {
		t.Errorf("retry window = %#v, want 0s --> 10s", window)
	}
}

func TestRetryCandidateForCueKeepsOriginalWhenRetryHasNoValidTokens(t *testing.T) {
	t.Parallel()

	original := subtitleCue{
		Start:       time.Second,
		End:         2 * time.Second,
		Text:        "원본",
		Probability: 0.40,
	}
	candidate, err := retryCandidateForCue(original, []subtitleCue{{Text: "무효 재시도"}})
	if err != nil {
		t.Fatalf("retryCandidateForCue() error = %v", err)
	}
	if candidate.Text != original.Text || candidate.Start != original.Start || candidate.End != original.End {
		t.Errorf("retryCandidateForCue() = %#v, want original %#v", candidate, original)
	}
}

func TestRetryCandidateForCueKeepsOriginalWhenTokensAreOutsideOriginalRange(t *testing.T) {
	t.Parallel()

	original := subtitleCue{Start: time.Second, End: 2 * time.Second, Text: "원본", Probability: 0.40}
	retryCues := []subtitleCue{{
		Text: "범위 밖",
		Tokens: []subtitleToken{{
			Start:       3 * time.Second,
			End:         4 * time.Second,
			Text:        " 범위 밖",
			Probability: 0.90,
		}},
	}}
	candidate, err := retryCandidateForCue(original, retryCues)
	if err != nil {
		t.Fatalf("retryCandidateForCue() error = %v", err)
	}
	if candidate.Text != original.Text {
		t.Errorf("retryCandidateForCue() text = %q, want original %q", candidate.Text, original.Text)
	}
}

func TestRetryCandidateForCueUsesValidCueAfterCueWithoutTokens(t *testing.T) {
	t.Parallel()

	original := subtitleCue{Start: time.Second, End: 2 * time.Second, Text: "원본", Probability: 0.40}
	retryCues := []subtitleCue{
		{Text: "무효 재시도"},
		{
			Text: "유효 재시도",
			Tokens: []subtitleToken{{
				Start:       1100 * time.Millisecond,
				End:         1900 * time.Millisecond,
				Text:        " 유효 재시도",
				Probability: 0.90,
			}},
		},
	}
	candidate, err := retryCandidateForCue(original, retryCues)
	if err != nil {
		t.Fatalf("retryCandidateForCue() error = %v", err)
	}
	if candidate.Text != "유효 재시도" {
		t.Errorf("retryCandidateForCue() text = %q, want valid retry", candidate.Text)
	}
}
