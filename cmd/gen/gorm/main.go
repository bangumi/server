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

NOTICE:
	Don't use `UpdatedAt` and `CreatedAt` as field name, gorm may change these fields unexpectedly.
	Use `UpdatedTime` and `CreatedTime` instead.
*/

//nolint:all
package main

import (
	"path/filepath"
	"strings"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/dal"
	"github.com/bangumi/server/internal/pkg/driver"
)

var userIDTypeString = "uint32"         // reflect.TypeOf(new(model.UserID)).Elem().Name()
var personIDTypeString = "uint32"       // reflect.TypeOf(new(model.PersonID)).Elem().Name()
var characterIDTypeString = "uint32"    // reflect.TypeOf(new(model.CharacterID)).Elem().Name()
var episodeIDTypeString = "uint32"      // reflect.TypeOf(new(model.EpisodeID)).Elem().Name()
var subjectIDTypeString = "uint32"      // reflect.TypeOf(new(model.SubjectID)).Elem().Name()
var subjectTypeIDTypeString = "uint8"   // reflect.TypeOf(new(model.SubjectType)).Elem().Name()
var episodeTypeTypeString = "uint8"     // reflect.TypeOf(new(episode.Type)).Elem().Name()
var notificationIDTypeString = "uint32" // reflect.TypeOf(new(model.NotificationID)).Elem().Name()
var notificationFieldIDTypeString = "uint32"
var notificationTypeTypeString = "uint8"
var notificationStatusTypeString = "uint8"
var privateMessageIDTypeString = "uint32"
var privateMessageFolderTypeTypeString = "string"

func DeprecatedFiled(s string) gen.ModelOpt {
	return gen.FieldComment(s, "Deprecated")
}

const createdTime = "CreatedTime"
const updateTime = "UpdatedTime"

