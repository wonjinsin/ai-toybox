package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBinaryPreparesArchiveOutsideCheckout(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "ai-benchmark")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	prompt := []byte("Build the agreed scene.\r\n")
	if err := os.WriteFile(filepath.Join(directory, "task.md"), prompt, 0600); err != nil {
		t.Fatal(err)
	}
	config := fmt.Sprintf(`{"format_version":"0.1","status":"ready","experiment_id":"fixture",
"task":{"id":"solar","prompt_path":"task.md","prompt_version":"v1","prompt_sha256":"%x"},
"cli":{"name":"codex","required_version":"0.156.1","mode":"non_interactive"},
"model":"fixture-model","effort":"high","capabilities":{"skills":[],"tools":[],"mcp_servers":[],"agents":[]},
"artifact":{"delivery":"final_response"},"isolation_profile":"pending-review"}`, sha256.Sum256(prompt))
	if err := os.WriteFile(filepath.Join(directory, "experiment.json"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary, "prepare", "experiment.json", "--output", "archive")
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("prepare: %v %s", err, output)
	}
	var record map[string]any
	if json.Unmarshal(output, &record) != nil || record["record_status"] != "prepared" {
		t.Fatalf("binary did not prepare an archive: %s", output)
	}
	for _, name := range []string{"run.json", "prompt.md", "response.txt", "errors.log"} {
		if _, err := os.Stat(filepath.Join(directory, "archive", name)); err != nil {
			t.Fatal(err)
		}
	}
	blocked := exec.Command(binary, "run", "experiment.json", "--output", "blocked")
	blocked.Dir = directory
	output, err = blocked.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.ExitCode() != 3 || !strings.Contains(string(output), "isolation") {
		t.Fatalf("binary exit code: %v %s", err, output)
	}
	if _, err := os.Stat(filepath.Join(directory, "blocked")); !os.IsNotExist(err) {
		t.Fatal("blocked command created output")
	}
}
