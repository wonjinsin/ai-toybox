package experiment

import (
	"fmt"
	"io"
	"math"
	"os"
)

// ReadFile preserves bytes and refuses nonregular files before opening them.
func ReadFile(path, label string, limit int64) ([]byte, error) {
	if limit < 0 || limit == math.MaxInt64 {
		return nil, fmt.Errorf("%s input limit must be between 0 and %d", label, int64(math.MaxInt64-1))
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", label, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s must be an existing regular file", label)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", label, err)
	}
	defer file.Close()
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s must remain a regular file while reading", label)
	}
	content, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", label, err)
	}
	if int64(len(content)) > limit {
		return nil, fmt.Errorf("%s exceeds the %d-byte input limit", label, limit)
	}
	return content, nil
}