// generate code.
func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:      filepath.Clean("./dal/query/"),
		OutFile:      "gen.go",
		ModelPkgPath: "dao",

		WithUnitTest: false,
		// if you want the nullable field generation property to be pointer type, set FieldNullable true
		FieldNullable:     false,
		FieldCoverable:    false,
		FieldSignable:     false,
		FieldWithIndexTag: false,
		// if you want to generate type tags from database, set FieldWithTypeTag true
		FieldWithTypeTag: true,
		Mode:             0,
	})

	g.WithImportPkgPath(
		"github.com/bangumi/server/internal/model",
		"github.com/bangumi/server/dal/utiltype",
		"gorm.io/plugin/soft_delete",
	)

	g.WithJSONTagNameStrategy(func(_ string) string {
		return ""
	})

	// reuse the database connection in Project or create a connection here
	// if you want to use GenerateModel/GenerateModelAs, UseDB is necessary, otherwise it will panic
	c, err := config.NewAppConfig()
	if err != nil {
		panic("failed to read config: " + err.Error())
	}

	conn, err := driver.NewMysqlDriver(c)
	if err != nil {
		panic(err)
	}

	db, err := dal.NewGormDB(conn, c)
	if err != nil {
		panic(err)
	}

	g.UseDB(db)
	dataMap := map[string]func(detailType gorm.ColumnType) (dataType string){
		// bool mapping
		"tinyint": func(t gorm.ColumnType) (dataType string) {
			dt, ok := t.ColumnType()
			if !ok {
				panic("failed to get column type")
			}
			if strings.HasPrefix(dt, "tinyint(1)") {
				return "bool"
			}
			if strings.HasSuffix(dt, "unsigned") {
				return "uint8"
			}
			return "int8"
		},

		"smallint": func(t gorm.ColumnType) (dataType string) {
			dt, ok := t.ColumnType()
			if !ok {
				panic("failed to get column type")
			}

			if strings.HasSuffix(dt, "unsigned") {
				return "uint16"
			}
			return "int16"
		},

		"mediumint": func(t gorm.ColumnType) (dataType string) {
			dt, ok := t.ColumnType()
			if !ok {
				panic("failed to get column type")
			}

			if strings.HasSuffix(dt, "unsigned") {
				return "uint32"
			}
			return "int32"
		},

		"int": func(t gorm.ColumnType) (dataType string) {
			dt, ok := t.ColumnType()
			if !ok {
				panic("failed to get column type")
			}

			if strings.HasSuffix(dt, "unsigned") {
				return "uint32"
			}
			return "int32"
		},
	}
	g.WithDataTypeMap(dataMap)

	modelField := g.GenerateModelAs("chii_memberfields", "MemberField",
		gen.FieldType("uid", userIDTypeString),
		gen.FieldType("privacy", "[]byte"),
		gen.FieldIgnore("index_sort"),
		gen.FieldIgnore("user_agent"),
		gen.FieldIgnore("ignorepm"),
		gen.FieldIgnore("groupterms"),
		gen.FieldIgnore("authstr"),
		gen.FieldIgnoreReg("^(homepage|reg_source|invite_num|email_verified|reset_password_dateline|reset_password_token)$"),
		gen.FieldIgnoreReg("^(reset_password_force|email_verify_dateline|email_verify_token|email_verify_score)$"),
	)

	modelMember := g.GenerateModelAs("chii_members", "Member",
		gen.FieldRename("uid", "ID"),
		// gen.FieldIgnore("password_crypt"),
		gen.FieldIgnore("secques"),
		gen.FieldIgnore("gender"),
		gen.FieldIgnore("adminid"),
		gen.FieldIgnore("regip"),
		gen.FieldIgnore("lastip"),

		// gen.FieldIgnore("email"),
		gen.FieldIgnore("bday"),
		gen.FieldIgnore("styleid"),
		gen.FieldIgnore("newsletter"),
		gen.FieldIgnore("ukagaka_settings"),
		gen.FieldIgnore("username_lock"),
		gen.FieldIgnore("invited"),
		gen.FieldIgnore("img_chart"),

		gen.FieldType("uid", userIDTypeString),
		gen.FieldType("sign", "utiltype.HTMLEscapedString"),
		gen.FieldType("regdate", "int64"),
		gen.FieldType("password_crypt", "[]byte"),
		gen.FieldType("groupid", "uint8"),
		gen.FieldRelate(field.HasOne, "Fields", modelField, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"uid"}, "references": []string{"uid"}},
		}))

	g.ApplyBasic(modelMember)

	g.ApplyBasic(g.GenerateModelAs("chii_os_web_sessions", "WebSession",
		gen.FieldType("user_id", userIDTypeString),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_usergroup", "UserGroup",
		gen.FieldTrimPrefix("usr_grp_"),
		gen.FieldType("usr_grp_id", "uint8"),
		gen.FieldType("usr_grp_perm", "[]byte"),
	))

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
		gen.FieldType("interest_uid", userIDTypeString),
		gen.FieldType("interest_comment", "utiltype.HTMLEscapedString"),
		gen.FieldRename("interest_uid", "UserID"),

		gen.FieldRename("interest_create_ip", "CreateIP"),
		gen.FieldRename("interest_lasttouch_ip", "LastUpdateIP"),
		gen.FieldRename("interest_collect_dateline", "DoneTime"),
		gen.FieldRename("interest_doing_dateline", "DoingTime"),
		gen.FieldRename("interest_on_hold_dateline", "OnHoldTime"),
		gen.FieldRename("interest_dropped_dateline", "droppedTime"),
		gen.FieldRename("interest_wish_dateline", "WishTime"),
		gen.FieldType("interest_subject_id", subjectIDTypeString),
		gen.FieldType("interest_private", "uint8"),
		gen.FieldRename("interest_lasttouch", "UpdatedTime"),
		gen.FieldTrimPrefix("interest_"),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_person_collects", "PersonCollect",
		gen.FieldTrimPrefix("prsn_clt_"),
		gen.FieldType("prsn_clt_id", "uint32"),
		gen.FieldType("prsn_clt_cat", "string"),
		gen.FieldType("prsn_clt_uid", userIDTypeString),
		gen.FieldType("prsn_clt_mid", "uint32"),
		gen.FieldType("prsn_clt_dateline", "uint32"),
		gen.FieldRename("prsn_clt_cat", "Category"),
		gen.FieldRename("prsn_clt_uid", "UserID"),
		gen.FieldRename("prsn_clt_mid", "TargetID"),
		gen.FieldRename("prsn_clt_dateline", createdTime),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_index", "Index",
		gen.FieldTrimPrefix("idx_"),
		gen.FieldType("idx_id", "uint32"),
		gen.FieldType("idx_uid", userIDTypeString),
		gen.FieldType("idx_collects", "uint32"),
		// 变量重命名
		gen.FieldRename("idx_uid", "CreatorID"),
		gen.FieldRename("idx_dateline", "CreatedTime"),
		gen.FieldRename("idx_lasttouch", "UpdatedTime"),
		gen.FieldRename("idx_replies", "ReplyCount"),
		gen.FieldRename("idx_collects", "CollectCount"),
		gen.FieldRename("idx_subject_total", "SubjectCount"),

		gen.FieldType("idx_ban", "uint8"),
		gen.FieldRename("idx_ban", "Privacy"),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_index_collects", "IndexCollect",
		gen.FieldTrimPrefix("idx_"),
		gen.FieldType("idx_clt_id", "uint32"),
		gen.FieldType("idx_clt_mid", "uint32"),
		gen.FieldType("idx_clt_uid", userIDTypeString),

		gen.FieldRename("idx_clt_uid", "UserID"),
		gen.FieldRename("idx_clt_mid", "IndexID"),
		gen.FieldRename("idx_clt_dateline", createdTime),
	))

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
		DeprecatedFiled("prsn_img_anidb"),
		DeprecatedFiled("prsn_anidb_id"),
		gen.FieldType("prsn_id", personIDTypeString),
		gen.FieldType("prsn_illustrator", "bool"),
		gen.FieldType("prsn_writer", "bool"),
		gen.FieldType("prsn_redirect", personIDTypeString),
		gen.FieldRelate(field.HasOne, "Fields", modelPersonField, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"prsn_id"}, "polymorphic": []string{"Owner"}, "polymorphicValue": []string{"prsn"}},
		}),
	)
	g.ApplyBasic(modelPerson)

	modelCharacter := g.GenerateModelAs("chii_characters", "Character",
		gen.FieldTrimPrefix("crt_"),
		DeprecatedFiled("crt_img_anidb"),
		DeprecatedFiled("crt_anidb_id"),
		gen.FieldType("crt_id", characterIDTypeString),
		gen.FieldType("crt_redirect", characterIDTypeString),
		gen.FieldRelate(field.HasOne, "Fields", modelPersonField, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"crt_id"}, "polymorphic": []string{"Owner"}, "polymorphicValue": []string{"crt"}},
		}),
	)

	g.ApplyBasic(modelCharacter)

	modelSubjectFields := g.GenerateModelAs("chii_subject_fields", "SubjectField",
		// gen.FieldTrimPrefix("field_"),
		gen.FieldTrimPrefix("field_"),
		gen.FieldType("field_sid", subjectIDTypeString),
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
		gen.FieldRename("subject_collect", "Done"),
		gen.FieldRename("field_infobox", "infobox"),
		gen.FieldType("subject_id", subjectIDTypeString),
		gen.FieldType("subject_name", "utiltype.HTMLEscapedString"),
		gen.FieldType("field_infobox", "utiltype.HTMLEscapedString"),
		gen.FieldType("subject_name_cn", "utiltype.HTMLEscapedString"),
		gen.FieldType("subject_ban", "uint8"),
		gen.FieldType("subject_type_id", subjectTypeIDTypeString),
		gen.FieldType("subject_airtime", "uint8"),
		gen.FieldRelate(field.HasOne, "Fields", modelSubjectFields, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"subject_id"}, "references": []string{"field_sid"}},
		}),
	)
	g.ApplyBasic(modelSubject)

	g.ApplyBasic(g.GenerateModelAs("chii_episodes", "Episode",
		// gen.FieldTrimPrefix("field_"),
		gen.FieldTrimPrefix("ep_"),
		gen.FieldType("ep_id", episodeIDTypeString),
		gen.FieldType("ep_type", episodeTypeTypeString),
		gen.FieldType("ep_subject_id", subjectIDTypeString),
		gen.FieldRelate(field.BelongsTo, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"ep_subject_id"}, "references": []string{"subject_id"}},
		}),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_ep_status", "EpCollection",
		gen.FieldTrimPrefix("ep_stt"),
		gen.FieldType("ep_stt_sid", subjectIDTypeString),
		gen.FieldRename("ep_stt_sid", "SubjectID"),
		gen.FieldType("ep_stt_uid", userIDTypeString),
		gen.FieldRename("ep_stt_uid", "userID"),
		gen.FieldRename("ep_stt_lasttouch", "UpdatedTime"),
		gen.FieldType("ep_stt_status", "[]byte"),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_subject_relations", "SubjectRelation",
		// gen.FieldTrimPrefix("field_"),
		gen.FieldTrimPrefix("rlt_"),
		gen.FieldType("rlt_related_subject_id", subjectIDTypeString),
		gen.FieldType("rlt_subject_id", subjectIDTypeString),
		gen.FieldType("rlt_subject_type_id", subjectTypeIDTypeString),
		gen.FieldType("rlt_related_subject_type_id", subjectTypeIDTypeString),
		gen.FieldRelate(field.HasOne, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"rlt_related_subject_id"}, "references": []string{"subject_id"}},
		}),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_subject_revisions", "SubjectRevision",
		gen.FieldTrimPrefix("rev_"),
		gen.FieldRename("rev_name_cn", "NameCN"),
		gen.FieldRename("rev_creator", "CreatorID"),
		gen.FieldType("rev_creator", userIDTypeString),
		gen.FieldType("rev_subject_id", subjectIDTypeString),
		gen.FieldRelate(field.BelongsTo, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"rev_subject_id"}, "references": []string{"subject_id"}},
		}),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_crt_cast_index", "Cast",
		gen.FieldRename("prsn_id", "PersonID"),
		gen.FieldRename("crt_id", "CharacterID"),
		gen.FieldType("crt_id", characterIDTypeString),
		gen.FieldType("prsn_id", personIDTypeString),
		gen.FieldType("subject_id", subjectIDTypeString),
		gen.FieldRelate(field.HasOne, "Character", modelCharacter, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"crt_id"}, "references": []string{"crt_id"}},
		}),
		gen.FieldRelate(field.HasOne, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"subject_id"}, "references": []string{"subject_id"}},
		}),
		gen.FieldRelate(field.HasOne, "Person", modelPerson, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"prsn_id"}, "references": []string{"prsn_id"}},
		}),
	))

	g.ApplyBasic(
		g.GenerateModelAs("chii_crt_subject_index", "CharacterSubjects",
			gen.FieldIgnore("ctr_appear_eps"),
			gen.FieldRename("crt_id", "CharacterID"),
			gen.FieldType("crt_id", characterIDTypeString),
			gen.FieldType("subject_id", subjectIDTypeString),
			gen.FieldType("subject_type_id", subjectTypeIDTypeString),
			gen.FieldRelate(field.HasOne, "Character", modelCharacter, &field.RelateConfig{
				GORMTag: field.GormTag{"foreignKey": []string{"crt_id"}, "references": []string{"crt_id"}},
			}),
			gen.FieldRelate(field.HasOne, "Subject", modelSubject, &field.RelateConfig{
				GORMTag: field.GormTag{"foreignKey": []string{"subject_id"}, "references": []string{"subject_id"}},
			}),
		),
	)

	g.ApplyBasic(g.GenerateModelAs("chii_person_cs_index", "PersonSubjects",
		gen.FieldRename("prsn_id", "person_id"),
		gen.FieldType("prsn_id", personIDTypeString),
		gen.FieldType("subject_id", subjectIDTypeString),
		gen.FieldType("subject_type_id", subjectTypeIDTypeString),
		gen.FieldRelate(field.HasOne, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"subject_id"}, "references": []string{"subject_id"}},
		}),
		gen.FieldRelate(field.HasOne, "Person", modelPerson, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"prsn_id"}, "references": []string{"prsn_id"}},
		}),
	),
	)

	g.ApplyBasic(g.GenerateModelAs("chii_person_relations", "PersonRelation",
		gen.FieldTrimPrefix("rlt_"),
		gen.FieldRename("prsn_id", "PersonID"),
		gen.FieldRename("prsn_type", "PersonType"),
		gen.FieldRename("rlt_prsn_id", "RelatedPersonID"),
		gen.FieldRename("rlt_prsn_type", "RelatedPersonType"),
		gen.FieldRename("rlt_type", "RelationType"),
		gen.FieldType("prsn_id", personIDTypeString),
		gen.FieldType("rlt_spoiler", "bool"),
		gen.FieldType("rlt_ended", "bool"),
		gen.FieldType("rlt_vice_versa", "bool"),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_index_related", "IndexSubject",
		gen.FieldTrimPrefix("idx_rlt_"),
		gen.FieldType("idx_rlt_type", "uint8"),
		gen.FieldRelate(field.BelongsTo, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"idx_rlt_sid"}, "references": []string{"subject_id"}},
		}),
		// 变量重命名
		gen.FieldRename("idx_rlt_rid", "IndexID"),
		gen.FieldRename("idx_rlt_sid", "SubjectID"),
		gen.FieldRename("idx_rlt_type", "SubjectType"),
		gen.FieldRename("idx_rlt_dateline", "CreatedTime"),
		gen.FieldType("idx_rlt_ban", "soft_delete.DeletedAt"),
		gen.FieldRename("idx_rlt_ban", "Deleted"),
		gen.FieldGORMTag("idx_rlt_ban", func(tag field.GormTag) field.GormTag {
			tag["softDelete"] = []string{"flag"}
			return tag
		}),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_rev_text", "RevisionText",
		gen.FieldTrimPrefix("rev_")),
	)

	g.ApplyBasic(g.GenerateModelAs("chii_rev_history", "RevisionHistory",
		gen.FieldTrimPrefix("rev_"),
		gen.FieldRename("rev_edit_summary", "Summary"),
		gen.FieldRename("rev_dateline", createdTime),
		gen.FieldRename("rev_creator", "CreatorID"),
		gen.FieldType("rev_creator", userIDTypeString),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_friends", "Friend",
		gen.FieldTrimPrefix("frd_"),
		gen.FieldType("frd_uid", userIDTypeString),
		gen.FieldRename("frd_uid", "UserID"),
		gen.FieldType("frd_fid", userIDTypeString),
		gen.FieldRename("frd_fid", "FriendID"),
		gen.FieldRename("frd_dateline", createdTime),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_notify", "Notification",
		gen.FieldTrimPrefix("nt_"),
		gen.FieldType("nt_uid", userIDTypeString),
		gen.FieldType("nt_from_uid", userIDTypeString),
		gen.FieldType("nt_id", notificationIDTypeString),
		gen.FieldType("nt_mid", notificationFieldIDTypeString),
		gen.FieldType("nt_type", notificationTypeTypeString),
		gen.FieldType("nt_status", notificationStatusTypeString),
		gen.FieldRename("nt_uid", "ReceiverID"),
		gen.FieldRename("nt_from_uid", "SenderID"),
		gen.FieldRename("nt_mid", "FieldID"),
		gen.FieldRename("nt_dateline", createdTime),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_notify_field", "NotificationField",
		gen.FieldTrimPrefix("ntf_"),
		gen.FieldType("ntf_id", notificationFieldIDTypeString),
		gen.FieldRename("ntf_rid", "RelatedID"),
		gen.FieldRename("ntf_hash", "RelatedType"),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_pms", "PrivateMessage",
		gen.FieldTrimPrefix("msg_"),
		gen.FieldType("msg_id", privateMessageIDTypeString),
		gen.FieldType("msg_folder", privateMessageFolderTypeTypeString),
		gen.FieldType("msg_sid", userIDTypeString),
		gen.FieldType("msg_rid", userIDTypeString),
		gen.FieldType("msg_related_main", privateMessageIDTypeString),
		gen.FieldType("msg_related", privateMessageIDTypeString),
		gen.FieldRename("msg_dateline", createdTime),
		gen.FieldRename("msg_message", "Content"),
		gen.FieldRename("msg_sid", "SenderID"),
		gen.FieldRename("msg_rid", "ReceiverID"),
		gen.FieldRename("msg_related_main", "MainMessageID"),
		gen.FieldRename("msg_related", "RelatedMessageID"),
		gen.FieldRename("msg_sdeleted", "DeletedBySender"),
		gen.FieldRename("msg_rdeleted", "DeletedByReceiver"),
	))

	modelTagIndex := g.GenerateModelAs("chii_tag_neue_index", "TagIndex",
		gen.FieldTrimPrefix("tag_"),
		gen.FieldRename("tag_dateline", createdTime),
		gen.FieldRename("tag_lasttouch", updateTime),
		gen.FieldType("tag_type", "uint8"),
	)

	g.ApplyBasic(modelTagIndex)

	g.ApplyBasic(g.GenerateModelAs("chii_tag_neue_list", "TagList",
		gen.FieldTrimPrefix("tlt_"),
		gen.FieldRename("tlt_dateline", createdTime),

		gen.FieldRelate(field.HasOne, "Tag", modelTagIndex, &field.RelateConfig{
			GORMTag: field.GormTag{"foreignKey": []string{"tag_id"}, "references": []string{"tlt_tid"}},
		}),
	))

	// execute the action of code generation
	g.Execute()
}
