package experiment

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillDigestIncludesAllFilesInPythonPathOrder(t *testing.T) {
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "a"), 0700); err != nil {
		t.Fatal(err)
	}
	files := []struct{ path, content string }{
		{"SKILL.md", "Fixture\n"},
		{"a/item.txt", "nested\n"},
		{"a.md", "sibling\n"},
	}
	expected := sha256.New()
	for _, file := range files {
		writeFile(t, filepath.Join(directory, filepath.FromSlash(file.path)), []byte(file.content))
		expected.Write([]byte(file.path + "\x00"))
		contentHash := sha256.Sum256([]byte(file.content))
		expected.Write(contentHash[:])
	}
	got, err := SkillDigest(directory)
	if err != nil || got != fmt.Sprintf("%x", expected.Sum(nil)) {
		t.Fatalf("incompatible bundle digest: %q, %v", got, err)
	}
	writeFile(t, filepath.Join(directory, "extra.txt"), []byte("new instructions"))
	changed, err := SkillDigest(directory)
	if err != nil || changed == got {
		t.Fatalf("added file did not change digest: %q, %v", changed, err)
	}
}

func TestSkillDigestRejectsMissingManifestAndSymlinks(t *testing.T) {
	directory := t.TempDir()
	if _, err := SkillDigest(directory); err == nil || !strings.Contains(err.Error(), "SKILL.md") {
		t.Fatalf("missing manifest accepted: %v", err)
	}
	manifest := filepath.Join(directory, "SKILL.md")
	writeFile(t, manifest, []byte("Fixture"))
	rootLink := filepath.Join(t.TempDir(), "root-link")
	if err := os.Symlink(directory, rootLink); err != nil {
		t.Fatal(err)
	}
	if _, err := SkillDigest(rootLink); err == nil {
		t.Fatal("linked root accepted")
	}
	if _, err := SkillDigest(rootLink + string(os.PathSeparator)); err == nil {
		t.Fatal("linked root with a trailing separator accepted")
	}
	for _, target := range []string{manifest, directory, filepath.Join(directory, "missing")} {
		link := filepath.Join(directory, "linked")
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		if _, err := SkillDigest(directory); err == nil || !strings.Contains(err.Error(), "symbolic links") {
			t.Fatalf("linked entry accepted: %v", err)
		}
		if err := os.Remove(link); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadVerifiesSkillSelectionAndBundleHash(t *testing.T) {
	path, config, _ := fixture(t)
	directory := filepath.Join(filepath.Dir(path), "skill")
	if err := os.Mkdir(directory, 0700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(directory, "SKILL.md"), []byte("Fixture"))
	hash, err := SkillDigest(directory)
	if err != nil {
		t.Fatal(err)
	}
	selection := Skill{Name: "scene-skill", Path: "skill", SHA256: hash, Invocation: "explicit"}
	config.Capabilities.Skills = []Skill{selection}
	writeConfig(t, path, config)
	if _, err := Load(path); err != nil {
		t.Fatalf("valid selected skill rejected: %v", err)
	}
	cases := []struct {
		selection Skill
		message   string
	}{
		{Skill{Name: "", Path: "skill", SHA256: hash, Invocation: "explicit"}, "skill.name"},
		{Skill{Name: "scene-skill", Path: "", SHA256: hash, Invocation: "explicit"}, "skill.path"},
		{Skill{Name: "scene-skill", Path: "skill", SHA256: "invalid", Invocation: "explicit"}, "skill.sha256"},
		{Skill{Name: "scene-skill", Path: "skill", SHA256: hash, Invocation: "auto"}, "skill.invocation"},
		{Skill{Name: "scene-skill", Path: "skill", SHA256: strings.Repeat("0", 64), Invocation: "explicit"}, "skill hash"},
		{Skill{Name: "scene-skill", Path: "missing", SHA256: hash, Invocation: "explicit"}, "skill path"},
	}
	for _, item := range cases {
		t.Run(item.message, func(t *testing.T) {
			selected := config
			selected.Capabilities.Skills = []Skill{item.selection}
			writeConfig(t, path, selected)
			expectError(t, path, item.message)
		})
	}
	config.Capabilities.Skills = []Skill{selection, selection}
	writeConfig(t, path, config)
	expectError(t, path, "unique")
	selection.Path = directory
	config.Capabilities.Skills = []Skill{selection}
	writeConfig(t, path, config)
	if _, err := Load(path); err != nil {
		t.Fatalf("absolute skill path rejected: %v", err)
	}
	writeFile(t, filepath.Join(directory, "helper.txt"), []byte("Unapproved instructions"))
	expectError(t, path, "skill hash")
}
