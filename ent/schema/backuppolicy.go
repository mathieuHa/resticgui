package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// BackupPolicy holds the schema definition for the BackupPolicy entity.
type BackupPolicy struct {
	ent.Schema
}

// Annotations of the BackupPolicy.
func (BackupPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "BackupPolicy"},
	}
}

// Fields of the BackupPolicy.
func (BackupPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("nbLastBackup").Default(10),
		field.String("name"),
		field.Bool("oneByHour"),
		field.Bool("oneByDay"),
		field.Bool("oneByMonth"),
		field.Bool("keepAll").Default(true),
	}
}

// Edges of the BackupPolicy.
func (BackupPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("repository", Repository.Type).
			Ref("backupPolicy").
			Unique().
			// We add the "Required" method to the builder
			// to make this edge required on entity creation.
			// i.e. Card cannot be created without its owner.
			Required(),
	}
}
