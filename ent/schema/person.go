// nolint:gomnd
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Person is the table contain raw wiki text.
type Person struct {
	ent.Schema
}

// Annotations of the CharacterFields.
func (Person) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "chii_persons"},
	}
}

// Fields of the CharacterFields.
func (Person) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").StorageKey("prsn_id"),
		field.Text("name").MaxLen(255).StorageKey("prsn_name"),

		// '个人，公司，组合'
		field.Text("type").StorageKey(`prsn_type`).SchemaType(mysqlT("tinyint(4)")),
		field.Text("infobox").StorageKey(`prsn_infobox`).SchemaType(mysqlT("mediumtext")),

		field.Int("producer").StorageKey(`prsn_producer`).SchemaType(mysqlT("tinyint(1)")),
		field.Int("mangaka").StorageKey(`prsn_mangaka`).SchemaType(mysqlT("tinyint(1)")),
		field.Int("artist").StorageKey(`prsn_artist`).SchemaType(mysqlT("tinyint(1)")),
		field.Int("seiyu").StorageKey(`prsn_seiyu`).SchemaType(mysqlT("tinyint(1)")),
		field.Int("writer").StorageKey(`prsn_writer`).SchemaType(mysqlT("tinyint(4)")),
		field.Int("illustrator").StorageKey(`prsn_illustrator`).SchemaType(mysqlT("tinyint(4)")),
		field.Int("actor").StorageKey(`prsn_actor`).SchemaType(mysqlT("tinyint(1)")),

		field.Text("summary").StorageKey(`prsn_summary`).SchemaType(mysqlT("mediumtext")),
		field.Text("img").MaxLen(255).StorageKey(`prsn_img`).SchemaType(mysqlT("varchar(255)")),
		// 废弃字段
		field.Text("img_anidb").MaxLen(255).StorageKey(`prsn_img_anidb`).SchemaType(mysqlT("varchar(255)")),
		field.Int("comment").StorageKey(`prsn_comment`).SchemaType(mysqlT("mediumint(9)")),
		field.Int("collects").StorageKey(`prsn_collects`).SchemaType(mysqlT("mediumint(8)")),
		field.Int("dateline").StorageKey(`prsn_dateline`).SchemaType(mysqlT("int(10)")),
		field.Int("lastpost").StorageKey(`prsn_lastpost`).SchemaType(mysqlT("int(11)")),
		field.Int("lock").StorageKey(`prsn_lock`).SchemaType(mysqlT("tinyint(4)")),
		field.Text("anidb_id").StorageKey(`prsn_anidb_id`).SchemaType(mysqlT("mediumint(8)")),
		field.Int("ban").StorageKey(`prsn_ban`).SchemaType(mysqlT("tinyint(3)")),
		field.Int("redirect").StorageKey(`prsn_redirect`).SchemaType(mysqlT("int(10)")),
		field.Bool("nsfw").StorageKey(`prsn_nsfw`),
	}
}

func (Person) Indexes() []ent.Index {
	// entgo doesn't support multi columns primary key, so just leave it and never run migration
	return []ent.Index{
		index.Fields("type"),
		index.Fields(`producer`),
		index.Fields(`mangaka`),
		index.Fields(`artist`),
		index.Fields(`seiyu`),
		index.Fields(`writer`),
		index.Fields(`illustrator`),
		index.Fields(`lock`),
		index.Fields(`ban`),
		index.Fields(`actor`),
	}
}
