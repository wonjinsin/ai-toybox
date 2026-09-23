package domain

import (
	"fmt"
	"strings"
	"time"
)

const (
	RetrySubtitleProbability   = 0.50
	retryContextDuration       = time.Second
	minimumRetryConfidenceGain = 0.05
)

func CountLowConfidenceCues(cues []Cue) int {
	count := 0
	for _, cue := range cues {
		if cue.Probability < RetrySubtitleProbability {
			count++
		}
	}
	return count
}

func RetryWindowForCue(cue Cue, mediaDuration time.Duration) (Chunk, error) {
	start := max(time.Duration(0), cue.Start-retryContextDuration)
	end := cue.End + retryContextDuration
	if mediaDuration > 0 {
		end = min(end, mediaDuration)
	}
	if end <= start {
		return Chunk{}, fmt.Errorf("invalid retry window %s --> %s", start, end)
	}
	return Chunk{Start: start, End: end}, nil
}

func RetryCandidateForCue(original Cue, retryCues []Cue) (Cue, error) {
	tokens := make([]Token, 0)
	for _, retryCue := range retryCues {
		if len(retryCue.Tokens) == 0 {
			continue
		}
		for _, token := range retryCue.Tokens {
			midpoint := token.Start + (token.End-token.Start)/2
			if midpoint < original.Start || midpoint >= original.End {
				continue
			}
			token.Origin = original.Origin
			tokens = append(tokens, token)
		}
	}
	if len(tokens) == 0 {
		return original, nil
	}
	text := strings.TrimSpace(joinSubtitleTokenText(tokens))
	if text == "" {
		return original, nil
	}
	return Cue{
		Start:       tokens[0].Start,
		End:         tokens[len(tokens)-1].End,
		Text:        text,
		Probability: averageSubtitleTokenProbability(tokens, original.Probability),
		Tokens:      tokens,
		Origin:      original.Origin,
	}, nil
}

func SelectRetryCue(original, candidate Cue) Cue {
	if candidate.Probability < minimumSubtitleProbability {
		return original
	}
	if original.Probability < minimumSubtitleProbability || candidate.Probability >= original.Probability+minimumRetryConfidenceGain {
		return candidate
	}
	return original
}
