package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Path holds the schema definition for the Path entity.
type Path struct {
	ent.Schema
}

// Fields of the Path.
func (Path) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.String("abs_path").NotEmpty(),
		field.Bool("exists").Default(true),
		field.Bool("is_dir"),
		field.Bool("included").Default(true),
	}
}

// Edges of the Path.
func (Path) Edges() []ent.Edge {
	return nil
}
