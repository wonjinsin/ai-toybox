package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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

func TestPostProcessTranscriptRewritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.srt")
	if err := os.WriteFile(path, []byte("1\n00:00:00,000 --> 00:10:00,000\nうん\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := postProcessTranscript(path, nil); err != nil {
		t.Fatalf("postProcessTranscript() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "1\n00:00:00,000 --> 00:00:15,000\nうん\n" {
		t.Errorf("file content = %q", content)
	}
}

func TestPostProcessTranscriptPreservesSilenceGapBetweenVADSegments(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.srt")
	input := `1
00:00:09,890 --> 00:00:12,670
첫 번째 발화

2
00:00:12,670 --> 00:00:35,160
두 번째 발화
`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}

	segments := []speechSegment{
		{Start: 9890 * time.Millisecond, End: 12700 * time.Millisecond},
		{Start: 32650 * time.Millisecond, End: 40 * time.Second},
	}
	if err := postProcessTranscript(path, segments); err != nil {
		t.Fatalf("postProcessTranscript() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := `1
00:00:09,890 --> 00:00:12,670
첫 번째 발화

2
00:00:32,650 --> 00:00:35,160
두 번째 발화
`
	if string(content) != want {
		t.Errorf("file content = %q, want %q", content, want)
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
	segments := []speechSegment{
		{Start: 6950 * time.Millisecond, End: 9440 * time.Millisecond},
		{Start: 22000 * time.Millisecond, End: 24000 * time.Millisecond},
		{Start: 26000 * time.Millisecond, End: 29000 * time.Millisecond},
	}

	if got := alignCueStartsToSpeech(input, segments); got != input {
		t.Errorf("alignCueStartsToSpeech() = %q, want unchanged %q", got, input)
	}
}

func TestCleanSubtitleCuesDropsLowConfidenceSpeech(t *testing.T) {
	t.Parallel()

	cues := []subtitleCue{
		{Start: time.Second, End: 2 * time.Second, Text: "残す", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "落とす", Probability: 0.31},
	}

	got := cleanSubtitleCues(cues, "ja", nil, 10*time.Second)
	if len(got) != 1 || got[0].Text != "残す" {
		t.Fatalf("cleanSubtitleCues() = %#v, want only high-confidence cue", got)
	}
}

func TestCleanSubtitleCuesDropsJapaneseVocalizationButKeepsResponse(t *testing.T) {
	t.Parallel()

	cues := []subtitleCue{
		{Start: time.Second, End: 2 * time.Second, Text: "あー", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "はい", Probability: 0.9},
	}

	got := cleanSubtitleCues(cues, "ja", nil, 10*time.Second)
	if len(got) != 1 || got[0].Text != "はい" {
		t.Fatalf("cleanSubtitleCues() = %#v, want meaningful response only", got)
	}
}

func TestCleanSubtitleCuesDropsNearbyDuplicate(t *testing.T) {
	t.Parallel()

	cues := []subtitleCue{
		{Start: time.Second, End: 2 * time.Second, Text: "こんにちは", Probability: 0.9},
		{Start: 3 * time.Second, End: 4 * time.Second, Text: "こんにちは", Probability: 0.9},
		{Start: 20 * time.Second, End: 21 * time.Second, Text: "こんにちは", Probability: 0.9},
	}

	got := cleanSubtitleCues(cues, "ja", nil, 30*time.Second)
	if len(got) != 2 || got[0].Start != time.Second || got[1].Start != 20*time.Second {
		t.Fatalf("cleanSubtitleCues() = %#v, want distant duplicates only", got)
	}
}

func TestCleanSubtitleCuesAppliesExternalCorrections(t *testing.T) {
	t.Parallel()

	cues := []subtitleCue{{
		Start: time.Second, End: 2 * time.Second, Text: "レッサン帰り?", Probability: 0.9,
	}}

	got := cleanSubtitleCues(cues, "ja", map[string]string{"レッサン": "レッスン"}, 10*time.Second)
	if len(got) != 1 || got[0].Text != "レッスン帰り？" {
		t.Fatalf("cleanSubtitleCues() = %#v, want corrected Japanese text", got)
	}
}

func TestCleanSubtitleCuesCapsDisplayTimeBeforeNextCue(t *testing.T) {
	t.Parallel()

	cues := []subtitleCue{
		{Start: time.Second, End: 20 * time.Second, Text: "첫 번째", Probability: 0.9},
		{Start: 5 * time.Second, End: 6 * time.Second, Text: "두 번째", Probability: 0.9},
	}

	got := cleanSubtitleCues(cues, "ko", nil, 10*time.Second)
	if len(got) != 2 {
		t.Fatalf("len(cleanSubtitleCues()) = %d, want 2", len(got))
	}
	if got[0].End != 4950*time.Millisecond {
		t.Errorf("first cue end = %s, want 4.95s", got[0].End)
	}
	if got[0].End > got[1].Start {
		t.Errorf("cues overlap: %#v", got)
	}
}

func TestRenderTranscriptWrapsJapaneseSRTForReadability(t *testing.T) {
	t.Parallel()

	cues := []subtitleCue{{
		Start: time.Second,
		End:   3 * time.Second,
		Text:  "一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十",
	}}

	got, err := renderTranscript(cues, "srt", "ja")
	if err != nil {
		t.Fatalf("renderTranscript() error = %v", err)
	}
	want := "1\n00:00:01,000 --> 00:00:03,000\n一二三四五六七八九十一二三四五\n六七八九十一二三四五六七八九十\n"
	if got != want {
		t.Fatalf("renderTranscript() = %q, want %q", got, want)
	}
}

func TestCleanSubtitleCuesSplitsTextBeyondTwoLines(t *testing.T) {
	t.Parallel()

	text := "一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十一二三四五六七八九十"
	cues := []subtitleCue{{
		Start: time.Second, End: 5 * time.Second, Text: text, Probability: 0.9,
	}}

	got := cleanSubtitleCues(cues, "ja", nil, 10*time.Second)
	if len(got) != 2 {
		t.Fatalf("len(cleanSubtitleCues()) = %d, want 2", len(got))
	}
	if got[0].Text+got[1].Text != text {
		t.Errorf("split text = %q + %q, want original %q", got[0].Text, got[1].Text, text)
	}
	for _, cue := range got {
		if length := len([]rune(cue.Text)); length > 36 {
			t.Errorf("split cue length = %d, want at most 36", length)
		}
	}
}
