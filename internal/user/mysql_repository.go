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
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
)

func NewUserRepo(q *query.Query, log *zap.Logger) (domain.UserRepo, error) {
	return mysqlRepo{q: q, log: log.Named("user.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (m mysqlRepo) GetByID(ctx context.Context, userID model.UserID) (model.User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.ID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, domain.ErrUserNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))
		return model.User{}, errgo.Wrap(err, "dal")
	}

	return fromDao(u), nil
}

func (m mysqlRepo) GetByName(ctx context.Context, username string) (model.User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.Username.Eq(username)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, domain.ErrUserNotFound
		}

		m.log.Error("unexpected error happened", zap.Error(err))
		return model.User{}, errgo.Wrap(err, "dal")
	}

	return fromDao(u), nil
}

func (m mysqlRepo) GetByIDs(ctx context.Context, ids []model.UserID) (map[model.UserID]model.User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.ID.In(slice.ToValuer(ids)...)).Find()
	if err != nil {
		m.log.Error("unexpected error happened", zap.Error(err))
		return nil, errgo.Wrap(err, "dal")
	}

	var r = make(map[model.UserID]model.User, len(ids))

	for _, member := range u {
		r[member.ID] = fromDao(member)
	}

	return r, nil
}

func (m mysqlRepo) GetFriends(ctx context.Context, userID model.UserID) (map[model.UserID]domain.FriendItem, error) {
	friends, err := m.q.Friend.WithContext(ctx).Where(m.q.Friend.UserID.Eq(userID)).Find()
	if err != nil {
		return nil, errgo.Wrap(err, "friend.Find")
	}

	var r = make(map[model.UserID]domain.FriendItem, len(friends))
	for _, friend := range friends {
		r[friend.FriendID] = struct{}{}
	}

	return r, nil
}

func fromDao(m *dao.Member) model.User {
	return model.User{
		UserName:         m.Username,
		NickName:         m.Nickname,
		UserGroup:        m.Groupid,
		Avatar:           m.Avatar,
		Sign:             m.Sign,
		ID:               m.ID,
		RegistrationTime: time.Unix(m.Regdate, 0),
	}
}
