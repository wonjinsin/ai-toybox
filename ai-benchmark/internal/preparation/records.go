package preparation

import (
	assets "ai-benchmark"
	"ai-benchmark/internal/experiment"
	"encoding/json"
	"os"
	"path/filepath"
)

type snapshot struct {
	name    string
	content []byte
}

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

// Prepare archives an Experiment returned by experiment.Load in a new directory.
func Prepare(input experiment.Experiment, output string) (map[string]any, error) {
	resolved, err := Preview(input, output)
	if err != nil {
		return nil, err
	}
	output = resolved["output"].(string)
	var template map[string]any
	if err := json.Unmarshal(assets.RunTemplate(), &template); err != nil {
		return nil, err
	}
	config := input.Config
	var treatmentPath any
	if resolved["treatment"] != nil {
		treatmentPath = "treatment.md"
	}
	record := overlay(template, map[string]any{
		"record_status": "prepared", "task_id": config.Task.ID, "run_id": filepath.Base(output),
		"experiment": map[string]any{"path": "experiment.json", "sha256": hash(input.Raw)},
		"cli":        overlay(template["cli"].(map[string]any), map[string]any{"name": config.CLI.Name}),
		"requested": map[string]any{
			"model": config.Model, "effort": config.Effort, "skills": config.Capabilities.Skills,
			"tools": config.Capabilities.Tools, "mcp_servers": config.Capabilities.MCPServers, "agents": config.Capabilities.Agents,
		},
		"prompt": overlay(template["prompt"].(map[string]any), map[string]any{
			"version": config.Task.PromptVersion, "sha256": resolved["prompt_sha256"],
			"task_snapshot_path": "task-prompt.md", "task_sha256": resolved["task_prompt_sha256"],
			"treatment_sha256":        resolved["treatment_sha256"],
			"treatment_snapshot_path": treatmentPath,
		}),
	})
	previewJSON, err := json.MarshalIndent(resolved, "", "  ")
	if err != nil {
		return nil, err
	}
	recordJSON, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0700); err != nil {
		return nil, err
	}
	if err := os.Mkdir(output, 0700); err != nil {
		return nil, err
	}
	files := []snapshot{
		{"experiment.json", input.Raw}, {"task-prompt.md", input.Prompt},
		{"prompt.md", []byte(resolved["delivered_prompt"].(string))},
		{"response.txt", nil}, {"errors.log", nil}, {"preview.json", append(previewJSON, '\n')},
	}
	if treatment, ok := resolved["treatment"].(string); ok {
		files = append(files, snapshot{"treatment.md", []byte(treatment)})
	}
	files = append(files, snapshot{"run.json", append(recordJSON, '\n')})
	for _, file := range files {
		if err := writeNew(filepath.Join(output, file.name), file.content); err != nil {
			return nil, err
		}
	}
	return record, nil
}

func writeNew(path string, raw []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, writeErr := file.Write(raw)
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}
