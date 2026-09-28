package preparation

import (
	"ai-benchmark/internal/experiment"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func digest(raw []byte) string {
	value := sha256.Sum256(raw)
	return hex.EncodeToString(value[:])
}

func fixture(t *testing.T) experiment.Experiment {
	t.Helper()
	directory := t.TempDir()
	prompt := []byte("Build the agreed scene.\r\n")
	config := experiment.Config{
		FormatVersion: "0.1", Status: "ready", ExperimentID: "solar-high",
		Task:  experiment.Task{ID: "solar", PromptPath: "task.md", PromptVersion: "v1", PromptSHA256: digest(prompt)},
		CLI:   experiment.CLI{Name: "codex", RequiredVersion: "0.156.1", Mode: "non_interactive"},
		Model: "fixture-model", Effort: "high",
		Capabilities: experiment.Capabilities{Skills: []experiment.Skill{}, Tools: []string{}, MCPServers: []string{}, Agents: []string{}},
		Artifact:     experiment.Artifact{Delivery: "final_response"}, IsolationProfile: "pending-review",
	}
	raw, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	return experiment.Experiment{Path: filepath.Join(directory, "experiment.json"), Raw: raw, PromptPath: filepath.Join(directory, "task.md"), Prompt: prompt, Config: config}
}

func TestPreviewDisclosesUnverifiedControlsWithoutCreatingOutput(t *testing.T) {
	input := fixture(t)
	output := filepath.Join(t.TempDir(), "attempt-1")
	result, err := Preview(input, output)
	if err != nil {
		t.Fatal(err)
	}
	if result["launch_ready"] != false || result["capability_controls"] != "unverified" {
		t.Fatalf("preview must disclose unverified execution: %v", result)
	}
	if result["prompt_sha256"] != digest(input.Prompt) || result["delivered_prompt"] != string(input.Prompt) {
		t.Fatalf("prompt bytes changed: %v", result)
	}
	args, ok := result["candidate_argv"].([]string)
	if !ok || !strings.Contains(strings.Join(args, " "), `model_reasoning_effort="high"`) {
		t.Fatalf("missing explicit CLI arguments: %v", args)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("preview created output: %v", err)
	}
}

func TestPreviewSeparatesOrderedSkillTreatmentFromSharedTask(t *testing.T) {
	input := fixture(t)
	selected := input
	selected.Config.Capabilities.Skills = []experiment.Skill{{Name: "scene-skill"}, {Name: "second-skill"}}
	result, err := Preview(selected, filepath.Join(t.TempDir(), "attempt-1"))
	if err != nil {
		t.Fatal(err)
	}
	treatment := "$scene-skill\n$second-skill"
	if result["treatment"] != treatment || result["treatment_sha256"] != digest([]byte(treatment)) {
		t.Fatalf("skill treatment not preserved: %v", result)
	}
	if result["task_prompt_sha256"] != digest(input.Prompt) || result["delivered_prompt"] != treatment+"\n\n"+string(input.Prompt) {
		t.Fatalf("shared task or combined prompt changed: %v", result)
	}
	if result["prompt_sha256"] == result["task_prompt_sha256"] {
		t.Fatal("combined hash must include treatment")
	}
}

func TestPrepareFreezesInputsAndLeavesObservationsUnknown(t *testing.T) {
	input := fixture(t)
	output := filepath.Join(t.TempDir(), "archive", "attempt-1")
	record, err := Prepare(input, output)
	if err != nil {
		t.Fatal(err)
	}
	if record["record_status"] != "prepared" || record["run_id"] != "attempt-1" {
		t.Fatalf("invalid record: %v", record)
	}
	for name, expected := range map[string]string{
		"experiment.json": string(input.Raw), "task-prompt.md": string(input.Prompt), "prompt.md": string(input.Prompt), "response.txt": "", "errors.log": "",
	} {
		raw, err := os.ReadFile(filepath.Join(output, name))
		if err != nil || string(raw) != expected {
			t.Fatalf("incorrect %s snapshot: %q, %v", name, raw, err)
		}
	}
	if record["execution"].(map[string]any)["status"] != "not_started" || record["reported"].(map[string]any)["model"] != nil {
		t.Fatalf("preparation invented observations: %v", record)
	}
	if record["cli"].(map[string]any)["version"] != nil {
		t.Fatal("CLI version was not measured")
	}
	if record["context"].(map[string]any)["isolation_status"] != "unverified" {
		t.Fatal("isolation was not verified")
	}
	if record["requested"].(map[string]any)["model"] != input.Config.Model {
		t.Fatal("lost requested model")
	}
	for _, name := range []string{"run.json", "preview.json"} {
		raw, err := os.ReadFile(filepath.Join(output, name))
		if err != nil || !json.Valid(raw) {
			t.Fatalf("invalid %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(output, "treatment.md")); !os.IsNotExist(err) {
		t.Fatalf("unexpected treatment file: %v", err)
	}
}

func TestPreparePreservesTreatmentAsSeparateEvidence(t *testing.T) {
	input := fixture(t)
	input.Config.Capabilities.Skills = []experiment.Skill{{Name: "scene-skill"}}
	output := filepath.Join(t.TempDir(), "attempt-1")
	record, err := Prepare(input, output)
	if err != nil {
		t.Fatal(err)
	}
	prompt := record["prompt"].(map[string]any)
	if prompt["treatment_snapshot_path"] != "treatment.md" {
		t.Fatalf("missing treatment reference: %v", prompt)
	}
	raw, err := os.ReadFile(filepath.Join(output, "treatment.md"))
	if err != nil || string(raw) != "$scene-skill" {
		t.Fatalf("treatment bytes lost: %q, %v", raw, err)
	}
	combined, err := os.ReadFile(filepath.Join(output, "prompt.md"))
	if err != nil || string(combined) != "$scene-skill\n\n"+string(input.Prompt) {
		t.Fatalf("combined prompt incorrect: %q %v", combined, err)
	}
}

func TestPrepareRefusesExistingArchive(t *testing.T) {
	input := fixture(t)
	output := filepath.Join(t.TempDir(), "attempt-1")
	if _, err := Prepare(input, output); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(output, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Prepare(input, output); !os.IsExist(err) {
		t.Fatalf("existing archive accepted: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(output, "run.json"))
	if err != nil || string(after) != string(before) {
		t.Fatalf("existing evidence changed: %v", err)
	}
}

func TestRunRefusesUnreviewedIsolationBeforeCreatingOutput(t *testing.T) {
	input := fixture(t)
	output := filepath.Join(t.TempDir(), "attempt-1")
	err := Run(input, output)
	if !errors.Is(err, ErrUnsupportedCondition) || !strings.Contains(err.Error(), "isolation") {
		t.Fatalf("missing isolation refusal: %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("blocked run created output: %v", err)
	}
}

func TestStoredRecordMatchesReturnValueAndKeepsEmptyAllowlists(t *testing.T) {
	output := filepath.Join(t.TempDir(), "attempt-1")
	record, err := Prepare(fixture(t), output)
	if err != nil {
		t.Fatal(err)
	}
	returned, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	disk, err := os.ReadFile(filepath.Join(output, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	var expected, actual map[string]any
	if err := json.Unmarshal(returned, &expected); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(disk, &actual); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Fatal("returned and stored records differ")
	}
	requested := actual["requested"].(map[string]any)
	for _, key := range []string{"skills", "tools", "mcp_servers", "agents"} {
		entries, ok := requested[key].([]any)
		if !ok || len(entries) != 0 {
			t.Fatalf("%s [] became null or nonempty: %v", key, requested[key])
		}
	}
}

func TestPrepareRefusesEmptyDirectoriesAndSymlinks(t *testing.T) {
	for _, kind := range []string{"empty-directory", "symlink", "dangling-symlink", "parent-file"} {
		t.Run(kind, func(t *testing.T) {
			directory := t.TempDir()
			output := filepath.Join(directory, "archive")
			target := filepath.Join(directory, "target")
			switch kind {
			case "empty-directory":
				if err := os.Mkdir(output, 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Mkdir(target, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, output); err != nil {
					t.Skip(err)
				}
			case "dangling-symlink":
				if err := os.Symlink(target, output); err != nil {
					t.Skip(err)
				}
			case "parent-file":
				if err := os.WriteFile(output, []byte("retain"), 0600); err != nil {
					t.Fatal(err)
				}
				output = filepath.Join(output, "child")
			}
			if _, err := Prepare(fixture(t), output); err == nil {
				t.Fatal("existing destination accepted")
			}
			if kind == "symlink" {
				entries, err := os.ReadDir(target)
				if err != nil || len(entries) != 0 {
					t.Fatalf("symlink target changed: %v", err)
				}
			}
		})
	}
}
