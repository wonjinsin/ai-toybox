package app

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Sparse VAD segments make whisper.cpp stretch a cue's end time to the next
// segment, producing subtitles that stay on screen for minutes. Real speech
// segments are capped at 30s by -vmsd, so anything longer is a timing artifact.
const maxCueDuration = 15 * time.Second

// Hallucination on non-speech audio repeats the same text across consecutive
// cues; real dialogue rarely repeats an identical line this many times in a row.
const repeatedCueRunLimit = 3

const minimumSubtitleProbability = 0.32

const (
	maximumReadableCueDuration = 5500 * time.Millisecond
	minimumReadableCueDuration = 650 * time.Millisecond
	shortCueDisplayDuration    = 1100 * time.Millisecond
	minimumCueGap              = 50 * time.Millisecond
)

const (
	minimumVADGapCorrection = time.Second
	vadBoundaryTolerance    = 100 * time.Millisecond
)

var cueTimingPattern = regexp.MustCompile(`^(\d{2,}):(\d{2}):(\d{2})([,.])(\d{3}) --> (\d{2,}):(\d{2}):(\d{2})[,.](\d{3})$`)

type speechSegment struct {
	Start time.Duration
	End   time.Duration
}

func cleanSubtitleCues(cues []subtitleCue, language string, corrections map[string]string, mediaDuration time.Duration) []subtitleCue {
	ordered := append([]subtitleCue(nil), cues...)
	sort.SliceStable(ordered, func(first, second int) bool {
		if ordered[first].Start == ordered[second].Start {
			return ordered[first].End < ordered[second].End
		}
		return ordered[first].Start < ordered[second].Start
	})

	cleaned := make([]subtitleCue, 0, len(ordered))
	for _, cue := range ordered {
		cue.Text = normalizeSubtitleText(cue.Text, language, corrections)
		if cue.Probability < minimumSubtitleProbability {
			continue
		}
		if language == "ja" && isJapaneseNonlexical(cue.Text) {
			continue
		}
		duplicate := false
		for index := len(cleaned) - 1; index >= 0 && index >= len(cleaned)-8; index-- {
			prior := cleaned[index]
			if cue.Text == prior.Text && cue.Start-prior.Start < 10*time.Second {
				duplicate = true
				break
			}
		}
		if duplicate {
			continue
		}
		cleaned = append(cleaned, cue)
	}
	expanded := make([]subtitleCue, 0, len(cleaned))
	for _, cue := range cleaned {
		appendSplitSubtitleCue(&expanded, cue, 2*subtitleLineWidth(language))
	}
	cleaned = expanded

	timed := make([]subtitleCue, 0, len(cleaned))
	for index, cue := range cleaned {
		if cue.Start < 0 || (mediaDuration > 0 && cue.Start >= mediaDuration) {
			continue
		}
		nextStart := time.Duration(1<<63 - 1)
		if index+1 < len(cleaned) {
			nextStart = cleaned[index+1].Start
		}
		end := min(cue.End, cue.Start+maximumReadableCueDuration)
		if mediaDuration > 0 {
			end = min(end, mediaDuration)
		}
		end = min(end, nextStart-minimumCueGap)
		if end-cue.Start < minimumReadableCueDuration {
			end = cue.Start + shortCueDisplayDuration
			if mediaDuration > 0 {
				end = min(end, mediaDuration)
			}
			end = min(end, nextStart-minimumCueGap)
		}
		if end <= cue.Start {
			continue
		}
		cue.End = end
		timed = append(timed, cue)
	}
	return timed
}

func appendSplitSubtitleCue(destination *[]subtitleCue, cue subtitleCue, maximumCharacters int) {
	characters := []rune(cue.Text)
	if len(characters) <= maximumCharacters {
		*destination = append(*destination, cue)
		return
	}

	splitAt := subtitleSplitPosition(characters)
	splitTime := cue.Start + time.Duration(float64(cue.End-cue.Start)*float64(splitAt)/float64(len(characters)))
	appendSplitSubtitleCue(destination, subtitleCue{
		Start: cue.Start, End: splitTime, Text: string(characters[:splitAt]), Probability: cue.Probability,
	}, maximumCharacters)
	appendSplitSubtitleCue(destination, subtitleCue{
		Start: splitTime, End: cue.End, Text: string(characters[splitAt:]), Probability: cue.Probability,
	}, maximumCharacters)
}

