package app

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Sparse VAD segments make whisper.cpp stretch a cue's end time to the next
// segment, producing subtitles that stay on screen for minutes. Real speech
// segments are capped at 30s by -vmsd, so anything longer is a timing artifact.
const maxCueDuration = 15 * time.Second

// Hallucination on non-speech audio repeats the same text across consecutive
// cues; real dialogue rarely repeats an identical line this many times in a row.
const repeatedCueRunLimit = 3

var cueTimingPattern = regexp.MustCompile(`^(\d{2,}):(\d{2}):(\d{2})([,.])(\d{3}) --> (\d{2,}):(\d{2}):(\d{2})[,.](\d{3})$`)

func postProcessTranscript(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read transcript %q: %w", path, err)
	}
	processed := clampCueDurations(collapseRepeatedCues(string(content)))
	return os.WriteFile(path, []byte(processed), 0o644)
}

// collapseRepeatedCues keeps only the first cue of a run of repeatedCueRunLimit
// or more consecutive cues with identical text, then renumbers SRT indexes.
func collapseRepeatedCues(content string) string {
	hadTrailingNewline := strings.HasSuffix(content, "\n")
	blocks := strings.Split(strings.TrimSuffix(content, "\n"), "\n\n")

	texts := make([]string, len(blocks))
	for i, block := range blocks {
		texts[i] = cueText(block)
	}

	kept := make([]string, 0, len(blocks))
	cueNumber := 0
	for i := 0; i < len(blocks); {
		if texts[i] == "" {
			kept = append(kept, blocks[i])
			i++
			continue
		}
		runEnd := i
		for runEnd < len(blocks) && texts[runEnd] == texts[i] {
			runEnd++
		}
		runLength := runEnd - i
		if runLength >= repeatedCueRunLimit {
			runLength = 1
		}
		for k := i; k < i+runLength; k++ {
			kept = append(kept, renumberCue(blocks[k], &cueNumber))
		}
		i = runEnd
	}

	result := strings.Join(kept, "\n\n")
	if hadTrailingNewline {
		result += "\n"
	}
	return result
}

func clampCueDurations(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		match := cueTimingPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		start := cueTime(match[1], match[2], match[3], match[5])
		end := cueTime(match[6], match[7], match[8], match[9])
		if end-start <= maxCueDuration {
			continue
		}
		lines[i] = fmt.Sprintf("%s --> %s", formatCueTime(start, match[4]), formatCueTime(start+maxCueDuration, match[4]))
	}
	return strings.Join(lines, "\n")
}

// cueText returns the text lines of a subtitle block, or "" for non-cue blocks
// such as the WEBVTT header.
func cueText(block string) string {
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		if cueTimingPattern.MatchString(line) {
			return strings.TrimSpace(strings.Join(lines[i+1:], "\n"))
		}
	}
	return ""
}

// renumberCue rewrites an SRT numeric index line with the next sequential
// number; VTT cues without an index line pass through unchanged.
func renumberCue(block string, cueNumber *int) string {
	lines := strings.Split(block, "\n")
	if _, err := strconv.Atoi(strings.TrimSpace(lines[0])); err != nil {
		return block
	}
	*cueNumber++
	lines[0] = strconv.Itoa(*cueNumber)
	return strings.Join(lines, "\n")
}

func cueTime(hours, minutes, seconds, milliseconds string) time.Duration {
	h, _ := strconv.Atoi(hours)
	m, _ := strconv.Atoi(minutes)
	s, _ := strconv.Atoi(seconds)
	ms, _ := strconv.Atoi(milliseconds)
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second + time.Duration(ms)*time.Millisecond
}

func formatCueTime(t time.Duration, separator string) string {
	return fmt.Sprintf("%02d:%02d:%02d%s%03d",
		int(t.Hours()), int(t.Minutes())%60, int(t.Seconds())%60, separator, t.Milliseconds()%1000)
}
