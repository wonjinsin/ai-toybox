package runner

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareRecordFreezesInputsWithoutClaimingExecution(t *testing.T) {
	path, prompt := writeInput(t)
	in, err := loadInput(path)
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "run-001")
	if _, err := prepareRecord(in, output); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{"prompt.md": string(prompt), "experiment.json": string(in.ConfigBytes), "response.txt": "", "errors.log": ""} {
		raw, err := os.ReadFile(filepath.Join(output, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != want {
			t.Fatalf("%s changed", name)
		}
	}
	raw, err := os.ReadFile(filepath.Join(output, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	if record["run_id"] != "run-001" || record["format_version"] != "0.2" {
		t.Fatal(record)
	}
	if record["execution"].(map[string]any)["status"] != "not_started" {
		t.Fatal("claimed execution")
	}
	if record["prompt"].(map[string]any)["sha256"] != fmt.Sprintf("%x", sha256.Sum256(prompt)) {
		t.Fatal("prompt hash mismatch")
	}
	if record["reported"].(map[string]any)["model"] != nil {
		t.Fatal("requested model claimed as observed")
	}
	if _, err := prepareRecord(in, output); err == nil {
		t.Fatal("existing archive overwritten")
	}
}
