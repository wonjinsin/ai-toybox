package domain

import (
	"strings"
	"testing"
	"time"
)

func TestValidateSubtitleCuesRejectsMediaOverflow(t *testing.T) {
	t.Parallel()

	err := ValidateSubtitleCues([]Cue{{
		Start: 9 * time.Second,
		End:   11 * time.Second,
		Text:  "overflow",
	}}, 10*time.Second)
	if err == nil || !strings.Contains(err.Error(), "exceeds media duration") {
		t.Fatalf("ValidateSubtitleCues() error = %v, want media duration error", err)
	}
}
