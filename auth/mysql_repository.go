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
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/strparse"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.AuthRepo, error) {
	return mysqlRepo{q: q, log: log.Named("auth.mysqlRepo")}, nil
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

	id, err := strparse.Uint32(access.UserID)
	if err != nil {
		m.log.Error("wrong UserID in OAuth Access table", zap.String("UserID", access.UserID))

		return domain.Auth{}, errgo.Wrap(err, "parsing user id")
	}

	u, err := m.q.Member.WithContext(ctx).GetByID(id)
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
			m.log.Error("can't find permission for group", zap.Uint8("id", groupID))
			return domain.Permission{}, nil
		}

		m.log.Error("unexpected error", zap.Error(err))
		return domain.Permission{}, errgo.Wrap(err, "dal")
	}

	var p phpPermission
	if err := phpserialize.Unmarshal(r.Perm, &p); err != nil {
		m.log.Error("failed to decode php serialized content", zap.Error(err))
		return domain.Permission{}, nil
	}

	return domain.Permission{
		UserList:          parseBool(p.UserList),
		ManageUserGroup:   parseBool(p.ManageUserGroup),
		ManageUser:        parseBool(p.ManageUser),
		DoujinSubjectLock: parseBool(p.DoujinSubjectLock),
		SubjectEdit:       parseBool(p.SubjectEdit),
		SubjectLock:       parseBool(p.SubjectLock),
		SubjectRefresh:    parseBool(p.SubjectRefresh),
		SubjectRelated:    parseBool(p.SubjectRelated),
		MonoMerge:         parseBool(p.MonoMerge),
		MonoErase:         parseBool(p.MonoErase),
		EpEdit:            parseBool(p.EpEdit),
		EpMove:            parseBool(p.EpMove),
		Report:            parseBool(p.Report),
		AppErase:          parseBool(p.AppErase),
	}, nil
}

func parseBool(s string) bool {
	return s == "1"
}

type phpPermission struct {
	UserList          string `php:"user_list"`
	ManageUserGroup   string `php:"manage_user_group"`
	ManageUser        string `php:"manage_user"`
	DoujinSubjectLock string `php:"doujin_subject_lock"`
	SubjectEdit       string `php:"subject_edit"`
	SubjectLock       string `php:"subject_lock"`
	SubjectRefresh    string `php:"subject_refresh"`
	SubjectRelated    string `php:"subject_related"`
	MonoMerge         string `php:"mono_merge"`
	MonoErase         string `php:"mono_erase"`
	EpEdit            string `php:"ep_edit"`
	EpMove            string `php:"ep_move"`
	Report            string `php:"report"`
	AppErase          string `php:"app_erase"`
}
