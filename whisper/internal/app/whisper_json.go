package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"
)

type subtitleCue struct {
	Start       time.Duration
	End         time.Duration
	Text        string
	Probability float64
}

type whisperOffsets struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

type whisperToken struct {
	Text        string         `json:"text"`
	Offsets     whisperOffsets `json:"offsets"`
	Probability float64        `json:"p"`
}

type whisperSegment struct {
	Text    string         `json:"text"`
	Offsets whisperOffsets `json:"offsets"`
	Tokens  []whisperToken `json:"tokens"`
}

type whisperPayload struct {
	Transcription []whisperSegment `json:"transcription"`
}

func parseWhisperJSON(content []byte, chunkStart time.Duration) ([]subtitleCue, error) {
	var payload whisperPayload
	if err := json.Unmarshal(content, &payload); err != nil {
		return nil, fmt.Errorf("decode whisper JSON: %w", err)
	}

	cues := make([]subtitleCue, 0, len(payload.Transcription))
	for _, segment := range payload.Transcription {
		text := strings.TrimSpace(segment.Text)
		if text == "" {
			continue
		}
		start := chunkStart + time.Duration(segment.Offsets.From)*time.Millisecond
		end := chunkStart + time.Duration(segment.Offsets.To)*time.Millisecond
		if end <= start {
			return nil, fmt.Errorf("invalid whisper segment timing %d --> %d", segment.Offsets.From, segment.Offsets.To)
		}
		cues = append(cues, subtitleCue{
			Start:       start,
			End:         end,
			Text:        text,
			Probability: averageTokenProbability(segment.Tokens),
		})
	}
	return cues, nil
}

func averageTokenProbability(tokens []whisperToken) float64 {
	total := 0.0
	count := 0
	for _, token := range tokens {
		if strings.HasPrefix(token.Text, "[_") || isPunctuationText(token.Text) {
			continue
		}
		total += token.Probability
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func isPunctuationText(text string) bool {
	hasRune := false
	for _, character := range text {
		if unicode.IsSpace(character) {
			continue
		}
		hasRune = true
		if !unicode.IsPunct(character) {
			return false
		}
	}
	return hasRune
}
