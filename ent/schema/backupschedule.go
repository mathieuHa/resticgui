package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// BackupSchedule holds the schema definition for the BackupSchedule entity.
type BackupSchedule struct {
	ent.Schema
}

// Annotations of the BackupSchedule.
func (BackupSchedule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "BackupSchedule"},
	}
}

// Fields of the BackupSchedule.
func (BackupSchedule) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("choice").Values("cron", "interval", "monthly", "weekly").Default("cron"),
		field.Bool("everyDay").Default(false),
		field.Bool("everyWeek").Default(false),
		field.Bool("everyMonth").Default(false),
		field.Bool("everyMonday").Default(false),
		field.Bool("everyTuesday").Default(false),
		field.Bool("everyWednesday").Default(false),
		field.Bool("everyThursday").Default(false),
		field.Bool("everyFriday").Default(false),
		field.Bool("everySaturday").Default(false),
		field.Bool("everySunday").Default(false),
		field.String("hour").Optional().Default(""),
		field.String("day").Optional().Default(""),
		field.String("interval").Optional().Default(""),
		field.String("cronSyntax").Optional().Default(""),
		field.Int("cronID").Optional().Nillable(),
	}
}

// Edges of the BackupSchedule.
func (BackupSchedule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("repository", Repository.Type).
			Ref("backupSchedule").
			Unique().
			// We add the "Required" method to the builder
			// to make this edge required on entity creation.
			// i.e. Card cannot be created without its owner.
			Required(),
	}
}
