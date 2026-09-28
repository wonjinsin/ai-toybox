package experiment

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"unicode/utf8"
)

// SkillDigest hashes sorted relative paths and every regular file's content.
func SkillDigest(directory string) (string, error) {
	directory = filepath.Clean(directory)
	root, err := os.Lstat(directory)
	if err != nil || !root.IsDir() || root.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("skill path must be a real directory containing SKILL.md")
	}
	manifest, err := os.Stat(filepath.Join(directory, "SKILL.md"))
	if err != nil || !manifest.Mode().IsRegular() {
		return "", fmt.Errorf("skill path must be a real directory containing SKILL.md")
	}
	digest := sha256.New()
	err = filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkError error) error {
		if walkError != nil {
			return fmt.Errorf("cannot walk skill bundle: %w", walkError)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("skill bundle must not contain symbolic links")
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("skill bundle must contain only regular files")
		}
		relative, err := filepath.Rel(directory, path)
		if err != nil {
			return fmt.Errorf("cannot resolve skill file path: %w", err)
		}
		name := filepath.ToSlash(relative)
		if !utf8.ValidString(name) {
			return fmt.Errorf("skill file paths must contain valid UTF-8 text")
		}
		content, err := ReadFile(path, "skill file", 10*1024*1024)
		if err != nil {
			return err
		}
		digest.Write([]byte(name + "\x00"))
		contentHash := sha256.Sum256(content)
		digest.Write(contentHash[:])
		return nil
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}

func validateSkills(skills []Skill, base string) error {
	for index, skill := range skills {
		for _, prior := range skills[:index] {
			if prior.Name == skill.Name {
				return fmt.Errorf("skills must have unique explicit names")
			}
		}
		for _, entry := range []struct{ value, label, pattern string }{
			{skill.Name, "skill.name", `[A-Za-z0-9][A-Za-z0-9_-]*`},
			{skill.Path, "skill.path", ""},
			{skill.SHA256, "skill.sha256", `[0-9a-f]{64}`},
		} {
			if err := validText(entry.value, entry.label, entry.pattern); err != nil {
				return err
			}
		}
		if skill.Invocation != "explicit" {
			return fmt.Errorf("skill.invocation must be 'explicit'")
		}
		path := skill.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(base, path)
		}
		digest, err := SkillDigest(path)
		if err != nil {
			return err
		}
		if digest != skill.SHA256 {
			return fmt.Errorf("skill hash does not match the selected bundle")
		}
	}
	return nil
}
