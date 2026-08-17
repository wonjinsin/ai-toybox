package app

import (
	"strings"
	"time"
	"unicode"
)

func reconcileChunkBoundaries(cues []subtitleCue) []subtitleCue {
	reconciled := make([]subtitleCue, 0, len(cues))
	for _, cue := range cues {
		matching := -1
		merged := false
		for index, prior := range reconciled {
			if !isOverlappingAdjacentChunkCue(prior, cue) {
				continue
			}
			if matchingBoundaryText(prior.Text, cue.Text) {
				matching = index
				break
			}
			if combined, ok := mergeSharedBoundaryTokens(prior, cue); ok {
				reconciled[index] = combined
				matching = index
				merged = true
				break
			}
		}
		if matching < 0 {
			reconciled = append(reconciled, cue)
			continue
		}
		if merged {
			continue
		}
		if isMoreCompleteBoundaryCue(cue, reconciled[matching]) ||
			(normalizedBoundaryText(cue.Text) == normalizedBoundaryText(reconciled[matching].Text) && isHigherQualityBoundaryCue(cue, reconciled[matching])) {
			reconciled[matching] = cue
		}
	}
	return reconciled
}

type boundaryLexicalToken struct {
	Index int
	Text  string
	Token subtitleToken
}

func mergeSharedBoundaryTokens(first, second subtitleCue) (subtitleCue, bool) {
	overlapStart := max(first.Origin.Start, second.Origin.Start)
	overlapEnd := min(first.Origin.End, second.Origin.End)
	firstTokens := boundaryLexicalTokens(first.Tokens, overlapStart, overlapEnd)
	secondTokens := boundaryLexicalTokens(second.Tokens, overlapStart, overlapEnd)
	maximum := min(len(firstTokens), len(secondTokens))
	for count := maximum; count > 0; count-- {
		if !matchingBoundaryTokenSuffixPrefix(firstTokens, secondTokens, count) {
			continue
		}
		rightStart := secondTokens[count-1].Index + 1
		mergedTokens := append([]subtitleToken(nil), first.Tokens...)
		mergedTokens = append(mergedTokens, second.Tokens[rightStart:]...)
		text := strings.TrimSpace(joinSubtitleTokenText(mergedTokens))
		if text == "" {
			return subtitleCue{}, false
		}
		return subtitleCue{
			Start:       min(first.Start, second.Start),
			End:         max(first.End, second.End),
			Text:        text,
			Probability: averageSubtitleTokenProbability(mergedTokens, max(first.Probability, second.Probability)),
			Tokens:      mergedTokens,
			Origin:      first.Origin,
		}, true
	}
	return subtitleCue{}, false
}

func boundaryLexicalTokens(tokens []subtitleToken, overlapStart, overlapEnd time.Duration) []boundaryLexicalToken {
	result := make([]boundaryLexicalToken, 0, len(tokens))
	for index, token := range tokens {
		text := normalizedBoundaryText(token.Text)
		if text == "" || token.End <= overlapStart || token.Start >= overlapEnd {
			continue
		}
		result = append(result, boundaryLexicalToken{Index: index, Text: text, Token: token})
	}
	return result
}

func matchingBoundaryTokenSuffixPrefix(first, second []boundaryLexicalToken, count int) bool {
	for index := 0; index < count; index++ {
		if first[len(first)-count+index].Text != second[index].Text {
			return false
		}
	}
	return true
}

func joinSubtitleTokenText(tokens []subtitleToken) string {
	var text strings.Builder
	for _, token := range tokens {
		text.WriteString(token.Text)
	}
	return text.String()
}

func averageSubtitleTokenProbability(tokens []subtitleToken, fallback float64) float64 {
	total := 0.0
	count := 0
	for _, token := range tokens {
		if normalizedBoundaryText(token.Text) == "" {
			continue
		}
		total += token.Probability
		count++
	}
	if count == 0 {
		return fallback
	}
	return total / float64(count)
}

func matchingBoundaryText(first, second string) bool {
	first = normalizedBoundaryText(first)
	second = normalizedBoundaryText(second)
	return first != "" && second != "" && (first == second || strings.Contains(first, second) || strings.Contains(second, first))
}

func isMoreCompleteBoundaryCue(candidate, current subtitleCue) bool {
	return len([]rune(normalizedBoundaryText(candidate.Text))) > len([]rune(normalizedBoundaryText(current.Text)))
}

func isOverlappingAdjacentChunkCue(first, second subtitleCue) bool {
	if absoluteDifference(first.Origin.Index, second.Origin.Index) != 1 {
		return false
	}
	overlapStart := max(first.Origin.Start, second.Origin.Start)
	overlapEnd := min(first.Origin.End, second.Origin.End)
	if overlapEnd <= overlapStart {
		return false
	}
	return cueIntersectsInterval(first, overlapStart, overlapEnd) && cueIntersectsInterval(second, overlapStart, overlapEnd)
}

func cueIntersectsInterval(cue subtitleCue, start, end time.Duration) bool {
	return cue.Start < end && cue.End > start
}

func normalizedBoundaryText(text string) string {
	var normalized strings.Builder
	for _, character := range strings.ToLower(text) {
		if unicode.IsSpace(character) || unicode.IsPunct(character) {
			continue
		}
		normalized.WriteRune(character)
	}
	return normalized.String()
}

func isHigherQualityBoundaryCue(candidate, current subtitleCue) bool {
	if candidate.Probability != current.Probability {
		return candidate.Probability > current.Probability
	}
	return len([]rune(candidate.Text)) > len([]rune(current.Text))
}
