package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"

	"github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent"
)

// DB bundles the ent client with the underlying sql handle
// (the raw handle is used later for read-only AI-generated queries).
type DB struct {
	Client *ent.Client
	SQL    *sql.DB
}

// Open opens (creating parent dirs if needed) the SQLite file and runs migrations.
func Open(ctx context.Context, path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	if err := client.Schema.Create(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate schema: %w", err)
	}
	return &DB{Client: client, SQL: db}, nil
}

func (d *DB) Close() error {
	return d.Client.Close()
}
