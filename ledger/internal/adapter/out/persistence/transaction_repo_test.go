package persistence

import (
	"context"
	"testing"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

func seedTransactions(t *testing.T, db *DB) *TransactionRepo {
	t.Helper()
	ctx := context.Background()
	src, err := NewSourceRepo(db.Client).Create(ctx, "국민은행", domain.SourceKindBank)
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	repo := NewTransactionRepo(db.Client)
	txs := []domain.Transaction{
		{SourceID: src.ID, TxDate: "2026-07-01", Amount: -4500, Merchant: "스타벅스", Hash: "h1"},
		{SourceID: src.ID, TxDate: "2026-07-05", Amount: -520000, Merchant: "신한카드", Memo: "카드대금 출금", Hash: "h2"},
		{SourceID: src.ID, TxDate: "2026-07-10", Amount: 3000000, Merchant: "급여", Hash: "h3"},
	}
	if _, _, err := repo.BulkInsert(ctx, txs); err != nil {
		t.Fatalf("bulk insert: %v", err)
	}
	return repo
}

func TestTransactionRepoSearch(t *testing.T) {
	repo := seedTransactions(t, openTestDB(t))
	ctx := context.Background()

	all, err := repo.Search(ctx, "")
	if err != nil {
		t.Fatalf("search all: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("want 3 transactions, got %d", len(all))
	}

	matched, err := repo.Search(ctx, "카드대금")
	if err != nil {
		t.Fatalf("search match: %v", err)
	}
	if len(matched) != 1 || matched[0].Memo != "카드대금 출금" {
		t.Errorf("want the 카드대금 row, got %+v", matched)
	}
}

func TestTransactionRepoDelete(t *testing.T) {
	repo := seedTransactions(t, openTestDB(t))
	ctx := context.Background()

	target, err := repo.Search(ctx, "카드대금")
	if err != nil || len(target) != 1 {
		t.Fatalf("locate target: %v (%d rows)", err, len(target))
	}

	n, err := repo.Delete(ctx, []int{target[0].ID, 99999})
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if n != 1 {
		t.Errorf("want 1 deleted, got %d", n)
	}

	rest, err := repo.Search(ctx, "")
	if err != nil {
		t.Fatalf("search after delete: %v", err)
	}
	if len(rest) != 2 {
		t.Errorf("want 2 remaining, got %d", len(rest))
	}
}
