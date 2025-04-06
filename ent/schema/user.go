package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("username").Unique(),
		field.String("email").Unique(),
		field.Enum("role").Values("admin", "vendor", "user"),
		field.Bool("is_email_verified"),
		field.Time("created_at").Annotations(entsql.Default("CURRENT_TIMESTAMP")),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Nillable(),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username").Unique(),
		index.Fields("email").Unique(),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("profile", Profile.Type).Unique(),
	}
}
