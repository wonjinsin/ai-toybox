package ai

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/wonjinsin/ledger/internal/core/port/out"
)

// CLIRunner implements out.AIRunner by piping the prompt to a local AI CLI
// (claude, codex) via stdin. Stdin avoids OS arg-length limits on large prompts.
type CLIRunner struct {
	bin  string
	args []string
}

var _ out.AIRunner = (*CLIRunner)(nil)

func NewCLIRunner(bin string, args ...string) *CLIRunner {
	return &CLIRunner{bin: bin, args: args}
}

func NewClaude() *CLIRunner { return NewCLIRunner("claude", "-p") }
func NewCodex() *CLIRunner  { return NewCLIRunner("codex", "exec", "-") }

// NewRunner resolves a backend name from the --ai flag.
func NewRunner(name string) (*CLIRunner, error) {
	switch name {
	case "claude":
		return NewClaude(), nil
	case "codex":
		return NewCodex(), nil
	default:
		return nil, fmt.Errorf("unknown AI backend %q: must be claude or codex", name)
	}
}

func (r *CLIRunner) Run(ctx context.Context, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, r.bin, r.args...)
	cmd.Stdin = strings.NewReader(prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("%s: %w", r.bin, ctx.Err())
		}
		return "", fmt.Errorf("%s failed: %w (stderr: %s)", r.bin, err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}
