package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wonjinsin/ledger/internal/core/port/out"
)

// runJSON runs the prompt and unmarshals the AI's reply into v.
// AI CLIs often wrap JSON in code fences or prose; extractJSON tolerates that.
// On parse failure it re-asks once with an explicit correction, then errors.
func runJSON(ctx context.Context, runner out.AIRunner, prompt string, v any) error {
	reply, err := runner.Run(ctx, prompt)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(extractJSON(reply)), v); err == nil {
		return nil
	}

	retryPrompt := prompt + "\n\nYour previous reply was not valid JSON:\n" + reply +
		"\n\nReply again with ONLY valid JSON matching the requested schema. No prose, no code fences."
	reply, err = runner.Run(ctx, retryPrompt)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(extractJSON(reply)), v); err != nil {
		return fmt.Errorf("AI reply is not valid JSON after retry: %w", err)
	}
	return nil
}

// extractJSON returns the outermost JSON object or array embedded in s.
func extractJSON(s string) string {
	start := len(s)
	end := -1
	for _, pair := range [2][2]string{{"{", "}"}, {"[", "]"}} {
		if i := strings.Index(s, pair[0]); i != -1 && i < start {
			if j := strings.LastIndex(s, pair[1]); j > i {
				start, end = i, j
			}
		}
	}
	if end == -1 {
		return s
	}
	return s[start : end+1]
}
