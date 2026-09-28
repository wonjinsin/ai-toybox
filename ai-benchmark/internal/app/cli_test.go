package app

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func invoke(args ...string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestHelpListsCommandsAndExecutionLimit(t *testing.T) {
	code, out, err := invoke("--help")
	if code != 0 || err != "" || !strings.Contains(out, "prepare") || !strings.Contains(out, "reviewed") {
		t.Fatalf("help is incomplete: code=%d out=%q err=%q", code, out, err)
	}
}

func TestInvalidCommandArgumentsReturnCodeTwo(t *testing.T) {
	for _, args := range [][]string{
		{}, {"unknown"}, {"validate"}, {"validate", "one", "two"},
		{"preview", "input"}, {"prepare", "input", "--output"},
		{"validate", "input", "--output", "out"}, {"inspect", "--unknown"},
		{"run", "input", "--output", "out", "--output", "other"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			code, out, err := invoke(args...)
			if code != 2 || out != "" || !strings.Contains(err, "error:") {
				t.Fatalf("bad syntax accepted: %d %q %q", code, out, err)
			}
		})
	}
}

func TestInspectReturnsSourceAndUsageWithoutRewritingEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "captured.jsonl")
	raw := []byte("{\"type\":\"turn.completed\",\"usage\":{\"input_tokens\":12,\"cached_input_tokens\":4,\"output_tokens\":5}}\n")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	code, out, diagnostic := invoke("inspect", path)
	var result map[string]any
	if code != 0 || diagnostic != "" || json.Unmarshal([]byte(out), &result) != nil {
		t.Fatalf("inspection failed: %d %q %q", code, out, diagnostic)
	}
	sha := sha256.Sum256(raw)
	if result["source"].(map[string]any)["sha256"] != hex.EncodeToString(sha[:]) {
		t.Fatal("source hash mismatch")
	}
	usage := result["summary"].(map[string]any)["usage"].(map[string]any)
	if usage["input_tokens"] != float64(12) {
		t.Fatalf("wrong usage: %v", usage)
	}
	after, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(raw, after) {
		t.Fatal("inspection rewrote evidence")
	}
}

func writeExperiment(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	prompt := []byte("Build the agreed scene.\r\n")
	if err := os.WriteFile(filepath.Join(directory, "task.md"), prompt, 0600); err != nil {
		t.Fatal(err)
	}
	raw := fmt.Sprintf(`{
  "format_version":"0.1","status":"ready","experiment_id":"solar-high",
  "task":{"id":"solar","prompt_path":"task.md","prompt_version":"v1","prompt_sha256":"%x"},
  "cli":{"name":"codex","required_version":"0.156.1","mode":"non_interactive"},
  "model":"fixture-model","effort":"high",
  "capabilities":{"skills":[],"tools":[],"mcp_servers":[],"agents":[]},
  "artifact":{"delivery":"final_response"},"isolation_profile":"pending-review"
}`, sha256.Sum256(prompt))
	path := filepath.Join(directory, "experiment.json")
	if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidateReportsInputValidityWithoutClaimingLaunchReadiness(t *testing.T) {
	path := writeExperiment(t)
	code, out, diagnostic := invoke("validate", path)
	var result map[string]any
	if code != 0 || diagnostic != "" || json.Unmarshal([]byte(out), &result) != nil {
		t.Fatalf("validation failed: %d %q %q", code, out, diagnostic)
	}
	if result["configuration_valid"] != true || result["launch_ready"] != false || result["model_support"] != "not_checked" {
		t.Fatalf("validation claimed unsupported facts: %v", result)
	}
}

func TestPreparationCommandsDispatchToPreviewAndArchive(t *testing.T) {
	path := writeExperiment(t)
	output := filepath.Join(t.TempDir(), "attempt-1")
	for _, command := range []string{"preview", "prepare"} {
		code, out, diagnostic := invoke(command, path, "--output", output)
		var result map[string]any
		if code != 0 || diagnostic != "" || json.Unmarshal([]byte(out), &result) != nil {
			t.Fatalf("%s failed: %d %q %q", command, code, out, diagnostic)
		}
		if command == "preview" {
			if result["delivered_prompt"] != "Build the agreed scene.\r\n" {
				t.Fatalf("missing preview: %v", result)
			}
			if _, err := os.Stat(output); !os.IsNotExist(err) {
				t.Fatal("preview wrote archive")
			}
		} else if result["record_status"] != "prepared" {
			t.Fatalf("missing preparation record: %v", result)
		}
	}
	if _, err := os.Stat(filepath.Join(output, "run.json")); err != nil {
		t.Fatal(err)
	}
}

func TestLiveRunReturnsUnsupportedCodeWithoutStarting(t *testing.T) {
	path := writeExperiment(t)
	output := filepath.Join(t.TempDir(), "attempt-1")
	code, out, diagnostic := invoke("run", path, "--output", output)
	if code != 3 || out != "" || !strings.Contains(diagnostic, "isolation") {
		t.Fatalf("missing unsupported diagnostic: %d %q %q", code, out, diagnostic)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("blocked run created output")
	}
}

func TestOutputFlagCannotConsumeAnotherOptionAsDirectory(t *testing.T) {
	for _, value := range []string{"--help", "--output=other", "--unknown"} {
		if _, err := parse([]string{"prepare", "input", "--output", value}); err == nil {
			t.Fatalf("option accepted as output path: %q", value)
		}
	}
	parsed, err := parse([]string{"prepare", "input", "--output=--help"})
	if err != nil || parsed.output != "--help" {
		t.Fatalf("explicit directory spelling rejected: %+v %v", parsed, err)
	}
}
