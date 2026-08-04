package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// MerchantCategory caches AI-assigned categories per merchant.
type MerchantCategory struct {
	ent.Schema
}

func (MerchantCategory) Fields() []ent.Field {
	return []ent.Field{
		field.String("merchant").Unique().NotEmpty(),
		field.Int("category_id"),
	}
}
