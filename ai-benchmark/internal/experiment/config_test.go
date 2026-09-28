package experiment

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
}

func expectError(t *testing.T, path, fragment string) {
	t.Helper()
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), fragment) {
		t.Fatalf("expected error containing %q, got %v", fragment, err)
	}
}

func TestLoadRequiresAnExactExplicitSchema(t *testing.T) {
	path, _, _ := fixture(t)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ old, replacement, message string }{
		{`"effort": "high",`, "", "missing"},
		{`"effort": "high"`, `"reasoning": "high"`, "unknown"},
		{`"prompt_version": "v1"`, `"prompt_version": "v1", "extra": 1`, "unknown"},
		{`"tools": []`, `"tools": null`, "tools"},
		{`"skills": []`, `"skills": null`, "skills"},
		{`"mcp_servers": []`, `"mcp_servers": {}`, "mcp_servers"},
		{`"agents": []`, `"agents": "none"`, "agents"},
		{`"skills": []`, `"skills": [null]`, "skill"},
		{`"status": "ready"`, `"status": "draft"`, "ready"},
	}
	for _, item := range cases {
		t.Run(item.message+item.old, func(t *testing.T) {
			if !bytes.Contains(original, []byte(item.old)) {
				t.Fatal("fixture does not contain substitution")
			}
			writeFile(t, path, bytes.Replace(original, []byte(item.old), []byte(item.replacement), 1))
			expectError(t, path, item.message)
		})
	}
}

func TestLoadRejectsInvalidSettings(t *testing.T) {
	path, _, _ := fixture(t)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ old, replacement, message string }{
		{`"format_version": "0.1"`, `"format_version": "99"`, "format_version"},
		{`"effort": "high"`, `"effort": null`, "effort"},
		{`"effort": "high"`, `"effort": "HIGH"`, "effort"},
		{`"model": "fixture-model"`, `"model": "--unexpected-flag"`, "model"},
		{`"experiment_id": "solar-high"`, `"experiment_id": "../escape"`, "experiment_id"},
		{`"isolation_profile": "pending-review"`, `"isolation_profile": ""`, "isolation_profile"},
		{`"name": "codex"`, `"name": "unknown"`, "cli.name"},
		{`"mode": "non_interactive"`, `"mode": "interactive"`, "cli.mode"},
		{`"required_version": "0.156.1"`, `"required_version": "latest"`, "cli.required_version"},
		{`"id": "solar"`, `"id": "solar\u0000system"`, "task.id"},
		{`"id": "solar"`, `"id": "solar\u007fsystem"`, "task.id"},
		{`"prompt_version": "v1"`, `"prompt_version": " v1"`, "task.prompt_version"},
		{`"prompt_path": "task.md"`, `"prompt_path": ""`, "task.prompt_path"},
		{`"delivery": "final_response"`, `"delivery": "workspace_files"`, "file delivery"},
		{`"tools": []`, `"tools": [""]`, "tools"},
		{`"agents": []`, `"agents": [null]`, "agents"},
		{`"mcp_servers": []`, `"mcp_servers": [" test"]`, "mcp_servers"},
	}
	for _, item := range cases {
		t.Run(item.message+item.replacement, func(t *testing.T) {
			writeFile(t, path, bytes.Replace(original, []byte(item.old), []byte(item.replacement), 1))
			expectError(t, path, item.message)
		})
	}
}

func TestLoadRejectsAmbiguousAndInvalidJSON(t *testing.T) {
	path, _, _ := fixture(t)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		raw     []byte
		message string
	}{
		{[]byte(`[]`), "object"},
		{[]byte(`null`), "object"},
		{[]byte(`not-json`), "JSON"},
		{[]byte{0xff}, "UTF-8"},
		{[]byte(`{"status":"ready","status":"draft"}`), "duplicate"},
		{bytes.Replace(original, []byte(`"id": "solar"`), []byte(`"id":"solar", "id":"second"`), 1), "duplicate"},
		{bytes.Replace(original, []byte(`"id": "solar"`), []byte(`"id":"\ud800"`), 1), "UTF-8"},
		{bytes.Replace(original, []byte(`"id": "solar"`), []byte(`"id":"\udc00"`), 1), "UTF-8"},
		{bytes.Replace(original, []byte(`"id": "solar"`), []byte(`"id":"\ud800\u0041"`), 1), "UTF-8"},
		{bytes.Replace(original, []byte(`"id": "solar"`), []byte{'"', 'i', 'd', '"', ':', '"', 0xff, '"'}, 1), "UTF-8"},
		{[]byte(strings.Repeat("[", 2000) + strings.Repeat("]", 2000)), "nesting"},
	}
	for _, item := range cases {
		t.Run(item.message, func(t *testing.T) {
			writeFile(t, path, item.raw)
			expectError(t, path, item.message)
		})
	}
}

