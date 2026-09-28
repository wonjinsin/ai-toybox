package runner

import (
	assets "ai-benchmark"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func prepareRecord(in input, output string) (map[string]any, error) {
	var template map[string]any
	if err := json.Unmarshal(assets.RunTemplate(), &template); err != nil {
		return nil, err
	}
	record := overlay(template, map[string]any{
		"record_status": "prepared", "task_id": in.TaskID, "run_id": filepath.Base(output),
		"experiment": map[string]any{"path": "experiment.json", "sha256": digest(in.ConfigBytes)},
		"requested":  map[string]any{"model": in.Model, "effort": in.Effort, "skills": in.Skills, "tools": in.Tools, "mcp_servers": in.MCPServers, "agents": in.Agents},
		"prompt": overlay(template["prompt"].(map[string]any), map[string]any{
			"version": in.PromptVersion, "sha256": digest(in.PromptBytes), "task_snapshot_path": "prompt.md", "task_sha256": digest(in.PromptBytes),
		}),
	})
	if err := os.MkdirAll(filepath.Dir(output), 0700); err != nil {
		return nil, err
	}
	if err := os.Mkdir(output, 0700); err != nil {
		return nil, fmt.Errorf("create new run directory: %w", err)
	}
	for name, data := range map[string][]byte{"prompt.md": in.PromptBytes, "experiment.json": in.ConfigBytes, "response.txt": nil, "errors.log": nil} {
		if err := os.WriteFile(filepath.Join(output, name), data, 0600); err != nil {
			return nil, err
		}
	}
	return record, saveRecord(output, record)
}

func saveRecord(output string, record map[string]any) error {
	raw, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(output, "run.json"), append(raw, '\n'), 0600)
}

func digest(raw []byte) string { return fmt.Sprintf("%x", sha256.Sum256(raw)) }

func overlay(base, values map[string]any) map[string]any {
	result := make(map[string]any, len(base)+len(values))
	for key, value := range base {
		result[key] = value
	}
	for key, value := range values {
		result[key] = value
	}
	return result
}
