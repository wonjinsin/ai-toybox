package persistence

import (
	"context"

	"github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent"
	entrule "github.com/wonjinsin/ledger/internal/adapter/out/persistence/dao/ent/importrule"
	"github.com/wonjinsin/ledger/internal/core/domain"
)

type RuleRepo struct {
	client *ent.Client
}

func NewRuleRepo(client *ent.Client) *RuleRepo {
	return &RuleRepo{client: client}
}

func (r *RuleRepo) List(ctx context.Context) ([]*domain.ImportRule, error) {
	rows, err := r.client.ImportRule.Query().Order(ent.Asc(entrule.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	rules := make([]*domain.ImportRule, 0, len(rows))
	for _, row := range rows {
		rules = append(rules, toDomainRule(row))
	}
	return rules, nil
}

func (r *RuleRepo) Create(ctx context.Context, rule domain.ImportRule) (*domain.ImportRule, error) {
	b := r.client.ImportRule.Create().SetPattern(rule.Pattern).SetAction(rule.Action)
	if rule.SourceID != nil {
		b.SetSourceID(*rule.SourceID)
	}
	row, err := b.Save(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainRule(row), nil
}

func (r *RuleRepo) Delete(ctx context.Context, id int) error {
	return r.client.ImportRule.DeleteOneID(id).Exec(ctx)
}

func toDomainRule(row *ent.ImportRule) *domain.ImportRule {
	return &domain.ImportRule{ID: row.ID, SourceID: row.SourceID, Pattern: row.Pattern, Action: row.Action}
}
