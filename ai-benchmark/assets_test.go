package benchmark

import (
	"encoding/json"
	"testing"
)

func TestRunTemplateKeepsUnobservedFieldsUnknown(t *testing.T) {
	var record map[string]any
	if err := json.Unmarshal(RunTemplate(), &record); err != nil {
		t.Fatalf("expected embedded JSON template: %v", err)
	}
	if record["format_version"] != "0.2" || record["record_status"] != "template" {
		t.Fatalf("unexpected template identity: %v", record)
	}
	reported := record["reported"].(map[string]any)
	if reported["model"] != nil || reported["effort"] != nil {
		t.Fatalf("template must not invent observations: %v", reported)
	}
	first := RunTemplate()
	first[0] = '!'
	if !json.Valid(RunTemplate()) {
		t.Fatal("caller mutated shared template data")
	}
}
