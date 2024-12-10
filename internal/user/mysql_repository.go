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
	"strconv"
	"time"

	"github.com/trim21/errgo"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/gstr"
)

func NewMysqlRepo(q *query.Query, log *zap.Logger) (Repo, error) {
	return mysqlRepo{q: q, log: log.Named("user.mysqlRepo")}, nil
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (m mysqlRepo) GetFullUser(ctx context.Context, userID model.UserID) (FullUser, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.ID.Eq(userID)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FullUser{}, gerr.ErrUserNotFound
		}
		return FullUser{}, errgo.Wrap(err, "dal")
	}

	return FullUser{
		UserName:         u.Username,
		NickName:         u.Nickname,
		UserGroup:        u.Groupid,
		Avatar:           u.Avatar,
		Sign:             string(u.Sign),
		ID:               u.ID,
		RegistrationTime: time.Unix(u.Regdate, 0),
		TimeOffset:       parseTimeOffset(u.Timeoffset),
		Email:            u.Email,
	}, nil
}

// default time zone GMT+8.
const defaultTimeOffset = 8

func parseTimeOffset(s string) int8 {
	switch s {
	case "", "9999":
		return defaultTimeOffset
	}

	v, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return defaultTimeOffset
	}

	return int8(v)
}

func (m mysqlRepo) GetByID(ctx context.Context, userID model.UserID) (User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.ID.Eq(userID)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, gerr.ErrUserNotFound
		}
		return User{}, errgo.Wrap(err, "dal")
	}

	return fromDao(u), nil
}

func (m mysqlRepo) GetByName(ctx context.Context, username string) (User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.Username.Eq(username)).Take()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, gerr.ErrUserNotFound
		}
		return User{}, errgo.Wrap(err, "dal")
	}

	return fromDao(u), nil
}

func (m mysqlRepo) GetByIDs(ctx context.Context, ids []model.UserID) (map[model.UserID]User, error) {
	u, err := m.q.Member.WithContext(ctx).Where(m.q.Member.ID.In(ids...)).Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	var r = make(map[model.UserID]User, len(ids))

	for _, member := range u {
		r[member.ID] = fromDao(member)
	}

	return r, nil
}

func (m mysqlRepo) GetFriends(ctx context.Context, userID model.UserID) (map[model.UserID]FriendItem, error) {
	friends, err := m.q.Friend.WithContext(ctx).Where(m.q.Friend.UserID.Eq(userID)).Find()
	if err != nil {
		return nil, errgo.Wrap(err, "friend.Find")
	}

	var r = make(map[model.UserID]FriendItem, len(friends))
	for _, friend := range friends {
		r[friend.FriendID] = struct{}{}
	}

	return r, nil
}

func (m mysqlRepo) CheckIsFriendToOthers(
	ctx context.Context,
	selfID model.UserID,
	otherIDs ...model.UserID) (bool, error) {
	count, err := m.q.Friend.
		WithContext(ctx).
		Where(m.q.Friend.UserID.In(otherIDs...), m.q.Friend.FriendID.Eq(selfID)).Count()
	if err != nil {
		return false, errgo.Wrap(err, "dal")
	}
	return count >= int64(len(otherIDs)), nil
}

func (m mysqlRepo) GetFieldsByIDs(ctx context.Context,
	userIDs []model.UserID) (map[model.UserID]Fields, error) {
	if len(userIDs) == 0 {
		return make(map[model.UserID]Fields, 0), nil
	}
	users, err := m.q.Member.
		WithContext(ctx).
		Joins(m.q.Member.Fields).Select(m.q.Member.ID).
		Where(m.q.Member.ID.In(userIDs...)).Find()
	if err != nil {
		return nil, errgo.Wrap(err, "dal")
	}

	var r = make(map[model.UserID]Fields, len(users))
	for _, user := range users {
		var privacySettings PrivacySettings
		privacySettings.Unmarshal(user.Fields.Privacy)
		r[user.Fields.UID] = Fields{
			UID:  user.Fields.UID,
			Site: user.Fields.Site,
			Bio:  user.Fields.Bio,
			BlockList: slice.MapFilter(gstr.Split(user.Fields.Blocklist, ","), func(s string) (model.UserID, bool) {
				id, err := gstr.ParseUint32(s)
				if err != nil {
					return 0, false
				}
				return id, true
			}),
			Privacy: privacySettings,
		}
	}

	return r, nil
}

func fromDao(m *dao.Member) User {
	return User{
		UserName:         m.Username,
		NickName:         m.Nickname,
		UserGroup:        m.Groupid,
		Avatar:           m.Avatar,
		Sign:             string(m.Sign),
		ID:               m.ID,
		RegistrationTime: time.Unix(m.Regdate, 0),
	}
}
