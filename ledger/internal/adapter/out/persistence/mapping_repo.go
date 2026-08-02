package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent"
	entmapping "github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent/csvmapping"
	"github.com/wonjinsin/ledger/internal/core/domain"
)

type MappingRepo struct {
	client *ent.Client
}

func NewMappingRepo(client *ent.Client) *MappingRepo {
	return &MappingRepo{client: client}
}

func (r *MappingRepo) Get(ctx context.Context, headerHash string) (*domain.ColumnMapping, bool, error) {
	row, err := r.client.CSVMapping.Query().Where(entmapping.HeaderHash(headerHash)).Only(ctx)
	if ent.IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var m domain.ColumnMapping
	if err := json.Unmarshal([]byte(row.MappingJSON), &m); err != nil {
		return nil, false, fmt.Errorf("corrupt cached mapping %s: %w", headerHash, err)
	}
	return &m, true, nil
}

func (r *MappingRepo) Save(ctx context.Context, headerHash string, m *domain.ColumnMapping) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return r.client.CSVMapping.Create().
		SetHeaderHash(headerHash).
		SetMappingJSON(string(data)).
		OnConflictColumns(entmapping.FieldHeaderHash).
		UpdateMappingJSON().
		Exec(ctx)
}
