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
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/trim21/errgo"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/pkg/gstr"
	"github.com/bangumi/server/internal/user"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger, db *sqlx.DB) Repo {
	return mysqlRepo{
		q:   q,
		log: log.Named("auth.mysqlRepo"),
		db:  db,
	}
}

type mysqlRepo struct {
	q   *query.Query
	db  *sqlx.DB
	log *zap.Logger
}

func (m mysqlRepo) GetByToken(ctx context.Context, token string) (UserInfo, error) {
	var access struct {
		UserID string `db:"user_id"`
	}
	err := m.db.GetContext(ctx, &access,
		`select user_id from chii_oauth_access_tokens
               where access_token = BINARY ? and expires > ? limit 1`, token, time.Now())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserInfo{}, gerr.ErrNotFound
		}

		return UserInfo{}, errgo.Wrap(err, "gorm")
	}

	id, err := gstr.ParseUint32(access.UserID)
	if err != nil || id == 0 {
		m.log.Error("wrong UserID in OAuth Access table", zap.String("user_id", access.UserID))
		return UserInfo{}, errgo.Wrap(err, "parsing user id")
	}

	var u struct {
		Regdate int64
		GroupID user.GroupID
	}

	err = m.db.QueryRowContext(ctx, `select regdate, groupid from chii_members where uid = ? limit 1`, id).
		Scan(&u.Regdate, &u.GroupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UserInfo{}, gerr.ErrNotFound
		}

		return UserInfo{}, errgo.Wrap(err, "gorm")
	}

	return UserInfo{
		RegTime: time.Unix(u.Regdate, 0),
		ID:      id,
		GroupID: u.GroupID,
	}, nil
}

func (m mysqlRepo) GetPermission(ctx context.Context, groupID uint8) (Permission, error) {
	r, err := m.q.UserGroup.WithContext(ctx).Where(m.q.UserGroup.ID.Eq(groupID)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			m.log.Error("can't find permission for group", zap.Uint8("user_group_id", groupID))
			return Permission{}, nil
		}
		return Permission{}, errgo.Wrap(err, "dal")
	}

	p, err := parseSerializedPermission(r.Perm)
	if err != nil {
		m.log.Error("failed to decode php serialized content", zap.Error(err), zap.Uint8("user_group_id", groupID))
		return Permission{}, nil
	}

	return p, nil
}
