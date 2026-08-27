package app

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildSpeechChunksKeepsOriginalTimelineAndTwentySecondLimit(t *testing.T) {
	t.Parallel()

	segments := []speechSegment{
		{Start: 10 * time.Second, End: 12 * time.Second},
		{Start: 12500 * time.Millisecond, End: 15 * time.Second},
		{Start: 30 * time.Second, End: 55 * time.Second},
	}

	got := buildSpeechChunks(segments, 60*time.Second)
	want := []audioChunk{
		{Start: 9850 * time.Millisecond, End: 15150 * time.Millisecond},
		{Start: 29850 * time.Millisecond, End: 49850 * time.Millisecond},
		{Start: 49550 * time.Millisecond, End: 55150 * time.Millisecond},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("buildSpeechChunks() = %#v, want %#v", got, want)
	}
	for _, chunk := range got {
		if duration := chunk.End - chunk.Start; duration > 20*time.Second {
			t.Errorf("chunk duration = %s, want at most 20s", duration)
		}
	}
}

func TestBuildSpeechChunksCoalescesShortSilenceSplitByDedicatedVAD(t *testing.T) {
	t.Parallel()

	combined := buildSpeechChunks([]speechSegment{{Start: time.Second, End: 4 * time.Second}}, 10*time.Second)
	split := buildSpeechChunks([]speechSegment{
		{Start: time.Second, End: 2500 * time.Millisecond},
		{Start: 2500 * time.Millisecond, End: 4 * time.Second},
	}, 10*time.Second)

	if !reflect.DeepEqual(split, combined) {
		t.Errorf("split VAD chunks = %#v, want legacy-equivalent chunks %#v", split, combined)
	}
}

func TestBuildSpeechChunksKeepsLanguageContextAcrossConversationalPause(t *testing.T) {
	t.Parallel()

	got := buildSpeechChunks([]speechSegment{
		{Start: 10 * time.Second, End: 12 * time.Second},
		{Start: 14 * time.Second, End: 16 * time.Second},
	}, 30*time.Second)
	want := []audioChunk{{Start: 9850 * time.Millisecond, End: 16150 * time.Millisecond}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("buildSpeechChunks() = %#v, want one context-preserving chunk %#v", got, want)
	}
}
