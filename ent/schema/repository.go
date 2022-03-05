package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Repository holds the schema definition for the Repository entity.
type Repository struct {
	ent.Schema
}

// Fields of the Repository.
func (Repository) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").Unique().NotEmpty(),
		field.Time("createdAt").
			Default(time.Now).Immutable(),
		field.Enum("provider").Optional().Default("localProvider").Values("localProvider", "AWSS3Provider", "B2Provider", "MinioS3Provider"),
		field.Bool("active").Default(false),
		field.Bool("running").Default(false),
		field.String("password").Optional().Sensitive(),
		field.String("maxSize").Optional(),
	}
}

// Edges of the Repository.
func (Repository) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("backupSchedule", BackupSchedule.Type).Unique().Annotations(entsql.Annotation{
			OnDelete: entsql.Cascade,
		}),
		edge.To("backupPolicy", BackupPolicy.Type).Unique().Annotations(entsql.Annotation{
			OnDelete: entsql.Cascade,
		}),
		edge.To("localProvider", LocalProvider.Type).Unique().Annotations(entsql.Annotation{
			OnDelete: entsql.Cascade,
		}),
		edge.To("B2Provider", B2Provider.Type).Unique().Annotations(entsql.Annotation{
			OnDelete: entsql.Cascade,
		}),
		edge.To("AWSS3Provider", AWSS3Provider.Type).Unique().Annotations(entsql.Annotation{
			OnDelete: entsql.Cascade,
		}),
		edge.To("MinioS3Provider", MinioS3Provider.Type).Unique().Annotations(entsql.Annotation{
			OnDelete: entsql.Cascade,
		}),
		edge.To("paths", Path.Type).Annotations(entsql.Annotation{
			OnDelete: entsql.Cascade,
		}),
	}
}
