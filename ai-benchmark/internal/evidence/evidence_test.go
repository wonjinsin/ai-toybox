package evidence_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"ai-benchmark/internal/evidence"
)

func TestSingleCompletionPreservesIntegerUsage(t *testing.T) {
	raw := []byte(`{"type":"turn.completed","usage":{"input_tokens":9007199254740993,"cached_input_tokens":40,"output_tokens":12}}` + "\n")
	want := map[string]any{
		"usage": map[string]any{
			"capture_status":          "captured",
			"input_tokens":            json.Number("9007199254740993"),
			"cached_input_tokens":     json.Number("40"),
			"output_tokens":           json.Number("12"),
			"reasoning_output_tokens": nil,
			"scope":                   "single_turn_completed",
			"evidence_source":         "events.jsonl:1",
			"reason":                  nil,
		},
		"reported":    map[string]any{"model": nil, "model_version": nil, "effort": nil, "evidence_source": nil},
		"diagnostics": []map[string]any{},
	}
	if got := evidence.Summarize(raw); !reflect.DeepEqual(got, want) {
		t.Fatalf("Summarize() = %#v, want %#v", got, want)
	}
	encoded, err := json.Marshal(evidence.Summarize(raw))
	if err != nil || !bytes.Contains(encoded, []byte(`"input_tokens":9007199254740993`)) {
		t.Fatalf("Integer precision changed during JSON serialization: %s, %v", encoded, err)
	}
}

func TestUsageRetainsOnlyNonnegativeIntegerCategories(t *testing.T) {
	for _, value := range []string{"true", "false", "-1", "1.0", "1e1", "1.5", `"4"`, "null", "[]", "{}"} {
		t.Run(value, func(t *testing.T) {
			raw := []byte(fmt.Sprintf(`{"type":"turn.completed","usage":{"input_tokens":%s,"output_tokens":3}}`, value))
			result := evidence.Summarize(raw)
			usage := result["usage"].(map[string]any)
			if usage["input_tokens"] != nil || usage["cached_input_tokens"] != nil || usage["output_tokens"] != json.Number("3") {
				t.Fatalf("Unexpected retained counts: %#v", usage)
			}
			if usage["capture_status"] != "partial" || usage["reason"] != "incomplete_usage" {
				t.Fatalf("Incomplete usage not labeled: %#v", usage)
			}
			diagnostics := result["diagnostics"].([]map[string]any)
			if len(diagnostics) != 1 || diagnostics[0]["kind"] != "invalid_usage" {
				t.Fatalf("Invalid category missing diagnostic: %#v", diagnostics)
			}
		})
	}
	for _, fields := range []string{`"input_tokens":0`, `"input_tokens":-0`, `"input_tokens":184467440737095516160000`} {
		usage := evidence.Summarize([]byte(`{"type":"turn.completed","usage":{` + fields + `}}`))["usage"].(map[string]any)
		if usage["input_tokens"] == nil || usage["capture_status"] != "partial" || usage["cached_input_tokens"] != nil {
			t.Fatalf("Missing categories must stay unknown: %#v", usage)
		}
	}
}

func TestUnusableUsageRemainsUnavailable(t *testing.T) {
	for _, fields := range []string{"", `,"usage":null`, `,"usage":[]`, `,"usage":"unknown"`, `,"usage":{}`, `,"usage":{"input_tokens":-1}`} {
		t.Run(fields, func(t *testing.T) {
			result := evidence.Summarize([]byte(`{"type":"turn.completed"` + fields + `}`))
			usage := result["usage"].(map[string]any)
			if usage["capture_status"] != "unavailable" || usage["input_tokens"] != nil {
				t.Fatalf("Unusable usage must remain unavailable: %#v", usage)
			}
			if fields == "" && usage["reason"] != "missing_usage" {
				t.Fatalf("Missing usage reason = %v", usage["reason"])
			}
			if fields == `,"usage":null` && usage["reason"] != "invalid_usage" {
				t.Fatalf("Invalid usage reason = %v", usage["reason"])
			}
		})
	}
}

