package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// CSVMapping caches an AI-generated column mapping per CSV header format.
type CSVMapping struct {
	ent.Schema
}

func (CSVMapping) Fields() []ent.Field {
	return []ent.Field{
		field.String("header_hash").Unique().NotEmpty(),
		field.Text("mapping_json"),
	}
}
