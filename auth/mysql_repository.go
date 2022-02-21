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

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/strparse"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) (domain.AuthRepo, error) {
	return mysqlRepo{q: q, log: log.Named("user.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
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

	return domain.Auth{ID: u.UID}, nil
}
