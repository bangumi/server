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
	"github.com/trim21/go-phpserialize"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func parseBool(s string) bool {
	return s == "1"
}

type phpPermission struct {
	UserList           string `php:"user_list"`
	ManageUserGroup    string `php:"manage_user_group"`
	ManageUserPhoto    string `php:"manage_user_photo"`
	ManageTopicState   string `php:"manage_topic_state"`
	ManageReport       string `php:"manage_report"`
	UserBan            string `php:"user_ban"`
	ManageUser         string `php:"manage_user"`
	UserGroup          string `php:"user_group"`
	UserWikiApprove    string `php:"user_wiki_approve"`
	DoujinSubjectErase string `php:"doujin_subject_erase"`
	UserWikiApply      string `php:"user_wiki_apply"`
	DoujinSubjectLock  string `php:"doujin_subject_lock"`
	SubjectEdit        string `php:"subject_edit"`
	SubjectLock        string `php:"subject_lock"`
	SubjectRefresh     string `php:"subject_refresh"`
	SubjectRelated     string `php:"subject_related"`
	SubjectMerge       string `php:"subject_merge"`
	SubjectErase       string `php:"subject_erase"`
	SubjectCoverLock   string `php:"subject_cover_lock"`
	SubjectCoverErase  string `php:"subject_cover_erase"`
	MonoEdit           string `php:"mono_edit"`
	MonoLock           string `php:"mono_lock"`
	MonoMerge          string `php:"mono_merge"`
	MonoErase          string `php:"mono_erase"`
	EpEdit             string `php:"ep_edit"`
	EpMove             string `php:"ep_move"`
	EpMerge            string `php:"ep_merge"`
	EpLock             string `php:"ep_lock"`
	EpErase            string `php:"ep_erase"`
	Report             string `php:"report"`
	ManageApp          string `php:"manage_app"`
	AppErase           string `php:"app_erase"`
}

func parsePhpSerializedPermission(b []byte) (domain.Permission, error) {
	var p phpPermission
	if len(b) > 0 {
		err := phpserialize.Unmarshal(b, &p)
		if err != nil {
			return domain.Permission{}, errgo.Wrap(err, "parsing permission: phpserialize.Unmarshal")
		}
	}

	return domain.Permission{
		UserList:           parseBool(p.UserList),
		ManageUserGroup:    parseBool(p.ManageUserGroup),
		ManageUserPhoto:    parseBool(p.ManageUserPhoto),
		ManageTopicState:   parseBool(p.ManageTopicState),
		ManageReport:       parseBool(p.ManageReport),
		UserBan:            parseBool(p.UserBan),
		ManageUser:         parseBool(p.ManageUser),
		UserGroup:          parseBool(p.UserGroup),
		UserWikiApprove:    parseBool(p.UserWikiApprove),
		UserWikiApply:      parseBool(p.UserWikiApply),
		DoujinSubjectErase: parseBool(p.DoujinSubjectErase),
		DoujinSubjectLock:  parseBool(p.DoujinSubjectLock),
		SubjectEdit:        parseBool(p.SubjectEdit),
		SubjectLock:        parseBool(p.SubjectLock),
		SubjectRefresh:     parseBool(p.SubjectRefresh),
		SubjectRelated:     parseBool(p.SubjectRelated),
		SubjectMerge:       parseBool(p.SubjectMerge),
		SubjectErase:       parseBool(p.SubjectErase),
		SubjectCoverLock:   parseBool(p.SubjectCoverLock),
		SubjectCoverErase:  parseBool(p.SubjectCoverErase),
		MonoEdit:           parseBool(p.MonoEdit),
		MonoLock:           parseBool(p.MonoLock),
		MonoMerge:          parseBool(p.MonoMerge),
		MonoErase:          parseBool(p.MonoErase),
		EpEdit:             parseBool(p.EpEdit),
		EpMove:             parseBool(p.EpMove),
		EpMerge:            parseBool(p.EpMerge),
		EpLock:             parseBool(p.EpLock),
		EpErase:            parseBool(p.EpErase),
		Report:             parseBool(p.Report),
		ManageApp:          parseBool(p.ManageApp),
		AppErase:           parseBool(p.AppErase),
	}, nil
}