func TestLoadAcceptsValidUnicodeEscapes(t *testing.T) {
	path, _, _ := fixture(t)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{`"\ud83d\ude00"`, `"\\ud800"`, `"한글"`, `"\ufffd"`} {
		writeFile(t, path, bytes.Replace(original, []byte(`"solar"`), []byte(id), 1))
		if _, err := Load(path); err != nil {
			t.Fatalf("valid Unicode %s rejected: %v", id, err)
		}
	}
}

func TestLoadAllowsDraftAsPartOfAWord(t *testing.T) {
	path, config, _ := fixture(t)
	for _, content := range []string{"Status: Drafting", "Status: Drafté", "Status: Draft_edition"} {
		prompt := []byte(content)
		selected := config
		selected.Task.PromptSHA256 = digest(prompt)
		writeFile(t, filepath.Join(filepath.Dir(path), "task.md"), prompt)
		writeConfig(t, path, selected)
		if _, err := Load(path); err != nil {
			t.Fatalf("word beginning with Draft was rejected: %v", err)
		}
	}
}

func digest(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func fixture(t *testing.T) (string, Config, []byte) {
	t.Helper()
	directory, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	prompt := []byte("Build the agreed scene.\r\n")
	writeFile(t, filepath.Join(directory, "task.md"), prompt)
	config := Config{
		FormatVersion: "0.1", Status: "ready", ExperimentID: "solar-high",
		Task:  Task{ID: "solar", PromptPath: "task.md", PromptVersion: "v1", PromptSHA256: digest(prompt)},
		CLI:   CLI{Name: "codex", RequiredVersion: "0.156.1", Mode: "non_interactive"},
		Model: "fixture-model", Effort: "high",
		Capabilities: Capabilities{Skills: []Skill{}, Tools: []string{}, MCPServers: []string{}, Agents: []string{}},
		Artifact:     Artifact{Delivery: "final_response"}, IsolationProfile: "pending-review",
	}
	path := filepath.Join(directory, "experiment.json")
	writeConfig(t, path, config)
	return path, config, prompt
}

func TestLoadRejectsChangedEmptyDraftAndInvalidPrompts(t *testing.T) {
	path, config, _ := fixture(t)
	cases := []struct {
		content    []byte
		updateHash bool
		message    string
	}{
		{[]byte("changed"), false, "hash"},
		{[]byte(" \n\t"), true, "empty"},
		{[]byte("# Task\nStatus: Draft for review.\n"), true, "draft"},
		{[]byte("status:\u00a0draft\n"), true, "draft"},
		{[]byte("Status:\vDraft"), true, "draft"},
		{[]byte{0xff}, true, "UTF-8"},
	}
	for _, item := range cases {
		t.Run(item.message, func(t *testing.T) {
			writeFile(t, filepath.Join(filepath.Dir(path), "task.md"), item.content)
			selected := config
			if item.updateHash {
				selected.Task.PromptSHA256 = digest(item.content)
			}
			writeConfig(t, path, selected)
			expectError(t, path, item.message)
		})
	}
}

func TestLoadResolvesConfigAndPromptSymlinksAndAbsolutePromptPaths(t *testing.T) {
	path, config, prompt := fixture(t)
	link := filepath.Join(t.TempDir(), "experiment-link.json")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	result, err := Load(link)
	if err != nil || result.Path != path || !bytes.Equal(result.Prompt, prompt) {
		t.Fatalf("config target directory not used: %#v, %v", result, err)
	}
	absolutePrompt := filepath.Join(filepath.Dir(path), "task.md")
	config.Task.PromptPath = absolutePrompt
	writeConfig(t, path, config)
	result, err = Load(path)
	if err != nil || result.PromptPath != absolutePrompt {
		t.Fatalf("absolute prompt path not preserved: %#v, %v", result, err)
	}
	linkPrompt := filepath.Join(filepath.Dir(path), "linked-task.md")
	if err := os.Symlink(absolutePrompt, linkPrompt); err != nil {
		t.Fatal(err)
	}
	config.Task.PromptPath = "linked-task.md"
	writeConfig(t, path, config)
	result, err = Load(path)
	if err != nil || result.PromptPath != absolutePrompt {
		t.Fatalf("prompt link not resolved: %#v, %v", result, err)
	}
}

func writeConfig(t *testing.T, path string, config Config) {
	t.Helper()
	raw, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, raw)
}

func TestLoadPreservesExactInputs(t *testing.T) {
	path, config, prompt := fixture(t)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Load(path)
	if err != nil {
		t.Fatalf("valid experiment rejected: %v", err)
	}
	if !bytes.Equal(result.Raw, raw) || !bytes.Equal(result.Prompt, prompt) {
		t.Fatal("input bytes changed")
	}
	if result.Config.Model != config.Model || result.Path != path || result.PromptPath != filepath.Join(filepath.Dir(path), "task.md") {
		t.Fatalf("incorrect experiment: %#v", result)
	}
	if result.Config.Capabilities.Tools == nil || result.Config.Capabilities.Skills == nil {
		t.Fatal("explicit empty arrays became null")
	}
}
