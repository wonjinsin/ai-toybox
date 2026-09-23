package subtitle

import (
	"fmt"
	"strings"
	"time"

	"github.com/wonjinsin/ai-toybox/whisper/internal/domain"
)

func Render(cues []domain.Cue, format, language string) (string, error) {
	if format == "txt" {
		lines := make([]string, 0, len(cues))
		for _, cue := range cues {
			lines = append(lines, strings.ReplaceAll(cue.Text, "\n", " "))
		}
		if len(lines) == 0 {
			return "", nil
		}
		return strings.Join(lines, "\n") + "\n", nil
	}
	if format != "srt" && format != "vtt" {
		return "", fmt.Errorf("unsupported transcript format %q", format)
	}

	separator := ","
	if format == "vtt" {
		separator = "."
	}
	blocks := make([]string, 0, len(cues))
	for index, cue := range cues {
		text, err := wrapSubtitleText(cue.Text, domain.SubtitleLineWidth(language))
		if err != nil {
			return "", err
		}
		timing := fmt.Sprintf("%s --> %s", formatCueTime(cue.Start, separator), formatCueTime(cue.End, separator))
		if format == "srt" {
			blocks = append(blocks, fmt.Sprintf("%d\n%s\n%s", index+1, timing, text))
		} else {
			blocks = append(blocks, fmt.Sprintf("%s\n%s", timing, text))
		}
	}

	prefix := ""
	if format == "vtt" {
		prefix = "WEBVTT\n\n"
	}
	if len(blocks) == 0 {
		return prefix, nil
	}
	return prefix + strings.Join(blocks, "\n\n") + "\n", nil
}

func wrapSubtitleText(text string, width int) (string, error) {
	characters := []rune(text)
	if len(characters) <= width {
		return text, nil
	}
	if len(characters) > 2*width {
		return "", fmt.Errorf("subtitle text has %d characters; maximum is %d", len(characters), 2*width)
	}

	target := len(characters) / 2
	splitAt := target
	bestDistance := len(characters)
	for index, character := range characters {
		candidate := index + 1
		if !strings.ContainsRune("、。！？,.!?", character) || candidate < 6 || len(characters)-candidate < 6 {
			continue
		}
		distance := candidate - target
		if distance < 0 {
			distance = -distance
		}
		if distance < bestDistance && candidate <= width && len(characters)-candidate <= width {
			splitAt = candidate
			bestDistance = distance
		}
	}
	return string(characters[:splitAt]) + "\n" + string(characters[splitAt:]), nil
}

func formatCueTime(t time.Duration, separator string) string {
	return fmt.Sprintf("%02d:%02d:%02d%s%03d",
		int(t.Hours()), int(t.Minutes())%60, int(t.Seconds())%60, separator, t.Milliseconds()%1000)
}
