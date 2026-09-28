package runner

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunArchivesThenRefusesUnenforcedCLIConditions(t *testing.T) {
	path, _ := writeInput(t)
	output := filepath.Join(t.TempDir(), "run-001")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"run", "--output", output, path}, &stdout, &stderr)
	if code != 3 || !strings.Contains(stderr.String(), "cannot enforce") {
		t.Fatalf("exit=%d stdout=%s stderr=%s", code, &stdout, &stderr)
	}
	raw, err := os.ReadFile(filepath.Join(output, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	var record map[string]any
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatal(err)
	}
	if record["preflight"].(map[string]any)["status"] != "blocked" {
		t.Fatal(record)
	}
	if record["execution"].(map[string]any)["started_at"] != nil {
		t.Fatal("claimed model execution")
	}
	if record["cli"].(map[string]any)["version"] != nil {
		t.Fatal("invented CLI version")
	}
}

func TestOnlyRunCommandIsExposed(t *testing.T) {
	for _, args := range [][]string{{}, {"validate", "x"}, {"preview", "x"}, {"prepare", "x"}, {"inspect", "x"}, {"run"}, {"run", "--unknown"}, {"run", "a", "b"}} {
		var stdout, stderr bytes.Buffer
		if code := Run(args, &stdout, &stderr); code != 2 {
			t.Fatalf("%v: exit=%d", args, code)
		}
	}
	for _, args := range [][]string{{"--help"}, {"-h"}, {"run", "--help"}} {
		var stdout, stderr bytes.Buffer
		if code := Run(args, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "run") {
			t.Fatalf("%v: exit=%d", args, code)
		}
	}
}

func TestRunCreatesOneDefaultDirectoryAndNeverReusesIt(t *testing.T) {
	t.Chdir(t.TempDir())
	path, _ := writeInput(t)
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"run", path}, &stdout, &stderr); code != 3 {
		t.Fatalf("exit=%d: %s", code, &stderr)
	}
	entries, err := os.ReadDir("runs")
	if err != nil || len(entries) != 1 {
		t.Fatalf("default archive: %v %v", entries, err)
	}
	archive := filepath.Join("runs", entries[0].Name())
	before, err := os.ReadFile(filepath.Join(archive, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	if code := Run([]string{"run", "--output", archive, path}, &stdout, &stderr); code != 2 {
		t.Fatalf("existing output: exit=%d", code)
	}
	after, err := os.ReadFile(filepath.Join(archive, "run.json"))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("existing archive changed")
	}
}

func TestRunRejectsMissingInputWithoutCreatingAnArchive(t *testing.T) {
	t.Chdir(t.TempDir())
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"run", "missing.json"}, &stdout, &stderr); code != 2 {
		t.Fatalf("exit=%d", code)
	}
	if _, err := os.Stat("runs"); !os.IsNotExist(err) {
		t.Fatal("invalid input created output")
	}
}

func TestOutputFlagDoesNotConsumeAnotherOption(t *testing.T) {
	t.Chdir(t.TempDir())
	path, _ := writeInput(t)
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"run", "--output", "--help", path}, &stdout, &stderr); code != 2 {
		t.Fatalf("exit=%d", code)
	}
	if _, err := os.Stat("--help"); !os.IsNotExist(err) {
		t.Fatal("option used as output directory")
	}
}
