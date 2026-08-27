package app

import "time"

const (
	maxAudioChunkDuration = 20 * time.Second
	chunkBoundaryPadding  = 150 * time.Millisecond
	maxChunkGap           = 3 * time.Second
)

type audioChunk struct {
	Start time.Duration
	End   time.Duration
}

func buildSpeechChunks(segments []speechSegment, mediaDuration time.Duration) []audioChunk {
	maxSpeechDuration := maxAudioChunkDuration - 2*chunkBoundaryPadding
	speechParts := make([]speechSegment, 0, len(segments))
	for _, segment := range segments {
		start := max(segment.Start, time.Duration(0))
		end := segment.End
		if mediaDuration > 0 {
			end = min(end, mediaDuration)
		}
		for start < end {
			partEnd := min(start+maxSpeechDuration, end)
			speechParts = append(speechParts, speechSegment{Start: start, End: partEnd})
			start = partEnd
		}
	}

	groups := make([]speechSegment, 0, len(speechParts))
	var current speechSegment
	hasCurrent := false
	for _, part := range speechParts {
		if !hasCurrent {
			current = part
			hasCurrent = true
			continue
		}
		gap := part.Start - current.End
		span := part.End - current.Start
		if gap <= maxChunkGap && span <= maxSpeechDuration {
			current = speechSegment{Start: current.Start, End: max(current.End, part.End)}
			continue
		}
		groups = append(groups, current)
		current = part
	}
	if hasCurrent {
		groups = append(groups, current)
	}

	chunks := make([]audioChunk, 0, len(groups))
	for _, group := range groups {
		start := max(time.Duration(0), group.Start-chunkBoundaryPadding)
		end := group.End + chunkBoundaryPadding
		if mediaDuration > 0 {
			end = min(end, mediaDuration)
		}
		chunks = append(chunks, audioChunk{Start: start, End: end})
	}
	return chunks
}
