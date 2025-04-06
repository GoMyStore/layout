package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Profile holds the schema definition for the Profile entity.
type Profile struct {
	ent.Schema
}

// Fields of the Profile.
func (Profile) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}).Unique(),
		field.Enum("gender").Values("male", "female").Nillable(),
		field.Enum("type").Values("user", "vendor"),
		field.String("avatar_url"),
		field.String("phone"),
		field.String("first_name"),
		field.String("last_name"),
		field.JSON("metadata", map[string]string{}),
		field.Time("created_at").Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Nillable(),
	}
}

func (Profile) Indexes() []ent.Index {
	return nil
}

// Edges of the Profile.
func (Profile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).Field("user_id").Unique().Required(),
	}
}
