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
	Tokens      []subtitleToken
	Origin      transcriptionOrigin
}

type subtitleToken struct {
	Start       time.Duration
	End         time.Duration
	Text        string
	Probability float64
	Origin      transcriptionOrigin
}

type transcriptionOrigin struct {
	Index int
	Start time.Duration
	End   time.Duration
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
	return parseWhisperJSONForOrigin(content, transcriptionOrigin{Start: chunkStart})
}

func parseWhisperJSONForOrigin(content []byte, origin transcriptionOrigin) ([]subtitleCue, error) {
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
		start := origin.Start + time.Duration(segment.Offsets.From)*time.Millisecond
		end := origin.Start + time.Duration(segment.Offsets.To)*time.Millisecond
		if end <= start {
			return nil, fmt.Errorf("invalid whisper segment timing %d --> %d", segment.Offsets.From, segment.Offsets.To)
		}
		tokens, tokensValid := absoluteSubtitleTokens(segment.Tokens, segment.Offsets, origin)
		if !tokensValid {
			tokens = nil
		}
		cues = append(cues, subtitleCue{
			Start:       start,
			End:         end,
			Text:        text,
			Probability: averageTokenProbability(segment.Tokens, segment.Offsets),
			Tokens:      tokens,
			Origin:      origin,
		})
	}
	return cues, nil
}

func absoluteSubtitleTokens(tokens []whisperToken, segmentOffsets whisperOffsets, origin transcriptionOrigin) ([]subtitleToken, bool) {
	result := make([]subtitleToken, 0, len(tokens))
	for _, token := range tokens {
		if isWhisperControlToken(token.Text) || strings.TrimSpace(token.Text) == "" {
			continue
		}
		if !hasValidWhisperTokenTiming(token, segmentOffsets) {
			return nil, false
		}
		result = append(result, subtitleToken{
			Start:       origin.Start + time.Duration(token.Offsets.From)*time.Millisecond,
			End:         origin.Start + time.Duration(token.Offsets.To)*time.Millisecond,
			Text:        token.Text,
			Probability: token.Probability,
			Origin:      origin,
		})
	}
	return result, true
}

func averageTokenProbability(tokens []whisperToken, segmentOffsets whisperOffsets) float64 {
	total := 0.0
	count := 0
	for _, token := range tokens {
		if !hasValidWhisperTokenTiming(token, segmentOffsets) {
			continue
		}
		if isWhisperControlToken(token.Text) || isPunctuationText(token.Text) {
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

func hasValidWhisperTokenTiming(token whisperToken, segmentOffsets whisperOffsets) bool {
	return token.Offsets.From < token.Offsets.To &&
		token.Offsets.From >= segmentOffsets.From &&
		token.Offsets.To <= segmentOffsets.To
}

func isWhisperControlToken(text string) bool {
	return strings.HasPrefix(strings.TrimSpace(text), "[_")
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
