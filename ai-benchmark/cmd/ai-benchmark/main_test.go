package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBinarySavesBlockedAttemptOutsideCheckout(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "ai-benchmark")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	prompt := []byte("Build the agreed scene.\r\n")
	if err := os.WriteFile(filepath.Join(directory, "task.md"), prompt, 0600); err != nil {
		t.Fatal(err)
	}
	config := `{"task_id":"solar","prompt":"task.md","prompt_version":"v1",
"model":"fixture-model","effort":"high","skills":[],"tools":[],"mcp_servers":[],"agents":[]}`
	if err := os.WriteFile(filepath.Join(directory, "experiment.json"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary, "run", "--output", "archive", "experiment.json")
	command.Dir = directory
	output, err := command.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 3 || !strings.Contains(string(output), "cannot enforce") {
		t.Fatalf("binary exit code: %v %s", err, output)
	}
	var record map[string]any
	raw, err := os.ReadFile(filepath.Join(directory, "archive", "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	if json.Unmarshal(raw, &record) != nil || record["record_status"] != "prepared" {
		t.Fatalf("binary did not prepare an archive: %s", output)
	}
	for _, name := range []string{"run.json", "prompt.md", "response.txt", "errors.log"} {
		if _, err := os.Stat(filepath.Join(directory, "archive", name)); err != nil {
			t.Fatal(err)
		}
	}
}
