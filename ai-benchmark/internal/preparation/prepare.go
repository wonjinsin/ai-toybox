// Package preparation freezes inputs without invoking a benchmark model.
package preparation

import (
	"ai-benchmark/internal/experiment"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

var ErrUnsupportedCondition = errors.New("unsupported execution condition")

func Run(input experiment.Experiment, output string) error {
	resolved, err := Preview(input, output)
	if err != nil {
		return err
	}
	return fmt.Errorf("%w: %s", ErrUnsupportedCondition, resolved["blockers"].([]string)[0])
}

func hash(raw []byte) string {
	value := sha256.Sum256(raw)
	return hex.EncodeToString(value[:])
}

// Preview describes an Experiment returned by experiment.Load without executing it.
func Preview(input experiment.Experiment, output string) (map[string]any, error) {
	output, err := filepath.Abs(output)
	if err != nil {
		return nil, err
	}
	config := input.Config
	mentions := make([]string, len(config.Capabilities.Skills))
	for index, skill := range config.Capabilities.Skills {
		mentions[index] = "$" + skill.Name
	}
	treatmentText := strings.Join(mentions, "\n")
	var treatment, treatmentHash any
	delivered := string(input.Prompt)
	if treatmentText != "" {
		treatment, treatmentHash = treatmentText, hash([]byte(treatmentText))
		delivered = treatmentText + "\n\n" + delivered
	}
	return map[string]any{
		"configuration_valid":  true,
		"launch_ready":         false,
		"blockers":             []string{fmt.Sprintf("No reviewed isolation adapter is installed for %q.", config.IsolationProfile)},
		"cli_version_required": config.CLI.RequiredVersion,
		"model_support":        "not_checked",
		"capability_controls":  "unverified",
		"candidate_argv": []string{
			"codex", "exec", "--ephemeral", "--json", "--model", config.Model,
			"--config", "model_reasoning_effort=" + strconv.Quote(config.Effort),
			"--output-last-message", filepath.Join(output, "response.txt"), "-",
		},
		"task_prompt_sha256": hash(input.Prompt),
		"treatment":          treatment, "treatment_sha256": treatmentHash,
		"prompt_sha256":    hash([]byte(delivered)),
		"delivered_prompt": delivered,
		"output":           output,
	}, nil
}
