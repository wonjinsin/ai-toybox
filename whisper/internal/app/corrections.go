package app

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadCorrections(path string) (map[string]string, error) {
	if path == "" {
		return map[string]string{}, nil
	}
	resolved, err := requireRegularFile(path, "corrections")
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("read corrections file %q: %w", resolved, err)
	}
	corrections := make(map[string]string)
	if err := json.Unmarshal(content, &corrections); err != nil {
		return nil, fmt.Errorf("decode corrections file %q: %w", resolved, err)
	}
	if _, exists := corrections[""]; exists {
		return nil, fmt.Errorf("corrections file %q contains an empty source", resolved)
	}
	return corrections, nil
}
