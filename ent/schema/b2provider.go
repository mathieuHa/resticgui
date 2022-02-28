package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// B2Provider holds the schema definition for the B2Provider entity.
type B2Provider struct {
	ent.Schema
}

// Annotations of the B2Provider.
func (B2Provider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "B2Provider"},
	}
}

// Fields of the B2Provider.
func (B2Provider) Fields() []ent.Field {
	return []ent.Field{
		field.String("bucketName").NotEmpty(),
		field.String("accountKey").NotEmpty(),
		field.String("accountID").NotEmpty(),
		field.String("repoName").NotEmpty(),
	}
}

// Edges of the B2Provider.
func (B2Provider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("repository", Repository.Type).
			Ref("B2Provider").
			Unique().
			// We add the "Required" method to the builder
			// to make this edge required on entity creation.
			// i.e. Card cannot be created without its owner.
			Required(),
	}
}
