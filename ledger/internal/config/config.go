package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	DBPath string
	AI     string // claude | codex
}

func DefaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "ledger.db"
	}
	return filepath.Join(home, ".ledger", "ledger.db")
}
