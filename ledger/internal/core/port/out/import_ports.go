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
	// ExistingHashes returns which of the given hashes are already stored.
	ExistingHashes(ctx context.Context, hashes []string) (map[string]bool, error)
	// UncategorizedMerchants lists distinct merchants with NULL category.
	UncategorizedMerchants(ctx context.Context) ([]string, error)
	// ApplyMerchantCategories sets category_id on uncategorized transactions
	// of each merchant. Returns the number of updated rows.
	ApplyMerchantCategories(ctx context.Context, byMerchant map[string]int) (int, error)
}

type CategoryRepo interface {
	GetOrCreate(ctx context.Context, name string) (*domain.Category, error)
}

type MerchantCategoryRepo interface {
	// All returns the merchant -> category_id cache.
	All(ctx context.Context) (map[string]int, error)
	SaveBatch(ctx context.Context, entries map[string]int) error
}

type RuleRepo interface {
	List(ctx context.Context) ([]*domain.ImportRule, error)
	Create(ctx context.Context, rule domain.ImportRule) (*domain.ImportRule, error)
	Delete(ctx context.Context, id int) error
}
