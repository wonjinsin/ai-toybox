package persistence

import (
	"context"

	"github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent"
	entcategory "github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent/category"
	entmc "github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent/merchantcategory"
	"github.com/wonjinsin/ledger/internal/core/domain"
)

type CategoryRepo struct {
	client *ent.Client
}

func NewCategoryRepo(client *ent.Client) *CategoryRepo {
	return &CategoryRepo{client: client}
}

func (r *CategoryRepo) GetOrCreate(ctx context.Context, name string) (*domain.Category, error) {
	row, err := r.client.Category.Query().Where(entcategory.Name(name)).Only(ctx)
	if ent.IsNotFound(err) {
		row, err = r.client.Category.Create().SetName(name).Save(ctx)
	}
	if err != nil {
		return nil, err
	}
	return &domain.Category{ID: row.ID, Name: row.Name}, nil
}

type MerchantCategoryRepo struct {
	client *ent.Client
}

func NewMerchantCategoryRepo(client *ent.Client) *MerchantCategoryRepo {
	return &MerchantCategoryRepo{client: client}
}

func (r *MerchantCategoryRepo) All(ctx context.Context) (map[string]int, error) {
	rows, err := r.client.MerchantCategory.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	cache := make(map[string]int, len(rows))
	for _, row := range rows {
		cache[row.Merchant] = row.CategoryID
	}
	return cache, nil
}

func (r *MerchantCategoryRepo) SaveBatch(ctx context.Context, entries map[string]int) error {
	if len(entries) == 0 {
		return nil
	}
	builders := make([]*ent.MerchantCategoryCreate, 0, len(entries))
	for merchant, categoryID := range entries {
		builders = append(builders,
			r.client.MerchantCategory.Create().SetMerchant(merchant).SetCategoryID(categoryID))
	}
	return r.client.MerchantCategory.CreateBulk(builders...).
		OnConflictColumns(entmc.FieldMerchant).
		UpdateCategoryID().
		Exec(ctx)
}
