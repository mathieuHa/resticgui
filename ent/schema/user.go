package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username"),
		field.String("email"),
		field.String("password"),
		field.Time("LastConnectionDate").Optional(),
		field.Time("PasswordLastSet").Optional(),
		field.Int("PasswordFailCount").Default(0),
		field.Bool("mfa").Default(false),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
