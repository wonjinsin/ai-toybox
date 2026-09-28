package runner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeInput(t *testing.T) (string, []byte) {
	t.Helper()
	dir := t.TempDir()
	prompt := []byte("Write a simple HTML page.\r\n")
	if err := os.WriteFile(filepath.Join(dir, "prompt.md"), prompt, 0600); err != nil {
		t.Fatal(err)
	}
	cfg := settings{TaskID: "task-001", Prompt: "prompt.md", PromptVersion: "v1", Model: "fixture-model", Effort: "high", Skills: []string{}, Tools: []string{}, MCPServers: []string{}, Agents: []string{}}
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "experiment.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return path, prompt
}

func TestLoadInputRejectsIncompleteOrAmbiguousSettings(t *testing.T) {
	for _, change := range []struct{ name, before, after string }{
		{"missing model", `"model": "fixture-model"`, `"model": ""`},
		{"missing effort", `"effort": "high"`, `"effort": null`},
		{"undecided skills", `"skills": []`, `"skills": null`},
		{"undeclared tools", `"tools": [],`, ``},
		{"unknown field", `"task_id":`, `"typo": true, "task_id":`},
		{"duplicate setting", `"model":`, `"model": "other", "model":`},
		{"case alias", `"model":`, `"Model": "other", "model":`},
		{"unpaired surrogate", `"model": "fixture-model"`, `"model": "fixture\ud800"`},
		{"trailing json", `"agents": []`, `"agents": []} {"more": true`},
		{"empty capability", `"tools": []`, `"tools": [""]`},
		{"duplicate capability", `"tools": []`, `"tools": ["shell", "shell"]`},
	} {
		t.Run(change.name, func(t *testing.T) {
			path, _ := writeInput(t)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(strings.Replace(string(raw), change.before, change.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := loadInput(path); err == nil {
				t.Fatal("invalid settings accepted")
			}
		})
	}
}

func TestLoadInputRejectsEmptyDraftAndNonUTF8Prompts(t *testing.T) {
	for _, prompt := range [][]byte{[]byte(" \n"), []byte("# Task\nStatus: Draft\nWrite HTML."), []byte("Status:\u00a0Draft\nWrite HTML."), {0xff}} {
		path, _ := writeInput(t)
		if err := os.WriteFile(filepath.Join(filepath.Dir(path), "prompt.md"), prompt, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadInput(path); err == nil {
			t.Fatalf("invalid prompt accepted: %q", prompt)
		}
	}
}

func TestRelativePromptUsesResolvedExperimentLocation(t *testing.T) {
	path, prompt := writeInput(t)
	linkDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(linkDir, "prompt.md"), []byte("Wrong task"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(linkDir, "experiment.json")
	if err := os.Symlink(path, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	in, err := loadInput(link)
	if err != nil {
		t.Fatal(err)
	}
	if string(in.PromptBytes) != string(prompt) {
		t.Fatal("read prompt from symlink directory")
	}
}

func TestLoadInputPreservesPromptAndSettings(t *testing.T) {
	path, prompt := writeInput(t)
	got, err := loadInput(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Model != "fixture-model" || got.Effort != "high" || string(got.PromptBytes) != string(prompt) {
		t.Fatalf("input not preserved: %+v", got)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got.ConfigBytes) != string(raw) {
		t.Fatal("configuration snapshot changed")
	}
}
