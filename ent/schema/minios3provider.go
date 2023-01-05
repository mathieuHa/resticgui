package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// MinioS3Provider holds the schema definition for the MinioS3Provider entity.
type MinioS3Provider struct {
	ent.Schema
}

// Annotations of the MinioS3Provider.
func (MinioS3Provider) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "MinioS3Provider"},
	}
}

// Fields of the MinioS3Provider.
func (MinioS3Provider) Fields() []ent.Field {
	return []ent.Field{
		field.String("bucketName"),
		field.String("hostname"),
		field.String("port"),
		field.String("scheme"),
		field.String("accessKeyID"),
		field.String("secretAccessKey"),
		field.String("certificateAuthority"),
		field.Bool("valid").Default(false),
	}
}

// Edges of the MinioS3Provider.
func (MinioS3Provider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("repository", Repository.Type).
			Ref("MinioS3Provider").
			Unique().
			// We add the "Required" method to the builder
			// to make this edge required on entity creation.
			// i.e. Card cannot be created without its owner.
			Required(),
	}
}
