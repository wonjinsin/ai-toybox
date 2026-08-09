package app

import (
	"os"
	"path/filepath"
	"testing"
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

	if err := postProcessTranscript(path); err != nil {
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
