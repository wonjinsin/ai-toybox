package experiment

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"
)

func exactObject(value any, keys []string, label string) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", label)
	}
	var unknown, missing []string
	for key := range object {
		if !slices.Contains(keys, key) {
			unknown = append(unknown, key)
		}
	}
	for _, key := range keys {
		if _, exists := object[key]; !exists {
			missing = append(missing, key)
		}
	}
	if len(unknown) > 0 || len(missing) > 0 {
		sort.Strings(unknown)
		sort.Strings(missing)
		return nil, fmt.Errorf("%s: unknown keys %v; missing keys %v", label, unknown, missing)
	}
	return object, nil
}

func validText(value, label, pattern string) error {
	if value == "" || strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must be explicitly specified as nonempty text", label)
	}
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s must be valid UTF-8 text", label)
	}
	for _, character := range value {
		if character < 32 || character == 127 {
			return fmt.Errorf("%s must not contain control characters", label)
		}
	}
	if pattern != "" && !regexp.MustCompile("^(?:"+pattern+")$").MatchString(value) {
		return fmt.Errorf("%s has an invalid format", label)
	}
	return nil
}

func validateSettings(config Config) error {
	if config.FormatVersion != "0.1" {
		return fmt.Errorf("experiment.format_version must be '0.1'")
	}
	for _, entry := range []struct{ value, label, pattern string }{
		{config.ExperimentID, "experiment_id", `[A-Za-z0-9][A-Za-z0-9_.-]*`},
		{config.IsolationProfile, "isolation_profile", `[A-Za-z0-9][A-Za-z0-9_.-]*`},
		{config.Model, "model", `[A-Za-z0-9][A-Za-z0-9_./:-]*`},
		{config.Effort, "effort", `[a-z][a-z0-9_-]*`},
		{config.CLI.RequiredVersion, "cli.required_version", `[0-9]+\.[0-9]+\.[0-9]+(?:[-+][A-Za-z0-9.-]+)?`},
		{config.Task.ID, "task.id", ""},
		{config.Task.PromptPath, "task.prompt_path", ""},
		{config.Task.PromptVersion, "task.prompt_version", ""},
		{config.Task.PromptSHA256, "task.prompt_sha256", `[0-9a-f]{64}`},
	} {
		if err := validText(entry.value, entry.label, entry.pattern); err != nil {
			return err
		}
	}
	if config.CLI.Name != "codex" {
		return fmt.Errorf("cli.name: only the codex adapter is defined")
	}
	if config.CLI.Mode != "non_interactive" {
		return fmt.Errorf("cli.mode must be 'non_interactive'")
	}
	if config.Artifact.Delivery != "final_response" {
		return fmt.Errorf("artifact.delivery: only final_response is implemented; file delivery is unsupported")
	}
	for _, capability := range []struct {
		name    string
		entries []string
	}{
		{"tools", config.Capabilities.Tools},
		{"mcp_servers", config.Capabilities.MCPServers},
		{"agents", config.Capabilities.Agents},
	} {
		for _, entry := range capability.entries {
			if err := validText(entry, "capabilities."+capability.name+" entry", ""); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateStructure(value any) error {
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("experiment must be a JSON object")
	}
	if object["status"] != "ready" {
		return fmt.Errorf("experiment.status must be 'ready'; draft inputs cannot run")
	}
	_, err := exactObject(object, []string{
		"format_version", "status", "experiment_id", "task", "cli", "model",
		"effort", "capabilities", "artifact", "isolation_profile",
	}, "experiment")
	if err != nil {
		return err
	}
	for _, schema := range []struct {
		name string
		keys []string
	}{
		{"task", []string{"id", "prompt_path", "prompt_version", "prompt_sha256"}},
		{"cli", []string{"name", "required_version", "mode"}},
		{"artifact", []string{"delivery"}},
	} {
		if _, err := exactObject(object[schema.name], schema.keys, schema.name); err != nil {
			return err
		}
	}
	capabilities, err := exactObject(object["capabilities"], []string{"skills", "tools", "mcp_servers", "agents"}, "capabilities")
	if err != nil {
		return err
	}
	for _, key := range []string{"skills", "tools", "mcp_servers", "agents"} {
		entries, ok := capabilities[key].([]any)
		if !ok {
			return fmt.Errorf("capabilities.%s must be an explicit array; [] permits none", key)
		}
		if key == "skills" {
			for _, entry := range entries {
				if _, err := exactObject(entry, []string{"name", "path", "sha256", "invocation"}, "skill"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