func TestUsageRequiresExactlyOneCompletionInTheStream(t *testing.T) {
	completion := `{"type":"turn.completed","usage":{"input_tokens":7,"cached_input_tokens":2,"output_tokens":4}}`
	cases := []struct {
		name   string
		raw    string
		reason any
		source any
	}{
		{"empty", "", "no_completion_event", nil},
		{"unrelated usage", `{"type":"item.completed","usage":{"input_tokens":999}}`, "no_completion_event", nil},
		{"single completion", "{\"type\":\"thread.started\"}\n{\"type\":\"item.completed\"}\n" + completion, nil, "events.jsonl:3"},
		{"multiple completions", completion + "\n" + completion + "\n", "multiple_completion_events", nil},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result := evidence.Summarize([]byte(test.raw))
			usage := result["usage"].(map[string]any)
			if usage["reason"] != test.reason || usage["evidence_source"] != test.source {
				t.Fatalf("Unexpected usage scope: %#v", usage)
			}
			if test.reason != nil && (usage["input_tokens"] != nil || usage["scope"] != nil) {
				t.Fatalf("Unknown stream inferred usage: %#v", usage)
			}
			if test.reason == "multiple_completion_events" {
				diagnostics := result["diagnostics"].([]map[string]any)
				if len(diagnostics) != 1 || diagnostics[0]["kind"] != "ambiguous_usage_scope" {
					t.Fatalf("Missing ambiguity diagnostic: %#v", diagnostics)
				}
			}
		})
	}
}

func TestMalformedRecordsInvalidateUsageWithoutChangingRawBytes(t *testing.T) {
	completion := []byte("{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":1}}\n")
	cases := []struct{ name, broken, kind string }{
		{"truncated", `{"type":`, "invalid_json"},
		{"text", "not json\n", "invalid_json"},
		{"blank", "\n", "invalid_json"},
		{"multiple values", `{"type":"error"} {"type":"error"}`, "invalid_json"},
		{"array", "[]\n", "invalid_event"},
		{"null", "null\n", "invalid_event"},
		{"missing type", "{}\n", "invalid_event"},
		{"nonstring type", `{"type":3}`, "invalid_event"},
		{"duplicate type", `{"type":"turn.completed","type":"error"}`, "invalid_json"},
		{"duplicate nested key", `{"type":"turn.completed","usage":{"input_tokens":1,"input_tokens":2}}`, "invalid_json"},
		{"duplicate escaped key", `{"type":"error","message":"x","\u006dessage":"y"}`, "invalid_json"},
		{"nan", `{"type":"turn.completed","usage":{"input_tokens":NaN}}`, "invalid_json"},
		{"infinity", `{"type":"turn.completed","usage":{"input_tokens":Infinity}}`, "invalid_json"},
		{"invalid object key", `{"type":"error",1:"value"}`, "invalid_json"},
		{"truncated array", `{"type":"error","unknown":[`, "invalid_json"},
		{"invalid unicode escape", `{"type":"error","message":"\uZZZZ"}`, "invalid_json"},
		{"invalid utf8", "{\"type\":\"error\",\"message\":\"\xff\"}", "invalid_utf8"},
		{"excessive nesting", `{"type":"item.started","data":` + strings.Repeat("[", 2000) + "0" + strings.Repeat("]", 2000) + "}", "invalid_json"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			raw := append(bytes.Clone(completion), []byte(test.broken)...)
			original := bytes.Clone(raw)
			result := evidence.Summarize(raw)
			usage := result["usage"].(map[string]any)
			if usage["reason"] != "invalid_event_stream" || usage["input_tokens"] != nil {
				t.Fatalf("Malformed evidence retained usage: %#v", usage)
			}
			diagnostics := result["diagnostics"].([]map[string]any)
			if len(diagnostics) == 0 || diagnostics[0]["line"] != 2 || diagnostics[0]["kind"] != test.kind {
				t.Fatalf("Wrong malformed-line diagnostic: %#v", diagnostics)
			}
			if !bytes.Equal(raw, original) {
				t.Fatal("Summarize changed raw input")
			}
		})
	}
}

