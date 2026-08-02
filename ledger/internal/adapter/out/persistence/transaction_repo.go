package persistence

import (
	"context"

	"github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent"
	enttx "github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent/transaction"
	"github.com/wonjinsin/ledger/internal/core/domain"
)

type TransactionRepo struct {
	client *ent.Client
}

func NewTransactionRepo(client *ent.Client) *TransactionRepo {
	return &TransactionRepo{client: client}
}

func (r *TransactionRepo) ExistingHashes(ctx context.Context, hashes []string) (map[string]bool, error) {
	if len(hashes) == 0 {
		return map[string]bool{}, nil
	}
	rows, err := r.client.Transaction.Query().
		Where(enttx.HashIn(hashes...)).
		Select(enttx.FieldHash).
		Strings(ctx)
	if err != nil {
		return nil, err
	}
	existing := make(map[string]bool, len(rows))
	for _, h := range rows {
		existing[h] = true
	}
	return existing, nil
}

func (r *TransactionRepo) BulkInsert(ctx context.Context, txs []domain.Transaction) (int, int, error) {
	if len(txs) == 0 {
		return 0, 0, nil
	}
	hashes := make([]string, len(txs))
	for i, t := range txs {
		hashes[i] = t.Hash
	}
	seen, err := r.ExistingHashes(ctx, hashes)
	if err != nil {
		return 0, 0, err
	}

	var builders []*ent.TransactionCreate
	for _, t := range txs {
		if seen[t.Hash] {
			continue
		}
		seen[t.Hash] = true // also dedups within the batch itself
		b := r.client.Transaction.Create().
			SetSourceID(t.SourceID).
			SetTxDate(t.TxDate).
			SetAmount(t.Amount).
			SetMerchant(t.Merchant).
			SetMemo(t.Memo).
			SetRawLine(t.RawLine).
			SetHash(t.Hash)
		if t.CategoryID != nil {
			b.SetCategoryID(*t.CategoryID)
		}
		builders = append(builders, b)
	}
	if len(builders) > 0 {
		if err := r.client.Transaction.CreateBulk(builders...).Exec(ctx); err != nil {
			return 0, 0, err
		}
	}
	return len(builders), len(txs) - len(builders), nil
}
