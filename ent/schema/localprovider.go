package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// LocalProvider holds the schema definition for the LocalProvider entity.
type LocalProvider struct {
	ent.Schema
}

// Annotations of the LocalProvider.
func (LocalProvider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "LocalProvider"},
	}
}

// Fields of the LocalProvider.
func (LocalProvider) Fields() []ent.Field {
	return []ent.Field{
		field.String("repoPath"),
		field.Bool("valid").Default(false),
	}
}

// Edges of the LocalProvider.
func (LocalProvider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("repository", Repository.Type).
			Ref("localProvider").
			Unique().
			// We add the "Required" method to the builder
			// to make this edge required on entity creation.
			// i.e. Card cannot be created without its owner.
			Required(),
	}
}
