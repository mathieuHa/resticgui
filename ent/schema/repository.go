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
		field.Bool("active").Default(false).StructTag(`json:"active"`),
		field.Bool("running").Default(false).StructTag(`json:"running"`),
		field.Bool("initiated").Default(false).StructTag(`json:"initiated"`),
		field.Bool("valid").Default(false).StructTag(`json:"valid"`),
		field.Enum("version").Optional().Default("2").Values("1", "2"),
		// Password should be set as sensitive, but should be accessible for backups
		// field.String("password").Optional().Sensitive(),
		field.String("password").Optional(),
		field.String("maxSize").Optional(),
		field.String("repositorySize").Optional(),
		field.String("repositoryFileCount").Optional(),
		field.String("lastSnapshotSize").Optional(),
		field.String("lastSnapshotDuration").Optional(),
		field.Int("snapshotCount").Optional(),
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
		edge.To("logs", Logs.Type).Annotations(entsql.Annotation{
			OnDelete: entsql.Cascade,
		}),
	}
}
