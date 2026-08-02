package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Source holds the schema definition for an account source (bank account or card).
type Source struct {
	ent.Schema
}

func (Source) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique().NotEmpty(),
		field.String("kind"), // bank | card
	}
}
