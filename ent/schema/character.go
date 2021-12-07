// Package schema

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

func mysqlT(t string) map[string]string {
	return map[string]string{dialect.MySQL: t}
}

type PersonCsIndex struct {
	ent.Schema
}

func (PersonCsIndex) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "chii_person_cs_index"},
	}
}

func (PersonCsIndex) Fields() []ent.Field {
	return []ent.Field{
		// it's always `prsn` in this table
		field.Enum("prsn_type").Values("prsn", "crt").StructTag(`json:"-"`),
		field.Uint8("id").StorageKey("prsn_id").SchemaType(mysqlT("mediumint(9)")),
		field.Int("prsn_position").SchemaType(mysqlT("smallint(5)")),
		field.Int("subject_id").SchemaType(mysqlT("mediumint(9)")),
		field.Int("subject_type_id").SchemaType(mysqlT("tinyint(4)")),
		field.Text("summary").SchemaType(mysqlT("mediumtext")),
		field.Text("appear_eps").StorageKey("prsn_appear_eps").SchemaType(mysqlT("mediumtext")),
	}
}

func (PersonCsIndex) Indexes() []ent.Index {
	// entgo doesn't support multi columns primary key, so just leave it and never run migration
	return []ent.Index{
		index.Fields("prsn_type", "id", "subject_id", "prsn_position").Unique(), // primary key in schema
		index.Fields("id"),
		index.Fields("prsn_position"),
		index.Fields("subject_type_id"),
	}
}

// CharacterFields holds the schema definition for the CharacterFields entity.
type CharacterFields struct {
	ent.Schema
}

// Annotations of the CharacterFields.
func (CharacterFields) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "chii_person_fields"},
	}
}

// Fields of the CharacterFields.
func (CharacterFields) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("prsn_cat").Values("prsn", "crt").StructTag(`json:"-"`),
		field.Uint8("id").StorageKey("prsn_id"),
		field.Int("gender").SchemaType(mysqlT("tinyint(4)")),
		field.Int("bloodtype").SchemaType(mysqlT("tinyint(4)")),
		field.Int("birth_year").SchemaType(mysqlT("year(4)")),
		field.Int("birth_mon").SchemaType(mysqlT("tinyint(2)")),
		field.Int("birth_day").SchemaType(mysqlT("tinyint(2)")),
	}
}

func (CharacterFields) Indexes() []ent.Index {
	// entgo doesn't support multi columns primary key, so just leave it and never run migration
	return []ent.Index{
		index.Fields("id"),
		index.Fields("prsn_cat", "id").Unique(),
	}
}
