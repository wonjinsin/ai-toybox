package out

import (
	"context"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

type SourceRepo interface {
	Create(ctx context.Context, name, kind string) (*domain.Source, error)
	GetByName(ctx context.Context, name string) (*domain.Source, error)
	List(ctx context.Context) ([]*domain.Source, error)
}
