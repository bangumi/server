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

package user

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
)

func NewUserRepo(q *query.Query, log *zap.Logger) (domain.UserRepo, error) {
	return mysqlRepo{q: q, log: log.Named("user.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (m mysqlRepo) GetByID(ctx context.Context, userID uint32) (model.User, error) {
	u, err := m.q.Member.WithContext(ctx).GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))

		return model.User{}, errgo.Wrap(err, "gorm")
	}

	return fromDao(u), nil
}

func (m mysqlRepo) GetByName(ctx context.Context, username string) (model.User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.Username.Eq(username)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, domain.ErrNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))

		return model.User{}, errgo.Wrap(err, "gorm")
	}

	return fromDao(u), nil
}

func (m mysqlRepo) GetByIDs(ctx context.Context, ids ...uint32) (map[uint32]model.User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.UID.In(ids...)).Find()
	if err != nil {
		m.log.Error("unexpected error happened", zap.Error(err))

		return nil, errgo.Wrap(err, "gorm")
	}

	var r = make(map[uint32]model.User, len(ids))

	for _, member := range u {
		r[member.UID] = fromDao(member)
	}

	return r, nil
}

func fromDao(m *dao.Member) model.User {
	return model.User{
		UserName:  m.Username,
		NickName:  m.Nickname,
		UserGroup: m.Groupid,
		Avatar:    m.Avatar,
		Sign:      m.Sign,
		ID:        m.UID,
	}
}
