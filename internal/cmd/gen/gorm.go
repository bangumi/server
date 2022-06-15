// SPDX-License-Identifier: AGPL-3.0-only
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>

/*
scripts to generate ORM struct from mysql server
*/

// nolint
package main

// disable lint in this package as it's only a generator

import (
	"reflect"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/uber-go/tally/v4"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/model"
)

var userIDTypeString = reflect.TypeOf(new(model.UIDType)).Elem().Name()
var personIDTypeString = reflect.TypeOf(new(model.PersonIDType)).Elem().Name()
var characterIDTypeString = reflect.TypeOf(new(model.CharacterIDType)).Elem().Name()
var episodeIDTypeString = reflect.TypeOf(new(model.EpisodeIDType)).Elem().Name()
var subjectIDTypeString = reflect.TypeOf(new(model.SubjectIDType)).Elem().Name()
var subjectTypeIDTypeString = reflect.TypeOf(new(model.SubjectType)).Elem().Name()
var episodeTypeTypeString = reflect.TypeOf(new(model.EpTypeType)).Elem().Name()

// generate code.
func main() {
	// specify the output directory (default: "./query")
	// ### if you want to query without context constrain, set mode gen.WithoutContext ###
	const dalBase = "./internal/dal"
	g := gen.NewGenerator(gen.Config{
		OutPath:       dalBase + "/query",
		OutFile:       dalBase + "/query/gen.go",
		ModelPkgPath:  dalBase + "/dao",
		FieldNullable: false,
		// if you want the nullable field generation property to be pointer type, set FieldNullable true
		/* FieldNullable: true,*/
		// if you want to generate index tags from database, set FieldWithIndexTag true
		FieldWithIndexTag: true,
		// if you want to generate type tags from database, set FieldWithTypeTag true
		FieldWithTypeTag: true,
		// if you need unit tests for query code, set WithUnitTest true
		// WithUnitTest: true,
	})

	// reuse the database connection in Project or create a connection here
	// if you want to use GenerateModel/GenerateModelAs, UseDB is necessary, otherwise it will panic
	c := config.NewAppConfig()
	conn, err := driver.NewMysqlConnectionPool(c)
	if err != nil {
		panic(err)
	}

	db, err := dal.NewDB(conn, c, tally.NoopScope, prometheus.NewRegistry())
	if err != nil {
		panic(err)
	}

	g.UseDB(db)
	dataMap := map[string]func(detailType string) (dataType string){
		// bool mapping
		"tinyint": func(detailType string) (dataType string) {
			if strings.HasPrefix(detailType, "tinyint(1)") {
				return "bool"
			}
			if strings.HasSuffix(detailType, "unsigned") {
				return "uint8"
			}
			return "int8"
		},

		"smallint": func(detailType string) (dataType string) {
			if strings.HasSuffix(detailType, "unsigned") {
				return "uint16"
			}
			return "int16"
		},

		"mediumint": func(detailType string) (dataType string) {
			if strings.HasSuffix(detailType, "unsigned") {
				return "uint32"
			}
			return "int32"
		},

		"int": func(detailType string) (dataType string) {
			if strings.HasSuffix(detailType, "unsigned") {
				return "uint32"
			}
			return "int32"
		},
	}
	g.WithDataTypeMap(dataMap)

	modelField := g.GenerateModelAs("chii_memberfields", "MemberField",
		gen.FieldType("uid", "uint32"),
	)

	modelMember := g.GenerateModelAs("chii_members", "Member",
		gen.FieldType("uid", "uint32"),
		gen.FieldRename("SIGN", "Sign"),
		gen.FieldType("regdate", "int64"),
		gen.FieldType("password_crypt", "[]byte"),
		gen.FieldType("groupid", "uint8"),
		gen.FieldRelate(field.HasOne, "Fields", modelField, &field.RelateConfig{
			GORMTag: "foreignKey:UID;references:UID",
		}))

	g.ApplyBasic(modelMember)

	g.ApplyBasic(g.GenerateModelAs("chii_os_web_sessions", "WebSession"))

	g.ApplyBasic(g.GenerateModelAs("chii_usergroup", "UserGroup",
		gen.FieldTrimPrefix("usr_grp_"),
		gen.FieldType("usr_grp_id", "uint8"),
		gen.FieldType("usr_grp_perm", "[]byte"),
	))

	var oauthApp = g.GenerateModelAs("chii_apps", "App",
		gen.FieldTrimPrefix("app_"),
		gen.FieldType("app_id", "uint32"),
		gen.FieldRename("app_desc", "description"),
		gen.FieldType("app_type", "uint8"),
		gen.FieldRename("app_lasttouch", "UpdatedAt"),
		gen.FieldRename("app_timestamp", "CreatedAt"),
		gen.FieldType("app_creator", userIDTypeString),
	)

	g.ApplyBasic(g.GenerateModelAs("chii_oauth_clients", "OAuthClient",
		gen.FieldType("app_id", "uint32"),
		gen.FieldRelate(field.BelongsTo, "App", oauthApp, &field.RelateConfig{
			GORMTag: "foreignKey:app_id;references:app_id",
		}),
	))

	g.ApplyBasic(oauthApp)

	g.ApplyBasic(g.GenerateModelAs("chii_oauth_access_tokens", "AccessToken",
		gen.FieldType("type", "uint8"),
		gen.FieldType("id", "uint32"),
		gen.FieldType("scope", "*string"),
		gen.FieldType("info", "[]byte"),
		gen.FieldRename("expires", "ExpiredAt"),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_subject_interests", "SubjectCollection",
		gen.FieldType("interest_subject_type", subjectTypeIDTypeString),
		gen.FieldType("interest_type", "uint8"),
		gen.FieldType("interest_private", "uint8"),
		gen.FieldTrimPrefix("interest_")))

	g.ApplyBasic(g.GenerateModelAs("chii_index", "Index",
		gen.FieldTrimPrefix("idx_"),
		gen.FieldType("idx_id", "uint32"),
		gen.FieldType("idx_uid", "uint32"),
		gen.FieldType("idx_collects", "uint32")))

	modelPersonField := g.GenerateModelAs("chii_person_fields", "PersonField",
		gen.FieldTrimPrefix("prsn_"),
		gen.FieldType("prsn_id", personIDTypeString),
		gen.FieldRename("prsn_id", "OwnerID"),
		// mysql year(4) has range 1901 to 2155, uint16 has range 0-65535.
		gen.FieldType("birth_year", "uint16"),
		gen.FieldRename("prsn_cat", "OwnerType"),
	)

	g.ApplyBasic(modelPersonField)

	modelPerson := g.GenerateModelAs("chii_persons", "Person",
		gen.FieldTrimPrefix("prsn_"),
		gen.FieldType("prsn_illustrator", "bool"),
		gen.FieldType("prsn_writer", "bool"),
		gen.FieldType("prsn_redirect", personIDTypeString),
		gen.FieldRelate(field.HasOne, "Fields", modelPersonField, &field.RelateConfig{
			GORMTag: "foreignKey:prsn_id;polymorphic:Owner;polymorphicValue:prsn",
		}),
	)
	g.ApplyBasic(modelPerson)

	modelCharacter := g.GenerateModelAs("chii_characters", "Character",
		gen.FieldTrimPrefix("crt_"),
		gen.FieldType("crt_id", characterIDTypeString),
		gen.FieldType("crt_redirect", personIDTypeString),
		gen.FieldRelate(field.HasOne, "Fields", modelPersonField, &field.RelateConfig{
			GORMTag: "foreignKey:crt_id;polymorphic:Owner;polymorphicValue:crt",
		}),
	)

	g.ApplyBasic(modelCharacter)

	modelSubjectFields := g.GenerateModelAs("chii_subject_fields", "SubjectField",
		// gen.FieldTrimPrefix("field_"),
		gen.FieldTrimPrefix("field_"),
		gen.FieldType("field_airtime", "uint8"),
		gen.FieldType("field_week_day", "int8"),
		gen.FieldType("field_redirect", subjectIDTypeString),
		gen.FieldType("field_tags", "[]byte"),
		// gen.FieldType("field_date","string"),
	)

	g.ApplyBasic(modelSubjectFields)

	modelSubject := g.GenerateModelAs("chii_subjects", "Subject",
		// gen.FieldTrimPrefix("field_"),
		gen.FieldTrimPrefix("subject_"),
		gen.FieldRename("subject_name_cn", "NameCN"),
		gen.FieldRename("field_summary", "summary"),
		gen.FieldRename("field_eps", "eps"),
		gen.FieldRename("field_volumes", "volumes"),
		gen.FieldRename("field_infobox", "infobox"),
		gen.FieldType("subject_id", subjectIDTypeString),
		gen.FieldType("subject_ban", "uint8"),
		gen.FieldType("subject_type_id", subjectTypeIDTypeString),
		gen.FieldType("subject_airtime", "uint8"),
		gen.FieldRelate(field.HasOne, "Fields", modelSubjectFields, &field.RelateConfig{
			GORMTag: "foreignKey:subject_id;references:field_sid",
		}),
	)
	g.ApplyBasic(modelSubject)

	g.ApplyBasic(g.GenerateModelAs("chii_episodes", "Episode",
		// gen.FieldTrimPrefix("field_"),
		gen.FieldTrimPrefix("ep_"),
		gen.FieldType("ep_id", episodeIDTypeString),
		gen.FieldType("ep_type", episodeTypeTypeString),
		gen.FieldRelate(field.BelongsTo, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: "foreignKey:ep_subject_id;references:subject_id",
		}),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_subject_relations", "SubjectRelation",
		// gen.FieldTrimPrefix("field_"),
		gen.FieldTrimPrefix("rlt_"),
		gen.FieldType("rlt_related_subject_id", subjectIDTypeString),
		gen.FieldType("rlt_subject_id", subjectIDTypeString),
		gen.FieldType("rlt_subject_type_id", subjectTypeIDTypeString),
		gen.FieldType("rlt_related_subject_type_id", subjectTypeIDTypeString),
		gen.FieldRelate(field.HasOne, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: "foreignKey:rlt_related_subject_id;references:subject_id",
		}),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_subject_revisions", "SubjectRevision",
		gen.FieldTrimPrefix("rev_"),
		gen.FieldRename("rev_name_cn", "NameCN"),
		gen.FieldRelate(field.BelongsTo, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: "foreignKey:rev_subject_id;references:subject_id",
		}),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_crt_cast_index", "Cast",
		gen.FieldRename("prsn_id", "PersonID"),
		gen.FieldRename("crt_id", "CharacterID"),
		gen.FieldRelate(field.HasOne, "Character", modelCharacter, &field.RelateConfig{
			GORMTag: "foreignKey:crt_id;references:crt_id",
		}),
		gen.FieldRelate(field.HasOne, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: "foreignKey:subject_id;references:subject_id",
		}),
		gen.FieldRelate(field.HasOne, "Person", modelPerson, &field.RelateConfig{
			GORMTag: "foreignKey:prsn_id;references:prsn_id",
		}),
	))

	g.ApplyBasic(
		g.GenerateModelAs("chii_crt_subject_index", "CharacterSubjects",
			gen.FieldRename("crt_id", "CharacterID"),
			gen.FieldType("subject_id", subjectIDTypeString),
			gen.FieldType("subject_type_id", subjectTypeIDTypeString),
			gen.FieldRelate(field.HasOne, "Character", modelCharacter, &field.RelateConfig{
				GORMTag: "foreignKey:crt_id;references:crt_id",
			}),
			gen.FieldRelate(field.HasOne, "Subject", modelSubject, &field.RelateConfig{
				GORMTag: "foreignKey:subject_id;references:subject_id",
			}),
		),
	)

	g.ApplyBasic(g.GenerateModelAs("chii_person_cs_index", "PersonSubjects",
		gen.FieldRename("prsn_id", "person_id"),
		gen.FieldType("subject_id", subjectIDTypeString),
		gen.FieldType("subject_type_id", subjectTypeIDTypeString),
		gen.FieldRelate(field.HasOne, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: "foreignKey:subject_id;references:subject_id",
		}),
		gen.FieldRelate(field.HasOne, "Person", modelPerson, &field.RelateConfig{
			GORMTag: "foreignKey:prsn_id;references:prsn_id",
		}),
	),
	)

	g.ApplyBasic(g.GenerateModelAs("chii_index_related", "IndexSubject",
		gen.FieldTrimPrefix("idx_rlt_"),
		gen.FieldType("idx_rlt_type", "uint8"),
		gen.FieldRelate(field.BelongsTo, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: "foreignKey:idx_rlt_sid;references:subject_id",
		}),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_rev_text", "RevisionText",
		gen.FieldTrimPrefix("rev_")),
	)

	g.ApplyBasic(g.GenerateModelAs("chii_rev_history", "RevisionHistory",
		gen.FieldTrimPrefix("rev_"),
		gen.FieldRename("rev_edit_summary", "Summary"),
		gen.FieldRename("rev_dateline", "CreatedAt"),
		gen.FieldRename("rev_creator", "CreatorID"),
	))

	// execute the action of code generation
	g.Execute()
}
