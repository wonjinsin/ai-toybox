package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Category holds the schema definition for a spending category.
type Category struct {
	ent.Schema
}

func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique().NotEmpty(),
	}
}
