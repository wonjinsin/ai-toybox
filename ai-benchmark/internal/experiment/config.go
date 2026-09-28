// Package experiment validates benchmark input snapshots without starting a model.
package experiment

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Task struct {
	ID            string `json:"id"`
	PromptPath    string `json:"prompt_path"`
	PromptVersion string `json:"prompt_version"`
	PromptSHA256  string `json:"prompt_sha256"`
}

type CLI struct {
	Name            string `json:"name"`
	RequiredVersion string `json:"required_version"`
	Mode            string `json:"mode"`
}

type Skill struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	Invocation string `json:"invocation"`
}

type Capabilities struct {
	Skills     []Skill  `json:"skills"`
	Tools      []string `json:"tools"`
	MCPServers []string `json:"mcp_servers"`
	Agents     []string `json:"agents"`
}

type Artifact struct {
	Delivery string `json:"delivery"`
}

type Config struct {
	FormatVersion    string       `json:"format_version"`
	Status           string       `json:"status"`
	ExperimentID     string       `json:"experiment_id"`
	Task             Task         `json:"task"`
	CLI              CLI          `json:"cli"`
	Model            string       `json:"model"`
	Effort           string       `json:"effort"`
	Capabilities     Capabilities `json:"capabilities"`
	Artifact         Artifact     `json:"artifact"`
	IsolationProfile string       `json:"isolation_profile"`
}

type Experiment struct {
	Path       string
	Raw        []byte
	PromptPath string
	Prompt     []byte
	Config     Config
}

func Load(path string) (Experiment, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return Experiment{}, err
	}
	absolute, err = filepath.EvalSymlinks(absolute)
	if err != nil {
		return Experiment{}, fmt.Errorf("cannot resolve experiment: %w", err)
	}
	raw, err := ReadFile(absolute, "experiment", 1024*1024)
	if err != nil {
		return Experiment{}, err
	}
	structure, err := strictJSON(raw)
	if err != nil {
		return Experiment{}, err
	}
	if err := validateStructure(structure); err != nil {
		return Experiment{}, err
	}
	var config Config
	if err := json.Unmarshal(raw, &config); err != nil {
		return Experiment{}, err
	}
	if err := validateSettings(config); err != nil {
		return Experiment{}, err
	}
	if err := validateSkills(config.Capabilities.Skills, filepath.Dir(absolute)); err != nil {
		return Experiment{}, err
	}
	promptPath := config.Task.PromptPath
	if !filepath.IsAbs(promptPath) {
		promptPath = filepath.Join(filepath.Dir(absolute), promptPath)
	}
	promptPath, err = filepath.EvalSymlinks(promptPath)
	if err != nil {
		return Experiment{}, fmt.Errorf("cannot resolve prompt: %w", err)
	}
	prompt, err := ReadFile(promptPath, "prompt", 10*1024*1024)
	if err != nil {
		return Experiment{}, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(prompt)) != config.Task.PromptSHA256 {
		return Experiment{}, fmt.Errorf("task.prompt_sha256: prompt hash does not match the declared input")
	}
	if !utf8.Valid(prompt) {
		return Experiment{}, fmt.Errorf("prompt must contain valid UTF-8 text")
	}
	if strings.TrimFunc(string(prompt), textWhitespace) == "" {
		return Experiment{}, fmt.Errorf("task.prompt_path contains an empty prompt")
	}
	if hasDraftStatus(string(prompt)) {
		return Experiment{}, fmt.Errorf("task.prompt_path contains a draft; review and freeze the prompt first")
	}
	return Experiment{Path: absolute, Raw: raw, PromptPath: promptPath, Prompt: prompt, Config: config}, nil
}

func textWhitespace(character rune) bool {
	return unicode.IsSpace(character) || character >= 0x1c && character <= 0x1f
}

func hasDraftStatus(text string) bool {
	pattern := regexp.MustCompile(`(?im)^Status:[\s\p{Z}\x{000b}\x{0085}\x{001c}-\x{001f}]*Draft`)
	for _, match := range pattern.FindAllStringIndex(text, -1) {
		if match[1] == len(text) {
			return true
		}
		next, _ := utf8.DecodeRuneInString(text[match[1]:])
		if !unicode.IsLetter(next) && !unicode.IsNumber(next) && next != '_' {
			return true
		}
	}
	return false
}