func TestFailureMessagesSurviveOtherMalformedLines(t *testing.T) {
	raw := []byte("broken\n" + strings.Join([]string{
		`{"type":"error","message":"request failed: \ud55c\uae00"}`,
		`{"type":"turn.failed","error":{"message":"provider unavailable"}}`,
		`{"type":"item.failed","item":{"error":{"message":"command failed\nexit 2"}}}`,
		`{"type":"error","error":"direct error"}`,
		`{"type":"item.failed","message":7,"item":{"message":"nested message","error":false}}`,
		`{"type":"error","item":null,"error":12}`,
	}, "\n"))
	result := evidence.Summarize(raw)
	diagnostics := result["diagnostics"].([]map[string]any)
	want := []map[string]any{
		{"line": 2, "kind": "error", "message": "request failed: 한글"},
		{"line": 3, "kind": "turn.failed", "message": "provider unavailable"},
		{"line": 4, "kind": "item.failed", "message": "command failed\nexit 2"},
		{"line": 5, "kind": "error", "message": "direct error"},
		{"line": 6, "kind": "item.failed", "message": "nested message"},
	}
	if len(diagnostics) < 1 || !reflect.DeepEqual(diagnostics[1:], want) {
		t.Fatalf("Failure diagnostics = %#v, want parsing diagnostic then %#v", diagnostics, want)
	}
}

func TestUnpairedSurrogateEscapesAreDiagnosedWithoutReplacement(t *testing.T) {
	for _, rawText := range []string{
		`{"type":"error","message":"\ud800"}`,
		`{"type":"error","message":"\udfff"}`,
		`{"type":"error","message":"\ud800x"}`,
		`{"type":"error","message":"\ud800\u0041"}`,
		`{"type":"error","message":"\ud800\ud800"}`,
		`{"type":"error","\ud800":"bad key"}`,
	} {
		t.Run(rawText, func(t *testing.T) {
			raw := []byte(rawText)
			result := evidence.Summarize(raw)
			usage := result["usage"].(map[string]any)
			diagnostics := result["diagnostics"].([]map[string]any)
			if usage["reason"] != "invalid_event_stream" || len(diagnostics) != 1 || diagnostics[0]["kind"] != "invalid_json" {
				t.Fatalf("Unpaired surrogate silently decoded: %#v", result)
			}
			if string(raw) != rawText {
				t.Fatal("Surrogate diagnostic changed raw bytes")
			}
		})
	}
	valid := []byte(`{"type":"error","message":"\ud83d\ude80 \ufffd \\ud800 \"quote\""}`)
	diagnostics := evidence.Summarize(valid)["diagnostics"].([]map[string]any)
	if len(diagnostics) != 1 || diagnostics[0]["message"] != "🚀 � \\ud800 \"quote\"" {
		t.Fatalf("Valid Unicode/escaped text changed: %#v", diagnostics)
	}
}

func TestUnknownFieldsDoNotBecomeVerifiedObservations(t *testing.T) {
	raw := []byte(" \t" + `{"type":"turn.completed","model":"claimed-model","effort":"high","usage":{"input_tokens":1,"cached_input_tokens":0,"output_tokens":2,"reasoning_output_tokens":99},"unknown":[null,1.5,true,{"nested":"ignored"}]}` + "\r\n")
	result := evidence.Summarize(raw)
	wantReported := map[string]any{"model": nil, "model_version": nil, "effort": nil, "evidence_source": nil}
	if !reflect.DeepEqual(result["reported"], wantReported) {
		t.Fatalf("Unknown event fields became verified settings: %#v", result["reported"])
	}
	usage := result["usage"].(map[string]any)
	if usage["reasoning_output_tokens"] != nil || usage["capture_status"] != "captured" {
		t.Fatalf("Unexpected usage inference: %#v", usage)
	}
	if diagnostics := result["diagnostics"].([]map[string]any); len(diagnostics) != 0 {
		t.Fatalf("Valid unknown fields produced diagnostics: %#v", diagnostics)
	}
}
