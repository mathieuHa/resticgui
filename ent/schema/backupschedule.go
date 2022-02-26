package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// BackupSchedule holds the schema definition for the BackupSchedule entity.
type BackupSchedule struct {
	ent.Schema
}

// Fields of the BackupSchedule.
func (BackupSchedule) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("choice").Values("cron", "interval", "monthly", "weekly").Default("cron"),
	}
}

// Edges of the BackupSchedule.
func (BackupSchedule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("repository", Repository.Type).
			Ref("backupschedule").
			Unique().
			// We add the "Required" method to the builder
			// to make this edge required on entity creation.
			// i.e. Card cannot be created without its owner.
			Required(),
	}
}
