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

	"github.com/trim21/errgo"

	"github.com/bangumi/server/internal/pkg/serialize"
)

func parseSerializedPermission(raw []byte) (Permission, error) {
	permission, err := decodePermission(raw)
	permission.DefinedFields = nil
	return permission, err
}

func decodePermission(raw []byte) (Permission, error) {
	permission := Permission{DefinedFields: []string{}}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("[]")) {
		return permission, nil
	}

	fields := phpPermissionFields(&permission)
	if bytes.HasPrefix(raw, []byte("a:")) {
		var values map[string]string
		if err := serialize.Decode(raw, &values); err != nil {
			return Permission{}, errgo.Wrap(err, "decoding PHP permissions")
		}
		for name, value := range values {
			if field, ok := fields[name]; ok {
				*field = value != "" && value != "0"
				permission.DefinedFields = append(permission.DefinedFields, name)
			}
		}
		return permission, nil
	}

	var values map[string]any
	if err := json.Unmarshal(raw, &values); err != nil {
		return Permission{}, errgo.Wrap(err, "decoding JSON permissions")
	}
	for name, value := range values {
		field, ok := fields[name]
		if !ok {
			continue
		}
		switch v := value.(type) {
		case bool:
			*field = v
		case string:
			*field = v != "" && v != "0"
		case float64:
			*field = v != 0
		case nil:
			*field = false
		default:
			continue
		}
		permission.DefinedFields = append(permission.DefinedFields, name)
	}
	return permission, nil
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
