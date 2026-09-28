package collector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// resolvePath follows existing symlinks while allowing a nonexistent final path.
func resolvePath(path string) (string, error) {
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = cwd + string(os.PathSeparator) + path
	}
	remaining := strings.Split(path, string(os.PathSeparator))
	resolved := string(os.PathSeparator)
	links := 0
	for len(remaining) > 0 {
		part := remaining[0]
		remaining = remaining[1:]
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			resolved = filepath.Dir(resolved)
			continue
		}
		candidate := filepath.Join(resolved, part)
		info, err := os.Lstat(candidate)
		if os.IsNotExist(err) {
			resolved = candidate
			continue
		}
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink == 0 {
			resolved = candidate
			continue
		}
		links++
		if links > 40 {
			return "", fmt.Errorf("too many symlinks resolving %q", path)
		}
		target, err := os.Readlink(candidate)
		if err != nil {
			return "", err
		}
		if filepath.IsAbs(target) {
			resolved = string(os.PathSeparator)
		}
		remaining = append(strings.Split(target, string(os.PathSeparator)), remaining...)
	}
	return resolved, nil
}

func pathsOverlap(first, second string) (bool, error) {
	separator := string(os.PathSeparator)
	firstPrefix := strings.TrimRight(first, separator) + separator
	secondPrefix := strings.TrimRight(second, separator) + separator
	if first == second || strings.HasPrefix(first, secondPrefix) || strings.HasPrefix(second, firstPrefix) {
		return true, nil
	}
	return hasAncestorIdentity(first, second)
}

func hasAncestorIdentity(first, second string) (bool, error) {
	for _, pair := range [][2]string{{first, second}, {second, first}} {
		root, err := os.Stat(pair[0])
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return false, err
		}
		for candidate := pair[1]; ; candidate = filepath.Dir(candidate) {
			info, err := os.Stat(candidate)
			if err == nil && os.SameFile(root, info) {
				return true, nil
			}
			if err != nil && !os.IsNotExist(err) {
				return false, err
			}
			if filepath.Dir(candidate) == candidate {
				break
			}
		}
	}
	return false, nil
}
