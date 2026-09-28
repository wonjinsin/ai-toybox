package experiment

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFilePreservesBytesWithinAnExplicitLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input")
	raw := []byte{0xff, 0, 10}
	writeFile(t, path, raw)
	content, err := ReadFile(path, "fixture", 3)
	if err != nil || !bytes.Equal(content, raw) {
		t.Fatalf("bounded read changed bytes: %x, %v", content, err)
	}
	for _, limit := range []int64{-1, 0, 2, math.MaxInt64} {
		if _, err := ReadFile(path, "fixture", limit); err == nil || !strings.Contains(err.Error(), "limit") {
			t.Fatalf("limit %d must reject input: %v", limit, err)
		}
	}
	writeFile(t, path, nil)
	if content, err := ReadFile(path, "fixture", 0); err != nil || len(content) != 0 {
		t.Fatalf("zero-byte file with zero-byte limit: %v", err)
	}
}

func TestReadFileRequiresRegularFile(t *testing.T) {
	directory := t.TempDir()
	for _, path := range []string{directory, filepath.Join(directory, "missing")} {
		if _, err := ReadFile(path, "fixture", 10); err == nil || !strings.Contains(err.Error(), "fixture") {
			t.Fatalf("invalid file accepted: %s: %v", path, err)
		}
	}
	file := filepath.Join(directory, "file")
	writeFile(t, file, []byte("exact"))
	link := filepath.Join(directory, "link")
	if err := os.Symlink(file, link); err != nil {
		t.Fatal(err)
	}
	content, err := ReadFile(link, "fixture", 5)
	if err != nil || string(content) != "exact" {
		t.Fatalf("regular-file symlink rejected: %v", err)
	}
}

func TestLoadAndSkillDigestEnforceInputLimits(t *testing.T) {
	path, config, _ := fixture(t)
	writeFile(t, path, bytes.Repeat([]byte(" "), 1024*1024+1))
	expectError(t, path, "experiment exceeds")
	writeConfig(t, path, config)
	large := bytes.Repeat([]byte("x"), 10*1024*1024+1)
	writeFile(t, filepath.Join(filepath.Dir(path), "task.md"), large)
	expectError(t, path, "prompt exceeds")
	directory := t.TempDir()
	writeFile(t, filepath.Join(directory, "SKILL.md"), large)
	if _, err := SkillDigest(directory); err == nil || !strings.Contains(err.Error(), "skill file exceeds") {
		t.Fatalf("oversized skill file accepted: %v", err)
	}
}
