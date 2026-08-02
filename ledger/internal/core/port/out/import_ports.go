package out

import (
	"context"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

type MappingRepo interface {
	// Get returns (mapping, true, nil) on cache hit, (nil, false, nil) on miss.
	Get(ctx context.Context, headerHash string) (*domain.ColumnMapping, bool, error)
	Save(ctx context.Context, headerHash string, m *domain.ColumnMapping) error
}

type TransactionRepo interface {
	// BulkInsert stores transactions, silently skipping hash duplicates.
	// Returns (inserted, duplicatesSkipped).
	BulkInsert(ctx context.Context, txs []domain.Transaction) (int, int, error)
}
