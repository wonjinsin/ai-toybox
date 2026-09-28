// Package app implements the local benchmark preparation command interface.
package app

import (
	"ai-benchmark/internal/evidence"
	"ai-benchmark/internal/experiment"
	"ai-benchmark/internal/preparation"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
)

const usage = `Usage:
  ai-benchmark validate EXPERIMENT
  ai-benchmark preview EXPERIMENT --output DIRECTORY
  ai-benchmark prepare EXPERIMENT --output DIRECTORY
  ai-benchmark inspect EVENTS
  ai-benchmark run EXPERIMENT --output DIRECTORY

Prepare benchmark inputs; no reviewed live adapter is installed.
`

func Run(args []string, stdout, stderr io.Writer) int {
	selected, err := parse(args)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	if selected.help {
		fmt.Fprint(stdout, usage)
		return 0
	}
	result, err := dispatch(selected)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		if errors.Is(err, preparation.ErrUnsupportedCondition) {
			return 3
		}
		return 2
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	return 0
}

func dispatch(selected options) (any, error) {
	if selected.command != "inspect" {
		input, err := experiment.Load(selected.input)
		if err != nil {
			return nil, err
		}
		if selected.command == "preview" {
			return preparation.Preview(input, selected.output)
		}
		if selected.command == "prepare" {
			return preparation.Prepare(input, selected.output)
		}
		if selected.command == "run" {
			return nil, preparation.Run(input, selected.output)
		}
		return map[string]any{"configuration_valid": true, "launch_ready": false, "model_support": "not_checked"}, nil
	}
	path, err := filepath.EvalSymlinks(selected.input)
	if err != nil {
		return nil, err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	raw, err := experiment.ReadFile(path, "events", 64*1024*1024)
	if err != nil {
		return nil, err
	}
	sha := sha256.Sum256(raw)
	return map[string]any{
		"source":  map[string]any{"path": path, "sha256": hex.EncodeToString(sha[:])},
		"summary": evidence.Summarize(raw),
	}, nil
}
