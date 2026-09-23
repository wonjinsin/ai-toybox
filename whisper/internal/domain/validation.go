package domain

import (
	"fmt"
	"strings"
	"time"
)

func ValidateSubtitleCues(cues []Cue, mediaDuration time.Duration) error {
	for index, cue := range cues {
		if strings.TrimSpace(cue.Text) == "" {
			return fmt.Errorf("subtitle cue %d has empty text", index+1)
		}
		if cue.Start < 0 || cue.End <= cue.Start {
			return fmt.Errorf("subtitle cue %d has invalid timing %s --> %s", index+1, cue.Start, cue.End)
		}
		if mediaDuration > 0 && cue.End > mediaDuration {
			return fmt.Errorf("subtitle cue %d exceeds media duration", index+1)
		}
		if index > 0 && cue.Start < cues[index-1].End {
			return fmt.Errorf("subtitle cues %d and %d overlap", index, index+1)
		}
	}
	return nil
}
