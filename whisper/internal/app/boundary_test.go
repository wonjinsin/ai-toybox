package app

import (
	"testing"
	"time"
)

func TestCleanSubtitleCuesPrefersHigherConfidenceOverlappingChunkDuplicate(t *testing.T) {
	t.Parallel()

	firstOrigin := transcriptionOrigin{Index: 0, Start: 0, End: 20 * time.Second}
	secondOrigin := transcriptionOrigin{Index: 1, Start: 19800 * time.Millisecond, End: 40 * time.Second}
	cues := []subtitleCue{
		{
			Start: 19800 * time.Millisecond, End: 20 * time.Second, Text: "같은 발화", Probability: 0.6, Origin: firstOrigin,
			Tokens: []subtitleToken{{Start: 19800 * time.Millisecond, End: 20 * time.Second, Text: " 같은 발화", Probability: 0.6, Origin: firstOrigin}},
		},
		{
			Start: 19800 * time.Millisecond, End: 20 * time.Second, Text: "같은 발화", Probability: 0.9, Origin: secondOrigin,
			Tokens: []subtitleToken{{Start: 19800 * time.Millisecond, End: 20 * time.Second, Text: " 같은 발화", Probability: 0.9, Origin: secondOrigin}},
		},
	}

	got := cleanSubtitleCues(cues, "ko", nil, 40*time.Second)
	if len(got) != 1 {
		t.Fatalf("len(cleanSubtitleCues()) = %d, want 1", len(got))
	}
	if got[0].Probability != 0.9 || got[0].Origin.Index != 1 {
		t.Errorf("kept cue = %#v, want higher-confidence second chunk cue", got[0])
	}
}

func TestReconcileChunkBoundariesPrefersCompleteContainingCue(t *testing.T) {
	t.Parallel()

	firstOrigin := transcriptionOrigin{Index: 0, Start: 0, End: 20 * time.Second}
	secondOrigin := transcriptionOrigin{Index: 1, Start: 19800 * time.Millisecond, End: 40 * time.Second}
	cues := []subtitleCue{
		{Start: 19800 * time.Millisecond, End: 20 * time.Second, Text: "세계", Probability: 0.95, Origin: firstOrigin},
		{Start: 19800 * time.Millisecond, End: 21 * time.Second, Text: "안녕 세계", Probability: 0.8, Origin: secondOrigin},
	}

	got := reconcileChunkBoundaries(cues)
	if len(got) != 1 {
		t.Fatalf("len(reconcileChunkBoundaries()) = %d, want 1", len(got))
	}
	if got[0].Text != "안녕 세계" || got[0].Origin.Index != 1 {
		t.Errorf("kept cue = %#v, want complete second chunk cue", got[0])
	}
}

func TestReconcileChunkBoundariesMergesSharedTokenAtChunkBoundary(t *testing.T) {
	t.Parallel()

	firstOrigin := transcriptionOrigin{Index: 0, Start: 0, End: 20 * time.Second}
	secondOrigin := transcriptionOrigin{Index: 1, Start: 19800 * time.Millisecond, End: 40 * time.Second}
	cues := []subtitleCue{
		{
			Start: 19800 * time.Millisecond, End: 20 * time.Second, Text: "안녕 세계", Probability: 0.9, Origin: firstOrigin,
			Tokens: []subtitleToken{
				{Start: 19800 * time.Millisecond, End: 19900 * time.Millisecond, Text: " 안녕", Probability: 0.9, Origin: firstOrigin},
				{Start: 19900 * time.Millisecond, End: 20 * time.Second, Text: " 세계", Probability: 0.9, Origin: firstOrigin},
			},
		},
		{
			Start: 19900 * time.Millisecond, End: 21 * time.Second, Text: "세계 반가워", Probability: 0.8, Origin: secondOrigin,
			Tokens: []subtitleToken{
				{Start: 19900 * time.Millisecond, End: 20 * time.Second, Text: " 세계", Probability: 0.8, Origin: secondOrigin},
				{Start: 20 * time.Second, End: 21 * time.Second, Text: " 반가워", Probability: 0.8, Origin: secondOrigin},
			},
		},
	}

	got := reconcileChunkBoundaries(cues)
	if len(got) != 1 {
		t.Fatalf("len(reconcileChunkBoundaries()) = %d, want 1", len(got))
	}
	if got[0].Text != "안녕 세계 반가워" {
		t.Errorf("merged text = %q, want %q", got[0].Text, "안녕 세계 반가워")
	}
	if len(got[0].Tokens) != 3 {
		t.Errorf("merged token count = %d, want 3", len(got[0].Tokens))
	}
}

func TestReconcileChunkBoundariesKeepsAmbiguousOverlappingCues(t *testing.T) {
	t.Parallel()

	firstOrigin := transcriptionOrigin{Index: 0, Start: 0, End: 20 * time.Second}
	secondOrigin := transcriptionOrigin{Index: 1, Start: 19800 * time.Millisecond, End: 40 * time.Second}
	cues := []subtitleCue{
		{Start: 19800 * time.Millisecond, End: 20 * time.Second, Text: "첫 발화", Probability: 0.9, Origin: firstOrigin},
		{Start: 19800 * time.Millisecond, End: 21 * time.Second, Text: "다른 발화", Probability: 0.9, Origin: secondOrigin},
	}

	got := reconcileChunkBoundaries(cues)
	if len(got) != 2 {
		t.Fatalf("len(reconcileChunkBoundaries()) = %d, want 2", len(got))
	}
}
