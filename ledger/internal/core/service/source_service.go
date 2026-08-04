package service

import (
	"context"
	"fmt"

	"github.com/wonjinsin/ledger/internal/core/domain"
	"github.com/wonjinsin/ledger/internal/core/port/out"
)

type SourceService struct {
	repo out.SourceRepo
}

func NewSourceService(repo out.SourceRepo) *SourceService {
	return &SourceService{repo: repo}
}

func (s *SourceService) Add(ctx context.Context, name, kind string) (*domain.Source, error) {
	if kind != domain.SourceKindBank && kind != domain.SourceKindCard {
		return nil, fmt.Errorf("invalid kind %q: must be %s or %s", kind, domain.SourceKindBank, domain.SourceKindCard)
	}
	return s.repo.Create(ctx, name, kind)
}

func (s *SourceService) List(ctx context.Context) ([]*domain.Source, error) {
	return s.repo.List(ctx)
}
