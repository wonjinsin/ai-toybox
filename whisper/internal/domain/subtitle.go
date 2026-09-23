package domain

import (
	"sort"
	"strings"
	"time"
	"unicode"
)

const minimumSubtitleProbability = 0.32

const (
	minimumSpecificHallucinationOccurrences = 2
	minimumGoodNightOccurrences             = 3
	minimumGenericThanksOccurrences         = 4
)

const (
	maximumReadableCueDuration = 5500 * time.Millisecond
	minimumReadableCueDuration = 650 * time.Millisecond
	shortCueDisplayDuration    = 1100 * time.Millisecond
	minimumCueGap              = 50 * time.Millisecond
)

func CleanSubtitleCues(cues []Cue, language string, corrections map[string]string, mediaDuration time.Duration) []Cue {
	ordered := append([]Cue(nil), cues...)
	sort.SliceStable(ordered, func(first, second int) bool {
		if ordered[first].Start == ordered[second].Start {
			return ordered[first].End < ordered[second].End
		}
		return ordered[first].Start < ordered[second].Start
	})
	ordered = ReconcileChunkBoundaries(ordered)

	cleaned := make([]Cue, 0, len(ordered))
	for _, cue := range ordered {
		cue.Text = strings.TrimSpace(cue.Text)
		if cue.Probability < minimumSubtitleProbability {
			continue
		}
		if language == "ja" && isJapaneseNonlexical(cue.Text) {
			continue
		}
		cleaned = append(cleaned, cue)
	}
	if language == "ja" || language == "auto" {
		cleaned = removeRepeatedJapaneseHallucinations(cleaned)
	}
	expanded := make([]Cue, 0, len(cleaned))
	for _, cue := range cleaned {
		appendSplitSubtitleCue(&expanded, cue, 2*SubtitleLineWidth(language), corrections)
	}
	for index, cue := range expanded {
		cue.Text = normalizeSubtitleText(cue.Text, language, corrections)
		expanded[index] = cue
	}
	cleaned = expanded
	sort.SliceStable(cleaned, func(first, second int) bool {
		if cleaned[first].Start == cleaned[second].Start {
			return cleaned[first].End < cleaned[second].End
		}
		return cleaned[first].Start < cleaned[second].Start
	})

	timed := make([]Cue, 0, len(cleaned))
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

func removeRepeatedJapaneseHallucinations(cues []Cue) []Cue {
	counts := make(map[string]int)
	for _, cue := range cues {
		if key, _ := japaneseHallucinationRule(cue.Text); key != "" {
			counts[key]++
		}
	}
	filtered := make([]Cue, 0, len(cues))
	for _, cue := range cues {
		key, minimumOccurrences := japaneseHallucinationRule(cue.Text)
		if key != "" && counts[key] >= minimumOccurrences {
			continue
		}
		filtered = append(filtered, cue)
	}
	return filtered
}

func japaneseHallucinationRule(text string) (string, int) {
	compact := strings.Map(func(character rune) rune {
		if unicode.IsSpace(character) {
			return -1
		}
		return character
	}, text)
	compact = strings.TrimRightFunc(compact, func(character rune) bool {
		return strings.ContainsRune("。．.!！?？", character)
	})
	switch compact {
	case "ご視聴ありがとうございました", "ご視聴ありがとうございます",
		"ご視聴いただきありがとうございました", "ご視聴いただきありがとうございます",
		"最後までご視聴ありがとうございました", "最後までご視聴ありがとうございます",
		"最後までご視聴いただきありがとうございました", "最後までご視聴いただきありがとうございます",
		"ご覧いただきありがとうございました", "ご覧いただきありがとうございます":
		return "viewing-thanks", minimumSpecificHallucinationOccurrences
	case "次の動画でお会いしましょう", "また次の動画でお会いしましょう",
		"次回の動画でお会いしましょう", "また次回の動画でお会いしましょう":
		return "next-video", minimumSpecificHallucinationOccurrences
	case "おやすみなさい", "お休みなさい":
		return "good-night", minimumGoodNightOccurrences
	case "ありがとうございました":
		return "thanks-past", minimumGenericThanksOccurrences
	case "ありがとうございます":
		return "thanks-present", minimumGenericThanksOccurrences
	case "おめでとうございます", "おめでとうございました":
		return "congratulations", minimumSpecificHallucinationOccurrences
	default:
		return "", 0
	}
}

func appendSplitSubtitleCue(destination *[]Cue, cue Cue, maximumCharacters int, corrections map[string]string) {
	characters := []rune(cue.Text)
	if len(characters) <= maximumCharacters {
		*destination = append(*destination, cue)
		return
	}

	if first, second, ok := splitSubtitleCueAtTokenBoundary(cue, corrections); ok {
		appendSplitSubtitleCue(destination, first, maximumCharacters, corrections)
		appendSplitSubtitleCue(destination, second, maximumCharacters, corrections)
		return
	}

	splitAt, ok := subtitleSplitPositionOutsideCorrections(cue.Text, characters, corrections)
	if !ok {
		*destination = append(*destination, cue)
		return
	}
	splitTime := cue.Start + time.Duration(float64(cue.End-cue.Start)*float64(splitAt)/float64(len(characters)))
	appendSplitSubtitleCue(destination, Cue{
		Start: cue.Start, End: splitTime, Text: string(characters[:splitAt]), Probability: cue.Probability, Origin: cue.Origin,
	}, maximumCharacters, corrections)
	appendSplitSubtitleCue(destination, Cue{
		Start: splitTime, End: cue.End, Text: string(characters[splitAt:]), Probability: cue.Probability, Origin: cue.Origin,
	}, maximumCharacters, corrections)
}

type subtitleTokenBoundary struct {
	TokenIndex int
	Position   int
	End        time.Duration
	NextStart  time.Duration
	Preferred  bool
}

func splitSubtitleCueAtTokenBoundary(cue Cue, corrections map[string]string) (Cue, Cue, bool) {
	boundary, ok := subtitleTokenSplitBoundary(cue, corrections)
	if !ok {
		return Cue{}, Cue{}, false
	}
	characters := []rune(cue.Text)
	first := Cue{
		Start:       cue.Start,
		End:         boundary.End,
		Text:        strings.TrimSpace(string(characters[:boundary.Position])),
		Probability: cue.Probability,
		Tokens:      append([]Token(nil), cue.Tokens[:boundary.TokenIndex+1]...),
		Origin:      cue.Origin,
	}
	second := Cue{
		Start:       boundary.NextStart,
		End:         cue.End,
		Text:        strings.TrimSpace(string(characters[boundary.Position:])),
		Probability: cue.Probability,
		Tokens:      append([]Token(nil), cue.Tokens[boundary.TokenIndex+1:]...),
		Origin:      cue.Origin,
	}
	if first.End <= first.Start || second.End <= second.Start || first.Text == "" || second.Text == "" {
		return Cue{}, Cue{}, false
	}
	return first, second, true
}

func subtitleTokenSplitBoundary(cue Cue, corrections map[string]string) (subtitleTokenBoundary, bool) {
	if len(cue.Tokens) < 2 {
		return subtitleTokenBoundary{}, false
	}

	characters := []rune(cue.Text)
	prefix := ""
	boundaries := make([]subtitleTokenBoundary, 0, len(cue.Tokens)-1)
	for index, token := range cue.Tokens {
		prefix += token.Text
		if index == len(cue.Tokens)-1 {
			break
		}
		next := cue.Tokens[index+1]
		position := len([]rune(strings.TrimSpace(prefix)))
		if position < 6 || len(characters)-position < 6 || token.End > next.Start || isProtectedSubtitleSplit(cue.Text, position, corrections) {
			continue
		}
		boundaries = append(boundaries, subtitleTokenBoundary{
			TokenIndex: index,
			Position:   position,
			End:        token.End,
			NextStart:  next.Start,
			Preferred:  tokenEndsWithPunctuation(token.Text) || tokenStartsWithSpace(next.Text),
		})
	}
	if strings.TrimSpace(prefix) != cue.Text || len(boundaries) == 0 {
		return subtitleTokenBoundary{}, false
	}

	preferred := make([]subtitleTokenBoundary, 0, len(boundaries))
	for _, boundary := range boundaries {
		if boundary.Preferred {
			preferred = append(preferred, boundary)
		}
	}
	if len(preferred) > 0 {
		boundaries = preferred
	}

	target := len(characters) / 2
	best := boundaries[0]
	bestDistance := absoluteDifference(best.Position, target)
	for _, boundary := range boundaries[1:] {
		distance := absoluteDifference(boundary.Position, target)
		if distance < bestDistance {
			best = boundary
			bestDistance = distance
		}
	}
	return best, true
}

func subtitleSplitPositionOutsideCorrections(text string, characters []rune, corrections map[string]string) (int, bool) {
	if len(corrections) == 0 {
		return subtitleSplitPosition(characters), true
	}
	target := len(characters) / 2
	candidates := make([]int, 0, len(characters)-1)
	punctuationCandidates := make([]int, 0, len(characters)-1)
	for index, character := range characters {
		candidate := index + 1
		if candidate < 6 || len(characters)-candidate < 6 || isProtectedSubtitleSplit(text, candidate, corrections) {
			continue
		}
		candidates = append(candidates, candidate)
		if strings.ContainsRune("、。！？,.!?", character) {
			punctuationCandidates = append(punctuationCandidates, candidate)
		}
	}
	if len(punctuationCandidates) > 0 {
		candidates = punctuationCandidates
	}
	if len(candidates) == 0 {
		return 0, false
	}
	best := candidates[0]
	bestDistance := absoluteDifference(best, target)
	for _, candidate := range candidates[1:] {
		distance := absoluteDifference(candidate, target)
		if distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}
	return best, true
}

func isProtectedSubtitleSplit(text string, position int, corrections map[string]string) bool {
	for source := range corrections {
		if source == "" {
			continue
		}
		for searchStart := 0; searchStart < len(text); {
			relativeStart := strings.Index(text[searchStart:], source)
			if relativeStart < 0 {
				break
			}
			start := searchStart + relativeStart
			end := start + len(source)
			startPosition := len([]rune(text[:start]))
			endPosition := startPosition + len([]rune(source))
			if position > startPosition && position < endPosition {
				return true
			}
			searchStart = end
		}
	}
	return false
}

func absoluteDifference(first, second int) int {
	if first < second {
		return second - first
	}
	return first - second
}

func tokenEndsWithPunctuation(text string) bool {
	for index := len([]rune(text)) - 1; index >= 0; index-- {
		character := []rune(text)[index]
		if unicode.IsSpace(character) {
			continue
		}
		return unicode.IsPunct(character)
	}
	return false
}

func tokenStartsWithSpace(text string) bool {
	for _, character := range text {
		return unicode.IsSpace(character)
	}
	return false
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

func SubtitleLineWidth(language string) int {
	switch language {
	case "ja", "zh", "auto":
		return 18
	case "ko":
		return 22
	default:
		return 42
	}
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
