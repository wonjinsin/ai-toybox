package process

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunCommandIncludesProcessError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test helper executable uses POSIX shell")
	}

	executable := filepath.Join(t.TempDir(), "fails")
	writeExecutable(t, executable, "#!/bin/sh\nprintf 'decoder failed' >&2\nexit 7\n")

	err := Run(context.Background(), executable, nil)
	if err == nil {
		t.Fatal("Run() error = nil, want process error")
	}
	if !strings.Contains(err.Error(), "decoder failed") {
		t.Errorf("Run() error = %q, want process stderr", err)
	}
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
