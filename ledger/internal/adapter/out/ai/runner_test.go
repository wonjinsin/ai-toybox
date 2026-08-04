package ai

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeBin writes an executable shell script and returns its path.
func fakeBin(t *testing.T, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fake-ai")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunReturnsTrimmedStdout(t *testing.T) {
	bin := fakeBin(t, `cat > /dev/null; printf '  hello world \n'`)
	out, err := NewCLIRunner(bin).Run(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if out != "hello world" {
		t.Errorf("want %q, got %q", "hello world", out)
	}
}

func TestRunPassesPromptViaStdin(t *testing.T) {
	bin := fakeBin(t, `cat`)
	out, err := NewCLIRunner(bin).Run(context.Background(), "the-prompt")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if out != "the-prompt" {
		t.Errorf("prompt not passed via stdin: got %q", out)
	}
}

func TestRunFailureIncludesStderr(t *testing.T) {
	bin := fakeBin(t, `cat > /dev/null; echo "boom" >&2; exit 1`)
	_, err := NewCLIRunner(bin).Run(context.Background(), "prompt")
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("error should include stderr, got: %v", err)
	}
}

func TestRunRespectsContextTimeout(t *testing.T) {
	bin := fakeBin(t, `sleep 10`)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := NewCLIRunner(bin).Run(ctx, "prompt")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("want DeadlineExceeded, got %v", err)
	}
}

func TestNewRunnerResolvesBackends(t *testing.T) {
	if _, err := NewRunner("claude"); err != nil {
		t.Errorf("claude: %v", err)
	}
	if _, err := NewRunner("codex"); err != nil {
		t.Errorf("codex: %v", err)
	}
	if _, err := NewRunner("gpt"); err == nil {
		t.Error("want error for unknown backend")
	}
}
