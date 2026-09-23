package subtitle

import (
	"strings"
	"testing"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
)

func TestRenderTranscriptFormats(t *testing.T) {
	t.Parallel()

	cues := []domain.Cue{{Start: time.Second, End: 2500 * time.Millisecond, Text: "hello\nworld"}}
	for _, test := range []struct {
		format string
		want   string
	}{
		{format: "txt", want: "hello world\n"},
		{format: "srt", want: "1\n00:00:01,000 --> 00:00:02,500\nhello\nworld\n"},
		{format: "vtt", want: "WEBVTT\n\n00:00:01.000 --> 00:00:02.500\nhello\nworld\n"},
	} {
		t.Run(test.format, func(t *testing.T) {
			got, err := Render(cues, test.format, "en")
			if err != nil || got != test.want {
				t.Fatalf("Render() = %q, %v; want %q", got, err, test.want)
			}
		})
	}
}

func TestRenderEmptyTranscript(t *testing.T) {
	t.Parallel()

	for _, format := range []string{"txt", "srt", "vtt"} {
		t.Run(format, func(t *testing.T) {
			want := ""
			if format == "vtt" {
				want = "WEBVTT\n\n"
			}
			got, err := Render(nil, format, "ja")
			if err != nil || got != want {
				t.Fatalf("Render() = %q, %v; want %q", got, err, want)
			}
		})
	}
}

func TestRenderRejectsUnsupportedFormatAndOversizedCue(t *testing.T) {
	t.Parallel()

	if _, err := Render(nil, "json", "en"); err == nil || !strings.Contains(err.Error(), "unsupported transcript format") {
		t.Fatalf("Render() error = %v, want unsupported format", err)
	}
	cues := []domain.Cue{{Text: strings.Repeat("字", 37)}}
	if _, err := Render(cues, "srt", "ja"); err == nil || !strings.Contains(err.Error(), "maximum is 36") {
		t.Fatalf("Render() error = %v, want readability error", err)
	}
}

func TestPostProcessPreservesSilenceGapBetweenSpeechSegments(t *testing.T) {
	t.Parallel()

	input := "1\n00:00:09,890 --> 00:00:12,670\nfirst\n\n2\n00:00:12,670 --> 00:00:35,160\nsecond\n"
	segments := []domain.SpeechSegment{
		{Start: 9890 * time.Millisecond, End: 12700 * time.Millisecond},
		{Start: 32650 * time.Millisecond, End: 40 * time.Second},
	}
	want := "1\n00:00:09,890 --> 00:00:12,670\nfirst\n\n2\n00:00:32,650 --> 00:00:35,160\nsecond\n"
	if got := PostProcess(input, segments); got != want {
		t.Fatalf("PostProcess() = %q, want %q", got, want)
	}
}

func TestContainsCueTiming(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		content string
		want    bool
	}{
		{content: "WEBVTT\n\n00:00:01.000 --> 00:00:02.500\nhello\n", want: true},
		{content: "1\n00:00:01,000 --> 00:00:02,500\nhello\n", want: true},
		{content: "hello\nworld\n", want: false},
	} {
		if got := ContainsCueTiming(test.content); got != test.want {
			t.Errorf("ContainsCueTiming(%q) = %t, want %t", test.content, got, test.want)
		}
	}
}
