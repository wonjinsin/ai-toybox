//go:build darwin || linux

package experiment

import (
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestReadFileRejectsNamedPipeWithoutWaitingForWriter(t *testing.T) {
	pipe := filepath.Join(t.TempDir(), "pipe")
	if err := syscall.Mkfifo(pipe, 0600); err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() {
		_, err := ReadFile(pipe, "fixture", 10)
		result <- err
	}()
	select {
	case err := <-result:
		if err == nil || !strings.Contains(err.Error(), "regular file") {
			t.Fatalf("named pipe accepted: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("read waited for a named-pipe writer")
	}
}

func TestSkillDigestRejectsNamedPipe(t *testing.T) {
	directory := t.TempDir()
	writeFile(t, filepath.Join(directory, "SKILL.md"), []byte("Fixture"))
	if err := syscall.Mkfifo(filepath.Join(directory, "pipe"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := SkillDigest(directory); err == nil || !strings.Contains(err.Error(), "regular files") {
		t.Fatalf("named pipe accepted in skill: %v", err)
	}
}
