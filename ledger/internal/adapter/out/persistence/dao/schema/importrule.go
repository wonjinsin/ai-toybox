package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// ImportRule persists a user decision ("always skip/include rows matching pattern")
// made during interactive import.
type ImportRule struct {
	ent.Schema
}

func (ImportRule) Fields() []ent.Field {
	return []ent.Field{
		field.Int("source_id").Optional().Nillable(), // nil = applies to all sources
		field.String("pattern").NotEmpty(),           // matched against merchant/memo, case-insensitive substring
		field.String("action"),                       // skip | include
	}
}
