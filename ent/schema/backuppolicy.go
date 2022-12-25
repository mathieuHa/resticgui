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
		field.Uint("nbKeepLast").Default(10).StructTag(`json:"nbKeepLast"`),
		field.Uint("nbKeepHours").Default(1).StructTag(`json:"nbKeepHours"`),
		field.Uint("nbKeepDays").Default(1).StructTag(`json:"nbKeepDays"`),
		field.Uint("nbKeepWeeks").Default(1).StructTag(`json:"nbKeepWeeks"`),
		field.Uint("nbKeepMonths").Default(1).StructTag(`json:"nbKeepMonths"`),
		field.Bool("keepAll").Default(true).StructTag(`json:"keepAll"`),
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
