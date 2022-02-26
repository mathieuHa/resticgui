package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// AWSS3Provider holds the schema definition for the AWSS3Provider entity.
type AWSS3Provider struct {
	ent.Schema
}

// Fields of the AWSS3Provider.
func (AWSS3Provider) Fields() []ent.Field {
	return []ent.Field{
		field.String("bucketName").NotEmpty(),
		field.String("secretAccessKey").NotEmpty(),
		field.String("accessKeyID").NotEmpty(),
		field.String("defaultRegion").Optional(),
	}
}

// Edges of the AWSS3Provider.
func (AWSS3Provider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("repository", Repository.Type).
			Ref("awss3provider").
			Unique().
			// We add the "Required" method to the builder
			// to make this edge required on entity creation.
			// i.e. Card cannot be created without its owner.
			Required(),
	}
}
