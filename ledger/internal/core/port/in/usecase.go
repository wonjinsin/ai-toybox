package in

import (
	"context"

	"github.com/wonjinsin/ledger/internal/core/domain"
)

type SourceUsecase interface {
	Add(ctx context.Context, name, kind string) (*domain.Source, error)
	List(ctx context.Context) ([]*domain.Source, error)
}
