package persistence

import (
	"context"
	"fmt"

	"github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent"
	entsource "github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent/source"
	"github.com/wonjinsin/ledger/internal/core/domain"
)

type SourceRepo struct {
	client *ent.Client
}

func NewSourceRepo(client *ent.Client) *SourceRepo {
	return &SourceRepo{client: client}
}

func (r *SourceRepo) Create(ctx context.Context, name, kind string) (*domain.Source, error) {
	s, err := r.client.Source.Create().SetName(name).SetKind(kind).Save(ctx)
	if ent.IsConstraintError(err) {
		return nil, fmt.Errorf("source %q: %w", name, domain.ErrDuplicate)
	}
	if err != nil {
		return nil, err
	}
	return toDomainSource(s), nil
}

func (r *SourceRepo) GetByName(ctx context.Context, name string) (*domain.Source, error) {
	s, err := r.client.Source.Query().Where(entsource.Name(name)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainSource(s), nil
}

func (r *SourceRepo) List(ctx context.Context) ([]*domain.Source, error) {
	rows, err := r.client.Source.Query().Order(ent.Asc(entsource.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	sources := make([]*domain.Source, 0, len(rows))
	for _, s := range rows {
		sources = append(sources, toDomainSource(s))
	}
	return sources, nil
}

func toDomainSource(s *ent.Source) *domain.Source {
	return &domain.Source{ID: s.ID, Name: s.Name, Kind: s.Kind}
}
