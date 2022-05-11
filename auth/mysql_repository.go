// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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
	"context"
	"errors"
	"time"

	"github.com/elliotchance/phpserialize"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/strparse"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) domain.AuthRepo {
	return mysqlRepo{q: q, log: log.Named("auth.mysqlRepo")}
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (m mysqlRepo) GetByEmail(ctx context.Context, email string) (domain.Auth, []byte, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.Email.Eq(email)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Auth{}, nil, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))
		return domain.Auth{}, nil, errgo.Wrap(err, "gorm")
	}

	return domain.Auth{
		RegTime: time.Unix(u.Regdate, 0),
		ID:      u.UID,
		GroupID: u.Groupid,
	}, u.PasswordCrypt, nil
}

func (m mysqlRepo) GetByToken(ctx context.Context, token string) (domain.Auth, error) {
	access, err := m.q.OAuthAccessToken.WithContext(ctx).
		Where(m.q.OAuthAccessToken.AccessToken.Eq(token), m.q.OAuthAccessToken.Expires.Gte(time.Now())).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Auth{}, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))

		return domain.Auth{}, errgo.Wrap(err, "gorm")
	}

	id, err := strparse.UserID(access.UserID)
	if err != nil {
		m.log.Error("wrong UserID in OAuth Access table", zap.String("user_id", access.UserID))
		return domain.Auth{}, errgo.Wrap(err, "parsing user id")
	}

	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.UID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			m.log.Error("can't find user of access token",
				zap.String("token", token), zap.String("uid", access.UserID))

			return domain.Auth{}, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))

		return domain.Auth{}, errgo.Wrap(err, "gorm")
	}

	return domain.Auth{
		RegTime: time.Unix(u.Regdate, 0),
		ID:      u.UID,
		GroupID: u.Groupid,
	}, nil
}

func (m mysqlRepo) GetPermission(ctx context.Context, groupID uint8) (domain.Permission, error) {
	r, err := m.q.UserGroup.WithContext(ctx).Where(m.q.UserGroup.ID.Eq(groupID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			m.log.Error("can't find permission for group", log.GroupID(groupID))
			return domain.Permission{}, nil
		}

		m.log.Error("unexpected error", zap.Error(err))
		return domain.Permission{}, errgo.Wrap(err, "dal")
	}

	var p phpPermission
	if len(r.Perm) > 0 {
		d, err := phpserialize.UnmarshalAssociativeArray(r.Perm)
		if err != nil {
			m.log.Error("failed to decode php serialized content", zap.Error(err), log.GroupID(groupID))
			return domain.Permission{}, nil
		}
		if err = mapstructure.Decode(d, &p); err != nil {
			m.log.Error("failed to convert from map to struct when decoding permission",
				zap.Error(err), zap.Uint8("group_id", groupID))
			return domain.Permission{}, nil
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

func parseBool(s string) bool {
	return s == "1"
}

type phpPermission struct {
	UserList           string `mapstruct:"user_list"`
	ManageUserGroup    string `mapstruct:"manage_user_group"`
	ManageUserPhoto    string `mapstruct:"manage_user_photo"`
	ManageTopicState   string `mapstruct:"manage_topic_state"`
	ManageReport       string `mapstruct:"manage_report"`
	UserBan            string `mapstruct:"user_ban"`
	ManageUser         string `mapstruct:"manage_user"`
	UserGroup          string `mapstruct:"user_group"`
	UserWikiApprove    string `mapstruct:"user_wiki_approve"`
	DoujinSubjectErase string `mapstruct:"doujin_subject_erase"`
	UserWikiApply      string `mapstruct:"user_wiki_apply"`
	DoujinSubjectLock  string `mapstruct:"doujin_subject_lock"`
	SubjectEdit        string `mapstruct:"subject_edit"`
	SubjectLock        string `mapstruct:"subject_lock"`
	SubjectRefresh     string `mapstruct:"subject_refresh"`
	SubjectRelated     string `mapstruct:"subject_related"`
	SubjectMerge       string `mapstruct:"subject_merge"`
	SubjectErase       string `mapstruct:"subject_erase"`
	SubjectCoverLock   string `mapstruct:"subject_cover_lock"`
	SubjectCoverErase  string `mapstruct:"subject_cover_erase"`
	MonoEdit           string `mapstruct:"mono_edit"`
	MonoLock           string `mapstruct:"mono_lock"`
	MonoMerge          string `mapstruct:"mono_merge"`
	MonoErase          string `mapstruct:"mono_erase"`
	EpEdit             string `mapstruct:"ep_edit"`
	EpMove             string `mapstruct:"ep_move"`
	EpMerge            string `mapstruct:"ep_merge"`
	EpLock             string `mapstruct:"ep_lock"`
	EpErase            string `mapstruct:"ep_erase"`
	Report             string `mapstruct:"report"`
	ManageApp          string `mapstruct:"manage_app"`
	AppErase           string `mapstruct:"app_erase"`
}
