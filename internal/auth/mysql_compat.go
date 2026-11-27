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

package auth

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/trim21/errgo"

	"github.com/bangumi/server/internal/pkg/serialize"
)

// phpPermission mirrors the serialized permission column (PHP or JSON).
// Pointer fields distinguish an absent key, which inherits the group value,
// from a key explicitly set to false.
type phpPermission struct {
	UserList           *phpPermissionValue `php:"user_list" json:"user_list"`
	ManageUserGroup    *phpPermissionValue `php:"manage_user_group" json:"manage_user_group"`
	ManageUserPhoto    *phpPermissionValue `php:"manage_user_photo" json:"manage_user_photo"`
	ManageTopicState   *phpPermissionValue `php:"manage_topic_state" json:"manage_topic_state"`
	ManageReport       *phpPermissionValue `php:"manage_report" json:"manage_report"`
	UserBan            *phpPermissionValue `php:"user_ban" json:"user_ban"`
	ManageUser         *phpPermissionValue `php:"manage_user" json:"manage_user"`
	UserGroup          *phpPermissionValue `php:"user_group" json:"user_group"`
	UserWikiApprove    *phpPermissionValue `php:"user_wiki_approve" json:"user_wiki_approve"`
	DoujinSubjectErase *phpPermissionValue `php:"doujin_subject_erase" json:"doujin_subject_erase"`
	UserWikiApply      *phpPermissionValue `php:"user_wiki_apply" json:"user_wiki_apply"`
	DoujinSubjectLock  *phpPermissionValue `php:"doujin_subject_lock" json:"doujin_subject_lock"`
	SubjectEdit        *phpPermissionValue `php:"subject_edit" json:"subject_edit"`
	SubjectLock        *phpPermissionValue `php:"subject_lock" json:"subject_lock"`
	SubjectRefresh     *phpPermissionValue `php:"subject_refresh" json:"subject_refresh"`
	SubjectRelated     *phpPermissionValue `php:"subject_related" json:"subject_related"`
	SubjectMerge       *phpPermissionValue `php:"subject_merge" json:"subject_merge"`
	SubjectErase       *phpPermissionValue `php:"subject_erase" json:"subject_erase"`
	SubjectCoverLock   *phpPermissionValue `php:"subject_cover_lock" json:"subject_cover_lock"`
	SubjectCoverErase  *phpPermissionValue `php:"subject_cover_erase" json:"subject_cover_erase"`
	MonoEdit           *phpPermissionValue `php:"mono_edit" json:"mono_edit"`
	MonoLock           *phpPermissionValue `php:"mono_lock" json:"mono_lock"`
	MonoMerge          *phpPermissionValue `php:"mono_merge" json:"mono_merge"`
	MonoErase          *phpPermissionValue `php:"mono_erase" json:"mono_erase"`
	BanPost            *phpPermissionValue `php:"ban_post" json:"ban_post"`
	BanPostGroup       *phpPermissionValue `php:"ban_post_group" json:"ban_post_group"`
	BanPostGroupNSFW   *phpPermissionValue `php:"ban_post_group_nsfw" json:"ban_post_group_nsfw"`
	BanSubjectInterest *phpPermissionValue `php:"ban_subject_interest" json:"ban_subject_interest"`
	BanVisit           *phpPermissionValue `php:"ban_visit" json:"ban_visit"`
	Deleted            *phpPermissionValue `php:"deleted" json:"deleted"`
	EpEdit             *phpPermissionValue `php:"ep_edit" json:"ep_edit"`
	EpMove             *phpPermissionValue `php:"ep_move" json:"ep_move"`
	EpMerge            *phpPermissionValue `php:"ep_merge" json:"ep_merge"`
	EpLock             *phpPermissionValue `php:"ep_lock" json:"ep_lock"`
	EpErase            *phpPermissionValue `php:"ep_erase" json:"ep_erase"`
	Report             *phpPermissionValue `php:"report" json:"report"`
	ManageApp          *phpPermissionValue `php:"manage_app" json:"manage_app"`
	AppErase           *phpPermissionValue `php:"app_erase" json:"app_erase"`
}

// phpPermissionValue is a boolean stored as a PHP or JSON scalar.
type phpPermissionValue string

func (v *phpPermissionValue) UnmarshalJSON(raw []byte) error {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}

	switch value := value.(type) {
	case nil:
		*v = ""
	case bool:
		if value {
			*v = "1"
		} else {
			*v = "0"
		}
	case string:
		*v = phpPermissionValue(value)
	case float64:
		if value != 0 {
			*v = "1"
		} else {
			*v = "0"
		}
	default:
		return fmt.Errorf("unsupported permission value type %T", value)
	}

	return nil
}

func (v *phpPermissionValue) enabled() bool {
	return v != nil && *v != "" && *v != "0"
}

