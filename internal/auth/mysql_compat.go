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

func (v *phpPermissionValue) boolPtr() *bool {
	if v == nil {
		return nil
	}
	value := v.enabled()
	return &value
}

func (p phpPermission) toPermission() Permission {
	return Permission{
		UserList:           p.UserList.enabled(),
		ManageUserGroup:    p.ManageUserGroup.enabled(),
		ManageUserPhoto:    p.ManageUserPhoto.enabled(),
		ManageTopicState:   p.ManageTopicState.enabled(),
		ManageReport:       p.ManageReport.enabled(),
		UserBan:            p.UserBan.enabled(),
		ManageUser:         p.ManageUser.enabled(),
		UserGroup:          p.UserGroup.enabled(),
		UserWikiApply:      p.UserWikiApply.enabled(),
		UserWikiApprove:    p.UserWikiApprove.enabled(),
		DoujinSubjectErase: p.DoujinSubjectErase.enabled(),
		DoujinSubjectLock:  p.DoujinSubjectLock.enabled(),
		SubjectEdit:        p.SubjectEdit.enabled(),
		SubjectLock:        p.SubjectLock.enabled(),
		SubjectRefresh:     p.SubjectRefresh.enabled(),
		SubjectRelated:     p.SubjectRelated.enabled(),
		SubjectMerge:       p.SubjectMerge.enabled(),
		SubjectErase:       p.SubjectErase.enabled(),
		SubjectCoverLock:   p.SubjectCoverLock.enabled(),
		SubjectCoverErase:  p.SubjectCoverErase.enabled(),
		MonoEdit:           p.MonoEdit.enabled(),
		MonoLock:           p.MonoLock.enabled(),
		MonoMerge:          p.MonoMerge.enabled(),
		MonoErase:          p.MonoErase.enabled(),
		BanPost:            p.BanPost.enabled(),
		BanPostGroup:       p.BanPostGroup.enabled(),
		BanPostGroupNSFW:   p.BanPostGroupNSFW.enabled(),
		BanSubjectInterest: p.BanSubjectInterest.enabled(),
		BanVisit:           p.BanVisit.enabled(),
		Deleted:            p.Deleted.enabled(),
		EpEdit:             p.EpEdit.enabled(),
		EpMove:             p.EpMove.enabled(),
		EpMerge:            p.EpMerge.enabled(),
		EpLock:             p.EpLock.enabled(),
		EpErase:            p.EpErase.enabled(),
		Report:             p.Report.enabled(),
		ManageApp:          p.ManageApp.enabled(),
		AppErase:           p.AppErase.enabled(),
	}
}

func (p phpPermission) toPersonalPermission() PersonalPermission {
	return PersonalPermission{
		UserList:           p.UserList.boolPtr(),
		ManageUserGroup:    p.ManageUserGroup.boolPtr(),
		ManageUserPhoto:    p.ManageUserPhoto.boolPtr(),
		ManageTopicState:   p.ManageTopicState.boolPtr(),
		ManageReport:       p.ManageReport.boolPtr(),
		UserBan:            p.UserBan.boolPtr(),
		ManageUser:         p.ManageUser.boolPtr(),
		UserGroup:          p.UserGroup.boolPtr(),
		UserWikiApply:      p.UserWikiApply.boolPtr(),
		UserWikiApprove:    p.UserWikiApprove.boolPtr(),
		DoujinSubjectErase: p.DoujinSubjectErase.boolPtr(),
		DoujinSubjectLock:  p.DoujinSubjectLock.boolPtr(),
		SubjectEdit:        p.SubjectEdit.boolPtr(),
		SubjectLock:        p.SubjectLock.boolPtr(),
		SubjectRefresh:     p.SubjectRefresh.boolPtr(),
		SubjectRelated:     p.SubjectRelated.boolPtr(),
		SubjectMerge:       p.SubjectMerge.boolPtr(),
		SubjectErase:       p.SubjectErase.boolPtr(),
		SubjectCoverLock:   p.SubjectCoverLock.boolPtr(),
		SubjectCoverErase:  p.SubjectCoverErase.boolPtr(),
		MonoEdit:           p.MonoEdit.boolPtr(),
		MonoLock:           p.MonoLock.boolPtr(),
		MonoMerge:          p.MonoMerge.boolPtr(),
		MonoErase:          p.MonoErase.boolPtr(),
		BanPost:            p.BanPost.boolPtr(),
		BanPostGroup:       p.BanPostGroup.boolPtr(),
		BanPostGroupNSFW:   p.BanPostGroupNSFW.boolPtr(),
		BanSubjectInterest: p.BanSubjectInterest.boolPtr(),
		BanVisit:           p.BanVisit.boolPtr(),
		Deleted:            p.Deleted.boolPtr(),
		EpEdit:             p.EpEdit.boolPtr(),
		EpMove:             p.EpMove.boolPtr(),
		EpMerge:            p.EpMerge.boolPtr(),
		EpLock:             p.EpLock.boolPtr(),
		EpErase:            p.EpErase.boolPtr(),
		Report:             p.Report.boolPtr(),
		ManageApp:          p.ManageApp.boolPtr(),
		AppErase:           p.AppErase.boolPtr(),
	}
}

func decodePHP(raw []byte) (phpPermission, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("[]")) {
		return phpPermission{}, nil
	}

	var p phpPermission
	if err := serialize.Decode(raw, &p); err != nil {
		return phpPermission{}, errgo.Wrap(err, "serialize.Decode")
	}

	return p, nil
}

func parseSerializedPermission(raw []byte) (Permission, error) {
	p, err := decodePHP(raw)
	if err != nil {
		return Permission{}, errgo.Wrap(err, "parsing permission")
	}

	return p.toPermission(), nil
}

func decodePermission(raw []byte) (PersonalPermission, error) {
	p, err := decodePHP(raw)
	if err != nil {
		return PersonalPermission{}, errgo.Wrap(err, "parsing personal permissions")
	}

	return p.toPersonalPermission(), nil
}
