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
	"reflect"
	"strings"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/dal"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/driver"
)

var userIDTypeString = "model.UserID"           // reflect.TypeOf(new(model.UserID)).Elem().Name()
var personIDTypeString = "model.PersonID"       // reflect.TypeOf(new(model.PersonID)).Elem().Name()
var characterIDTypeString = "model.CharacterID" // reflect.TypeOf(new(model.CharacterID)).Elem().Name()
var episodeIDTypeString = "model.EpisodeID"     // reflect.TypeOf(new(model.EpisodeID)).Elem().Name()
var subjectIDTypeString = "model.SubjectID"     // reflect.TypeOf(new(model.SubjectID)).Elem().Name()
var groupIDTypeString = "model." + reflect.TypeOf(new(model.GroupID)).Elem().Name()
var timelineIDTypeString = "model." + reflect.TypeOf(new(model.TimeLineID)).Elem().Name()
var timelineCatTypeString = reflect.TypeOf(new(model.TimeLineCat)).Elem().Name()
var subjectTypeIDTypeString = reflect.TypeOf(new(model.SubjectType)).Elem().Name()
var episodeTypeTypeString = reflect.TypeOf(new(episode.Type)).Elem().Name()
var notificationIDTypeString = "model." + reflect.TypeOf(new(model.NotificationID)).Elem().Name()
var notificationFieldIDTypeString = "uint32"
var notificationTypeTypeString = "uint8"
var notificationStatusTypeString = "uint8"
var privateMessageIDTypeString = "model." + reflect.TypeOf(new(model.PrivateMessageID)).Elem().Name()
var privateMessageFolderTypeTypeString = "string"

func DeprecatedFiled(s string) gen.ModelOpt {
	return gen.FieldComment(s, "Deprecated")
}

const createdTime = "CreatedTime"

