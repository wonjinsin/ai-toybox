package subtitle

import (
	"testing"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
)

func TestClampCueDurationsSRT(t *testing.T) {
	input := `1
00:39:05,260 --> 00:48:02,920
こんにちは

2
00:48:05,930 --> 00:48:07,100
今日はいい天気ですね
`
	want := `1
00:39:05,260 --> 00:39:20,260
こんにちは

2
00:48:05,930 --> 00:48:07,100
今日はいい天気ですね
`
	if got := clampCueDurations(input); got != want {
		t.Errorf("clampCueDurations() = %q, want %q", got, want)
	}
}

func TestClampCueDurationsVTTKeepsDotSeparator(t *testing.T) {
	input := `WEBVTT

01:06:16.880 --> 01:21:34.100
こんにちは
`
	want := `WEBVTT

01:06:16.880 --> 01:06:31.880
こんにちは
`
	if got := clampCueDurations(input); got != want {
		t.Errorf("clampCueDurations() = %q, want %q", got, want)
	}
}

func TestCollapseRepeatedCuesDropsRunOfThreeOrMore(t *testing.T) {
	input := `1
00:48:00,000 --> 00:48:01,000
ご視聴ありがとうございました

2
00:48:01,000 --> 00:48:02,000
ご視聴ありがとうございました

3
00:48:02,000 --> 00:48:03,000
ご視聴ありがとうございました

4
00:48:03,000 --> 00:48:04,000
今日はいい天気ですね
`
	want := `1
00:48:00,000 --> 00:48:01,000
ご視聴ありがとうございました

2
00:48:03,000 --> 00:48:04,000
今日はいい天気ですね
`
	if got := collapseRepeatedCues(input); got != want {
		t.Errorf("collapseRepeatedCues() = %q, want %q", got, want)
	}
}

func TestCollapseRepeatedCuesKeepsRunOfTwo(t *testing.T) {
	input := `1
00:00:00,000 --> 00:00:01,000
うん

2
00:00:01,000 --> 00:00:02,000
うん
`
	if got := collapseRepeatedCues(input); got != input {
		t.Errorf("collapseRepeatedCues() = %q, want unchanged", got)
	}
}

func TestCollapseRepeatedCuesVTTWithoutIndexes(t *testing.T) {
	input := `WEBVTT

00:00:00.000 --> 00:00:01.000
ご視聴ありがとうございました

00:00:01.000 --> 00:00:02.000
ご視聴ありがとうございました

00:00:02.000 --> 00:00:03.000
ご視聴ありがとうございました
`
	want := `WEBVTT

00:00:00.000 --> 00:00:01.000
ご視聴ありがとうございました
`
	if got := collapseRepeatedCues(input); got != want {
		t.Errorf("collapseRepeatedCues() = %q, want %q", got, want)
	}
}

func TestAlignCueStartsToSpeechKeepsNormalLeadInAndMultiSegmentCue(t *testing.T) {
	input := `1
00:00:06,700 --> 00:00:09,440
발화보다 조금 먼저 표시

2
00:00:20,000 --> 00:00:29,000
짧은 쉼을 포함한 한 자막
`
	segments := []domain.SpeechSegment{
		{Start: 6950 * time.Millisecond, End: 9440 * time.Millisecond},
		{Start: 22000 * time.Millisecond, End: 24000 * time.Millisecond},
		{Start: 26000 * time.Millisecond, End: 29000 * time.Millisecond},
	}

	if got := alignCueStartsToSpeech(input, segments); got != input {
		t.Errorf("alignCueStartsToSpeech() = %q, want unchanged %q", got, input)
	}
}

func TestRenderTranscriptWrapsJapaneseSRTForReadability(t *testing.T) {
	t.Parallel()

	cues := []domain.Cue{{
		Start: time.Second,
		End:   3 * time.Second,
		Text:  "一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十",
	}}

	got, err := Render(cues, "srt", "ja")
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	want := "1\n00:00:01,000 --> 00:00:03,000\n一二三四五六七八九十一二三四五\n六七八九十一二三四五六七八九十\n"
	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}
