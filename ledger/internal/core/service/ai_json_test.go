package service

import (
	"context"
	"strings"
	"testing"
)

// scriptedRunner replays canned replies in order.
type scriptedRunner struct {
	replies []string
	prompts []string
}

func (s *scriptedRunner) Run(_ context.Context, prompt string) (string, error) {
	s.prompts = append(s.prompts, prompt)
	reply := s.replies[0]
	if len(s.replies) > 1 {
		s.replies = s.replies[1:]
	}
	return reply, nil
}

func TestRunJSONParsesFencedReply(t *testing.T) {
	r := &scriptedRunner{replies: []string{"Here you go:\n```json\n{\"name\":\"coffee\"}\n```"}}
	var v struct{ Name string }
	if err := runJSON(context.Background(), r, "p", &v); err != nil {
		t.Fatalf("runJSON: %v", err)
	}
	if v.Name != "coffee" {
		t.Errorf("want coffee, got %q", v.Name)
	}
	if len(r.prompts) != 1 {
		t.Errorf("want 1 call, got %d", len(r.prompts))
	}
}

func TestRunJSONRetriesOnceOnInvalidJSON(t *testing.T) {
	r := &scriptedRunner{replies: []string{"not json at all", `{"name":"ok"}`}}
	var v struct{ Name string }
	if err := runJSON(context.Background(), r, "p", &v); err != nil {
		t.Fatalf("runJSON: %v", err)
	}
	if v.Name != "ok" {
		t.Errorf("retry result not used: %+v", v)
	}
	if len(r.prompts) != 2 {
		t.Fatalf("want 2 calls, got %d", len(r.prompts))
	}
	if !strings.Contains(r.prompts[1], "not valid JSON") {
		t.Errorf("retry prompt should mention the failure")
	}
}

func TestRunJSONFailsAfterSecondInvalidReply(t *testing.T) {
	r := &scriptedRunner{replies: []string{"garbage", "still garbage"}}
	var v struct{ Name string }
	if err := runJSON(context.Background(), r, "p", &v); err == nil {
		t.Fatal("want error after two invalid replies")
	}
	if len(r.prompts) != 2 {
		t.Errorf("want exactly 2 calls, got %d", len(r.prompts))
	}
}

func TestExtractJSONPrefersOutermostObject(t *testing.T) {
	got := extractJSON("prefix {\"a\":[1,2]} suffix")
	if got != `{"a":[1,2]}` {
		t.Errorf("got %q", got)
	}
	if extractJSON(`[1,2,3]`) != `[1,2,3]` {
		t.Error("bare array should pass through")
	}
}
