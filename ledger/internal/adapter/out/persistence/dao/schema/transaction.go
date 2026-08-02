package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Transaction holds the schema definition for a ledger transaction.
// Plain FK int fields (no ent edges) keep column names predictable for
// AI-generated SQL (text-to-SQL prompts include this schema as-is).
type Transaction struct {
	ent.Schema
}

func (Transaction) Fields() []ent.Field {
	return []ent.Field{
		field.Int("source_id"),
		field.String("tx_date"), // YYYY-MM-DD; SQLite date() functions work on this format
		field.Int64("amount"),   // KRW; expense negative, income positive
		field.String("merchant"),
		field.String("memo").Default(""),
		field.Int("category_id").Optional().Nillable(),
		field.String("raw_line"),
		field.String("hash").Unique(), // sha256(source|date|amount|merchant)
	}
}

func (Transaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tx_date"),
		index.Fields("source_id"),
		index.Fields("category_id"),
	}
}
