// Package evidence summarizes recorded events without changing their source.
package evidence

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Summarize extracts supported observations from a recorded JSONL stream.
func Summarize(raw []byte) map[string]any {
	lines := bytes.Split(raw, []byte{'\n'})
	if len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	type completion struct {
		line  int
		event map[string]any
	}
	completions := []completion{}
	diagnostics := []map[string]any{}
	failures := []map[string]any{}
	for index, line := range lines {
		event, diagnostic := parseLine(index+1, line)
		if diagnostic != nil {
			diagnostics = append(diagnostics, diagnostic)
			continue
		}
		if event["type"] == "turn.completed" {
			completions = append(completions, completion{index + 1, event})
		}
		failures = append(failures, failureDiagnostics(index+1, event)...)
	}
	usage := unknownUsage("no_completion_event")
	if len(diagnostics) > 0 {
		usage = unknownUsage("invalid_event_stream")
	} else if len(completions) == 1 {
		usage, diagnostics = completionUsage(completions[0].line, completions[0].event)
	} else if len(completions) > 1 {
		usage = unknownUsage("multiple_completion_events")
		diagnostics = append(diagnostics, map[string]any{
			"line": nil, "kind": "ambiguous_usage_scope", "message": "Multiple turn.completed events; token counts are not aggregated",
		})
	}
	return map[string]any{
		"usage":       usage,
		"reported":    map[string]any{"model": nil, "model_version": nil, "effort": nil, "evidence_source": nil},
		"diagnostics": append(diagnostics, failures...),
	}
}

func failureDiagnostics(line int, event map[string]any) []map[string]any {
	kind := event["type"]
	if kind != "error" && kind != "turn.failed" && kind != "item.failed" {
		return nil
	}
	containers := []map[string]any{event}
	if item, ok := event["item"].(map[string]any); ok {
		containers = append(containers, item)
	}
	diagnostics := []map[string]any{}
	for _, container := range containers {
		errorValue := container["error"]
		if object, ok := errorValue.(map[string]any); ok {
			errorValue = object["message"]
		}
		for _, value := range []any{container["message"], errorValue} {
			if message, ok := value.(string); ok {
				diagnostics = append(diagnostics, map[string]any{"line": line, "kind": kind, "message": message})
			}
		}
	}
	return diagnostics
}

func unknownUsage(reason string) map[string]any {
	return map[string]any{
		"capture_status": "unavailable",
		"input_tokens":   nil, "cached_input_tokens": nil, "output_tokens": nil,
		"reasoning_output_tokens": nil, "scope": nil, "evidence_source": nil, "reason": reason,
	}
}

func completionUsage(line int, event map[string]any) (map[string]any, []map[string]any) {
	value, present := event["usage"]
	values, valid := value.(map[string]any)
	if !valid {
		reason := "invalid_usage"
		if !present {
			reason = "missing_usage"
		}
		return unknownUsage(reason), []map[string]any{{
			"line": line, "kind": "invalid_usage", "message": "turn.completed usage must be an object",
		}}
	}
	usage := map[string]any{
		"capture_status":          "captured",
		"reasoning_output_tokens": nil,
		"scope":                   "single_turn_completed",
		"evidence_source":         fmt.Sprintf("events.jsonl:%d", line),
		"reason":                  nil,
	}
	diagnostics := []map[string]any{}
	available := 0
	for _, key := range []string{"input_tokens", "cached_input_tokens", "output_tokens"} {
		value, present := values[key]
		usage[key] = integerCount(value)
		if usage[key] != nil {
			available++
		} else if present {
			diagnostics = append(diagnostics, map[string]any{
				"line": line, "kind": "invalid_usage", "message": fmt.Sprintf("usage.%s must be a nonnegative integer", key),
			})
		}
	}
	if available != 3 {
		usage["capture_status"] = "partial"
		usage["reason"] = "incomplete_usage"
	}
	if available == 0 {
		usage["capture_status"] = "unavailable"
	}
	return usage, diagnostics
}

func integerCount(value any) any {
	number, ok := value.(json.Number)
	if !ok {
		return nil
	}
	if number == "-0" {
		return json.Number("0")
	}
	for _, character := range number.String() {
		if character < '0' || character > '9' {
			return nil
		}
	}
	return number
}