func subtitleSplitPosition(characters []rune) int {
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
		if distance < bestDistance {
			splitAt = candidate
			bestDistance = distance
		}
	}
	return splitAt
}

func normalizeSubtitleText(text, language string, corrections map[string]string) string {
	text = strings.TrimSpace(text)
	keys := make([]string, 0, len(corrections))
	for source := range corrections {
		if source != "" {
			keys = append(keys, source)
		}
	}
	sort.Slice(keys, func(first, second int) bool {
		if len(keys[first]) == len(keys[second]) {
			return keys[first] < keys[second]
		}
		return len(keys[first]) > len(keys[second])
	})
	for _, source := range keys {
		text = strings.ReplaceAll(text, source, corrections[source])
	}
	if language == "ja" || language == "zh" {
		text = strings.ReplaceAll(text, "?", "？")
		text = strings.ReplaceAll(text, "!", "！")
	}
	return text
}

func renderTranscript(cues []subtitleCue, format, language string) (string, error) {
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
		text, err := wrapSubtitleText(cue.Text, subtitleLineWidth(language))
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

func subtitleLineWidth(language string) int {
	switch language {
	case "ja", "zh", "auto":
		return 18
	case "ko":
		return 22
	default:
		return 42
	}
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

func isJapaneseNonlexical(text string) bool {
	normalized := strings.TrimSpace(text)
	switch normalized {
	case "はい", "うん", "え", "え?", "え？", "えっ", "えっ?", "えっ？", "（笑）":
		return false
	}

	vocalizationCharacters := "あいうえおぁぃぅぇぉんっはふへほわアイウエオァィゥェォンッハフヘホワ"
	compact := make([]rune, 0, len([]rune(normalized)))
	for _, character := range normalized {
		if unicode.IsSpace(character) || unicode.IsPunct(character) || strings.ContainsRune("ー〜～…‥", character) {
			continue
		}
		compact = append(compact, character)
	}
	if len(compact) == 0 {
		return true
	}
	for _, character := range compact {
		if !strings.ContainsRune(vocalizationCharacters, character) {
			return false
		}
	}
	return true
}

func postProcessTranscript(path string, vadSegments []speechSegment) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read transcript %q: %w", path, err)
	}
	processed := collapseRepeatedCues(string(content))
	processed = alignCueStartsToSpeech(processed, vadSegments)
	processed = clampCueDurations(processed)
	return os.WriteFile(path, []byte(processed), 0o644)
}

func alignCueStartsToSpeech(content string, vadSegments []speechSegment) string {
	lines := strings.Split(content, "\n")
	for index, line := range lines {
		match := cueTimingPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		start := cueTime(match[1], match[2], match[3], match[5])
		end := cueTime(match[6], match[7], match[8], match[9])
		overlappingSegments := make([]speechSegment, 0, 1)
		for _, segment := range vadSegments {
			if segment.End > start+vadBoundaryTolerance && segment.Start < end {
				overlappingSegments = append(overlappingSegments, segment)
			}
		}
		if len(overlappingSegments) != 1 {
			continue
		}
		segment := overlappingSegments[0]
		startsInSegment := start >= segment.Start && start < segment.End
		if startsInSegment || segment.Start-start < minimumVADGapCorrection {
			continue
		}
		lines[index] = fmt.Sprintf("%s --> %s", formatCueTime(segment.Start, match[4]), formatCueTime(end, match[4]))
	}
	return strings.Join(lines, "\n")
}

func containsCueTiming(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if cueTimingPattern.MatchString(line) {
			return true
		}
	}
	return false
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
