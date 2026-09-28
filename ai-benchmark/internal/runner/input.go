// Package runner implements the single-command benchmark workflow.
package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type settings struct {
	TaskID        string   `json:"task_id"`
	Prompt        string   `json:"prompt"`
	PromptVersion string   `json:"prompt_version"`
	Model         string   `json:"model"`
	Effort        string   `json:"effort"`
	Skills        []string `json:"skills"`
	Tools         []string `json:"tools"`
	MCPServers    []string `json:"mcp_servers"`
	Agents        []string `json:"agents"`
}

type input struct {
	settings
	ConfigBytes, PromptBytes []byte
}

func loadInput(path string) (input, error) {
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return input{}, err
	}
	raw, err := readText(path, 1024*1024)
	if err != nil {
		return input{}, err
	}
	if err := uniqueKeys(raw); err != nil {
		return input{}, err
	}
	if err := validateScalarEscapes(raw); err != nil {
		return input{}, err
	}
	var config settings
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return input{}, err
	}
	for name, value := range map[string]string{"task_id": config.TaskID, "prompt": config.Prompt, "prompt_version": config.PromptVersion, "model": config.Model, "effort": config.Effort} {
		if strings.TrimSpace(value) == "" {
			return input{}, fmt.Errorf("%s is required", name)
		}
	}
	for name, values := range map[string][]string{"skills": config.Skills, "tools": config.Tools, "mcp_servers": config.MCPServers, "agents": config.Agents} {
		if values == nil {
			return input{}, fmt.Errorf("%s must be explicit; use [] for none", name)
		}
		seen := map[string]bool{}
		for _, value := range values {
			if strings.TrimSpace(value) == "" || seen[value] {
				return input{}, fmt.Errorf("%s contains an empty or duplicate entry", name)
			}
			seen[value] = true
		}
	}
	promptPath := config.Prompt
	if !filepath.IsAbs(promptPath) {
		promptPath = filepath.Join(filepath.Dir(path), promptPath)
	}
	prompt, err := readText(promptPath, 10*1024*1024)
	if err != nil {
		return input{}, err
	}
	if strings.TrimSpace(string(prompt)) == "" {
		return input{}, fmt.Errorf("prompt is empty")
	}
	if regexp.MustCompile(`(?im)^Status:[\s\p{Z}\x{0085}]*Draft\b`).Match(prompt) {
		return input{}, fmt.Errorf("prompt is a draft; finalize it before running")
	}
	return input{config, raw, prompt}, nil
}

// Settings are flat, so rejecting duplicate top-level keys avoids hidden overrides.
func uniqueKeys(raw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return fmt.Errorf("experiment must be a JSON object")
	}
	seen := map[string]bool{}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return err
		}
		name, ok := key.(string)
		if !ok || seen[name] {
			return fmt.Errorf("duplicate or invalid setting %q", key)
		}
		switch name {
		case "task_id", "prompt", "prompt_version", "model", "effort", "skills", "tools", "mcp_servers", "agents":
		default:
			return fmt.Errorf("unknown setting %q", name)
		}
		seen[name] = true
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return err
		}
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("unexpected data after experiment JSON")
	}
	return nil
}

func readText(path string, limit int64) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("%s must be a regular file of at most %d bytes", path, limit)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit || !utf8.Valid(raw) {
		return nil, fmt.Errorf("%s exceeds the size limit or contains invalid UTF-8", path)
	}
	return raw, nil
}

// encoding/json replaces unpaired surrogates; reject them before decoding.
func validateScalarEscapes(raw []byte) error {
	inString := false
	for index := 0; index < len(raw); index++ {
		if raw[index] == '"' {
			inString = !inString
			continue
		}
		if !inString || raw[index] != '\\' || index+1 >= len(raw) {
			continue
		}
		if raw[index+1] != 'u' {
			index++
			continue
		}
		code, valid := unicodeEscape(raw, index)
		if !valid {
			return fmt.Errorf("invalid Unicode escape")
		}
		switch {
		case code >= 0xd800 && code <= 0xdbff:
			low, valid := unicodeEscape(raw, index+6)
			if !valid || low < 0xdc00 || low > 0xdfff {
				return fmt.Errorf("unpaired Unicode surrogate")
			}
			index += 11
		case code >= 0xdc00 && code <= 0xdfff:
			return fmt.Errorf("unpaired Unicode surrogate")
		default:
			index += 5
		}
	}
	return nil
}

func unicodeEscape(raw []byte, offset int) (uint64, bool) {
	if offset+6 > len(raw) || raw[offset] != '\\' || raw[offset+1] != 'u' {
		return 0, false
	}
	value, err := strconv.ParseUint(string(raw[offset+2:offset+6]), 16, 16)
	return value, err == nil
}