// generate code.
func main() {
	// specify the output directory (default: "./query")
	// ### if you want to query without context constrain, set mode gen.WithoutContext ###
	const dalBase = "./dal"
	g := gen.NewGenerator(gen.Config{
		OutPath:      dalBase + "/query",
		OutFile:      dalBase + "/query/gen.go",
		ModelPkgPath: dalBase + "/dao",
		// if you want the nullable field generation property to be pointer type, set FieldNullable true
		FieldNullable: false,
		// if you want to generate type tags from database, set FieldWithTypeTag true
		FieldWithTypeTag: true,
		// if you need unit tests for query code, set WithUnitTest true
		// WithUnitTest: true,
	})

	g.WithImportPkgPath("github.com/bangumi/server/internal/model")
	g.WithJSONTagNameStrategy(func(_ string) string {
		return ""
	})

	// reuse the database connection in Project or create a connection here
	// if you want to use GenerateModel/GenerateModelAs, UseDB is necessary, otherwise it will panic
	c, err := config.NewAppConfig()
	if err != nil {
		panic("failed to read config: " + err.Error())
	}

	conn, err := driver.NewMysqlConnectionPool(c)
	if err != nil {
		panic(err)
	}

	db, err := dal.NewDB(conn, c)
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
		gen.FieldType("uid", userIDTypeString),
		gen.FieldType("privacy", "[]byte"),
	)

	modelMember := g.GenerateModelAs("chii_members", "Member",
		gen.FieldRename("uid", "ID"),
		gen.FieldType("uid", userIDTypeString),
		gen.FieldType("regdate", "int64"),
		gen.FieldType("password_crypt", "[]byte"),
		gen.FieldType("groupid", "uint8"),
		gen.FieldRelate(field.HasOne, "Fields", modelField, &field.RelateConfig{
			GORMTag: "foreignKey:uid;references:uid",
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

	var oauthApp = g.GenerateModelAs("chii_apps", "App",
		gen.FieldTrimPrefix("app_"),
		gen.FieldType("app_id", "uint32"),
		gen.FieldRename("app_desc", "description"),
		gen.FieldType("app_type", "uint8"),
		gen.FieldRename("app_lasttouch", "UpdatedTime"),
		gen.FieldRename("app_timestamp", createdTime),
		gen.FieldType("app_creator", userIDTypeString),
	)

	g.ApplyBasic(g.GenerateModelAs("chii_oauth_clients", "OAuthClient",
		gen.FieldType("app_id", "uint32"),
		gen.FieldRelate(field.BelongsTo, "App", oauthApp, &field.RelateConfig{
			GORMTag: "foreignKey:app_id;references:app_id",
		}),
	))

	g.ApplyBasic(oauthApp)

	g.ApplyBasic(g.GenerateModelAs("chii_timeline", "TimeLine",
		gen.FieldTrimPrefix("tml_"),
		gen.FieldType("tml_id", timelineIDTypeString),
		gen.FieldType("tml_uid", userIDTypeString),
		gen.FieldType("tml_cat", timelineCatTypeString),
		gen.FieldType("tml_img", "[]byte"),
		gen.FieldType("tml_memo", "[]byte"),
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
			GORMTag: "foreignKey:prsn_id;polymorphic:Owner;polymorphicValue:prsn",
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
			GORMTag: "foreignKey:crt_id;polymorphic:Owner;polymorphicValue:crt",
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
		gen.FieldType("ep_subject_id", subjectIDTypeString),
		gen.FieldRelate(field.BelongsTo, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: "foreignKey:ep_subject_id;references:subject_id",
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
			GORMTag: "foreignKey:rlt_related_subject_id;references:subject_id",
		}),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_subject_revisions", "SubjectRevision",
		gen.FieldTrimPrefix("rev_"),
		gen.FieldRename("rev_name_cn", "NameCN"),
		gen.FieldRename("rev_creator", "CreatorID"),
		gen.FieldType("rev_creator", userIDTypeString),
		gen.FieldType("rev_subject_id", subjectIDTypeString),
		gen.FieldRelate(field.BelongsTo, "Subject", modelSubject, &field.RelateConfig{
			GORMTag: "foreignKey:rev_subject_id;references:subject_id",
		}),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_crt_cast_index", "Cast",
		gen.FieldRename("prsn_id", "PersonID"),
		gen.FieldRename("crt_id", "CharacterID"),
		gen.FieldType("crt_id", characterIDTypeString),
		gen.FieldType("prsn_id", personIDTypeString),
		gen.FieldType("subject_id", subjectIDTypeString),
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
			gen.FieldType("crt_id", characterIDTypeString),
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
		gen.FieldType("prsn_id", personIDTypeString),
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
		// 变量重命名
		gen.FieldRename("idx_rlt_rid", "IndexID"),
		gen.FieldRename("idx_rlt_sid", "SubjectID"),
		gen.FieldRename("idx_rlt_type", "SubjectType"),
		gen.FieldRename("idx_rlt_dateline", "CreatedTime"),
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

	g.ApplyBasic(g.GenerateModelAs("chii_groups", "Group",
		gen.FieldTrimPrefix("grp_"),
		gen.FieldType("grp_id", groupIDTypeString),
		gen.FieldType("grp_accessible", "uint8"),
		gen.FieldType("grp_creator", userIDTypeString),
		gen.FieldRename("grp_creator", "CreatorID"),
		gen.FieldRename("grp_desc", "Description"),
		gen.FieldRename("grp_builddate", createdTime),
		gen.FieldRename("grp_lastpost", "LastPostedTime"),
		gen.FieldComment("grp_lastpost", "目前永远是0"),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_group_members", "GroupMember",
		gen.FieldTrimPrefix("gmb_"),
		gen.FieldRename("gmb_uid", "UserID"),
		gen.FieldType("gmb_uid", userIDTypeString),
		gen.FieldType("gmb_gid", groupIDTypeString),
		gen.FieldRename("gmb_gid", "GroupID"),
		gen.FieldRename("gmb_dateline", createdTime),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_subject_topics", "SubjectTopic",
		gen.FieldTrimPrefix("sbj_tpc_"),
		gen.FieldRename("sbj_tpc_subject_id", "SubjectID"),
		gen.FieldRename("sbj_tpc_dateline", createdTime),
		gen.FieldRename("sbj_tpc_lastpost", "UpdatedTime"),
		gen.FieldRename("sbj_tpc_display", "Display"),
		gen.FieldType("sbj_tpc_state", "uint8"),
		gen.FieldType("sbj_tpc_display", "uint8"),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_group_topics", "GroupTopic",
		gen.FieldTrimPrefix("grp_tpc_"),
		gen.FieldRename("grp_tpc_gid", "GroupID"),
		gen.FieldRename("grp_tpc_dateline", createdTime),
		gen.FieldRename("grp_tpc_lastpost", "UpdatedTime"),
		gen.FieldRename("grp_tpc_display", "Display"),
		gen.FieldType("grp_tpc_state", "uint8"),
		gen.FieldType("grp_tpc_display", "uint8"),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_subject_posts", "SubjectTopicComment",
		gen.FieldTrimPrefix("sbj_pst_"),
		gen.FieldRename("sbj_pst_mid", "TopicID"),
		gen.FieldType("sbj_pst_state", "uint8"),
		gen.FieldRename("sbj_pst_dateline", createdTime),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_group_posts", "GroupTopicComment",
		gen.FieldTrimPrefix("grp_pst_"),
		gen.FieldRename("grp_pst_mid", "TopicID"),
		gen.FieldType("grp_pst_state", "uint8"),
		gen.FieldRename("grp_pst_dateline", createdTime),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_ep_comments", "EpisodeComment",
		gen.FieldTrimPrefix("ep_pst_"),
		gen.FieldRename("ep_pst_mid", "TopicID"),
		gen.FieldRename("ep_pst_dateline", createdTime),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_crt_comments", "CharacterComment",
		gen.FieldTrimPrefix("crt_pst_"),
		gen.FieldRename("crt_pst_mid", "TopicID"),
		gen.FieldRename("crt_pst_dateline", createdTime),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_index_comments", "IndexComment",
		gen.FieldTrimPrefix("idx_pst_"),
		gen.FieldRename("idx_pst_mid", "TopicID"),
		gen.FieldRename("idx_pst_dateline", createdTime),
	))

	g.ApplyBasic(g.GenerateModelAs("chii_prsn_comments", "PersonComment",
		gen.FieldTrimPrefix("prsn_pst_"),
		gen.FieldRename("prsn_pst_mid", "TopicID"),
		gen.FieldRename("prsn_pst_dateline", createdTime),
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

	// execute the action of code generation
	g.Execute()
}
