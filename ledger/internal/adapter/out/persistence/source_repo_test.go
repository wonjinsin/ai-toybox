package persistence

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpenCreatesAllTables(t *testing.T) {
	db := openTestDB(t)
	for _, table := range []string{
		"sources", "categories", "transactions",
		"merchant_categories", "csv_mappings", "import_rules",
	} {
		var name string
		err := db.SQL.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %s not created: %v", table, err)
		}
	}
}

func TestReadOnlyExecutorRejectsWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ro, err := NewReadOnlyExecutor(path)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()

	if _, _, err := ro.Query(context.Background(), "INSERT INTO sources(name, kind) VALUES ('x','card')"); err == nil {
		t.Fatal("INSERT on read-only connection must fail")
	}
	cols, _, err := ro.Query(context.Background(), "SELECT id, name FROM sources")
	if err != nil || len(cols) != 2 {
		t.Errorf("SELECT should work: cols=%v err=%v", cols, err)
	}
}

func TestSourceRepoCreateAndList(t *testing.T) {
	db := openTestDB(t)
	repo := NewSourceRepo(db.Client)
	ctx := context.Background()

	s, err := repo.Create(ctx, "신한체크", domain.SourceKindCard)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if s.ID == 0 || s.Name != "신한체크" || s.Kind != domain.SourceKindCard {
		t.Errorf("unexpected source: %+v", s)
	}

	sources, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("want 1 source, got %d", len(sources))
	}
}

func TestSourceRepoDuplicateName(t *testing.T) {
	db := openTestDB(t)
	repo := NewSourceRepo(db.Client)
	ctx := context.Background()

	if _, err := repo.Create(ctx, "신한체크", domain.SourceKindCard); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := repo.Create(ctx, "신한체크", domain.SourceKindBank)
	if !errors.Is(err, domain.ErrDuplicate) {
		t.Errorf("want ErrDuplicate, got %v", err)
	}
}
