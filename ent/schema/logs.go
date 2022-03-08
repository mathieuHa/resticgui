package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Logs holds the schema definition for the Logs entity.
type Logs struct {
	ent.Schema
}

// Annotations of the Logs.
func (Logs) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "Logs"},
	}
}

// Fields of the Logs.
func (Logs) Fields() []ent.Field {
	return []ent.Field{
		field.Time("createdAt").Default(time.Now),
		field.String("bucketName"),
		field.String("error"),
		field.String("stderr"),
		field.String("stdout"),
		field.Bool("success").Default(false),
	}
}

// Edges of the Logs.
func (Logs) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("repository", Repository.Type).
			Ref("logs").
			Unique(),
	}
}
