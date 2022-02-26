package schema

import (
	"time"

	"entgo.io/ent"
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
		field.String("name").Unique().NotEmpty().MinLen(3),
		field.Time("createdAt").
			Default(time.Now).Immutable(),
		field.Enum("provider").Values("local", "AWSS3", "B2", "MinioS3").Default("local"),
		field.Bool("active").Default(false),
		field.String("password").Optional().MinLen(4).Sensitive(),
		field.String("maxSize").Optional(),
	}
}

// Edges of the Repository.
func (Repository) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("backupschedule", BackupSchedule.Type).Unique(),
		edge.To("backuppolicy", BackupPolicy.Type).Unique(),
		edge.To("localprovider", LocalProvider.Type).Unique(),
		edge.To("b2provider", B2Provider.Type).Unique(),
		edge.To("awss3provider", AWSS3Provider.Type).Unique(),
		edge.To("minios3provider", MinioS3Provider.Type).Unique(),
		edge.To("path", Path.Type),
	}
}