func (p phpPermission) toPermission() Permission {
	permission := Permission{DefinedFields: []string{}}
	fields := phpPermissionFields(&permission)
	for _, field := range []struct {
		name  string
		value *phpPermissionValue
	}{
		{"user_list", p.UserList},
		{"manage_user_group", p.ManageUserGroup},
		{"manage_user_photo", p.ManageUserPhoto},
		{"manage_topic_state", p.ManageTopicState},
		{"manage_report", p.ManageReport},
		{"user_ban", p.UserBan},
		{"manage_user", p.ManageUser},
		{"user_group", p.UserGroup},
		{"user_wiki_approve", p.UserWikiApprove},
		{"user_wiki_apply", p.UserWikiApply},
		{"doujin_subject_erase", p.DoujinSubjectErase},
		{"doujin_subject_lock", p.DoujinSubjectLock},
		{"subject_edit", p.SubjectEdit},
		{"subject_lock", p.SubjectLock},
		{"subject_refresh", p.SubjectRefresh},
		{"subject_related", p.SubjectRelated},
		{"subject_merge", p.SubjectMerge},
		{"subject_erase", p.SubjectErase},
		{"subject_cover_lock", p.SubjectCoverLock},
		{"subject_cover_erase", p.SubjectCoverErase},
		{"mono_edit", p.MonoEdit},
		{"mono_lock", p.MonoLock},
		{"mono_merge", p.MonoMerge},
		{"mono_erase", p.MonoErase},
		{"ban_post", p.BanPost},
		{"ban_post_group", p.BanPostGroup},
		{"ban_post_group_nsfw", p.BanPostGroupNSFW},
		{"ban_subject_interest", p.BanSubjectInterest},
		{"ban_visit", p.BanVisit},
		{"deleted", p.Deleted},
		{"ep_edit", p.EpEdit},
		{"ep_move", p.EpMove},
		{"ep_merge", p.EpMerge},
		{"ep_lock", p.EpLock},
		{"ep_erase", p.EpErase},
		{"report", p.Report},
		{"manage_app", p.ManageApp},
		{"app_erase", p.AppErase},
	} {
		if field.value == nil {
			continue
		}

		*fields[field.name] = field.value.enabled()
		permission.DefinedFields = append(permission.DefinedFields, field.name)
	}

	return permission
}

func parseSerializedPermission(raw []byte) (Permission, error) {
	permission, err := decodePermission(raw)
	permission.DefinedFields = nil
	return permission, err
}

func decodePermission(raw []byte) (Permission, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("[]")) {
		return Permission{DefinedFields: []string{}}, nil
	}

	var p phpPermission
	if err := serialize.Decode(raw, &p); err != nil {
		return Permission{}, errgo.Wrap(err, "parsing personal permissions: serialize.Decode")
	}

	return p.toPermission(), nil
}

func phpPermissionFields(permission *Permission) map[string]*bool {
	return map[string]*bool{
		"user_list":            &permission.UserList,
		"manage_user_group":    &permission.ManageUserGroup,
		"manage_user_photo":    &permission.ManageUserPhoto,
		"manage_topic_state":   &permission.ManageTopicState,
		"manage_report":        &permission.ManageReport,
		"user_ban":             &permission.UserBan,
		"manage_user":          &permission.ManageUser,
		"user_group":           &permission.UserGroup,
		"user_wiki_apply":      &permission.UserWikiApply,
		"user_wiki_approve":    &permission.UserWikiApprove,
		"doujin_subject_erase": &permission.DoujinSubjectErase,
		"doujin_subject_lock":  &permission.DoujinSubjectLock,
		"subject_edit":         &permission.SubjectEdit,
		"subject_lock":         &permission.SubjectLock,
		"subject_refresh":      &permission.SubjectRefresh,
		"subject_related":      &permission.SubjectRelated,
		"subject_merge":        &permission.SubjectMerge,
		"subject_erase":        &permission.SubjectErase,
		"subject_cover_lock":   &permission.SubjectCoverLock,
		"subject_cover_erase":  &permission.SubjectCoverErase,
		"mono_edit":            &permission.MonoEdit,
		"mono_lock":            &permission.MonoLock,
		"mono_merge":           &permission.MonoMerge,
		"mono_erase":           &permission.MonoErase,
		"ban_post":             &permission.BanPost,
		"ban_post_group":       &permission.BanPostGroup,
		"ban_post_group_nsfw":  &permission.BanPostGroupNSFW,
		"ban_subject_interest": &permission.BanSubjectInterest,
		"ban_visit":            &permission.BanVisit,
		"deleted":              &permission.Deleted,
		"ep_edit":              &permission.EpEdit,
		"ep_move":              &permission.EpMove,
		"ep_merge":             &permission.EpMerge,
		"ep_lock":              &permission.EpLock,
		"ep_erase":             &permission.EpErase,
		"report":               &permission.Report,
		"manage_app":           &permission.ManageApp,
		"app_erase":            &permission.AppErase,
	}
}
