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

package ctrl

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/ctrl/internal/cachekey"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func (ctl Ctrl) GetUser(ctx context.Context, userID model.UserID) (model.User, error) {
	ctl.metricUserQueryCount.Inc(1)
	var key = cachekey.User(userID)

	// try to read from cache
	var r model.User
	ok, err := ctl.cache.Get(ctx, key, &r)
	if err != nil {
		return r, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		ctl.metricUserQueryCached.Inc(1)
		return r, nil
	}

	r, err = ctl.user.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return r, domain.ErrUserNotFound
		}

		return r, errgo.Wrap(err, "userRepo.Get")
	}

	if e := ctl.cache.Set(ctx, key, r, time.Hour); e != nil {
		ctl.log.Error("can't set response to cache", zap.Error(e))
	}

	return r, nil
}

func (ctl Ctrl) GetUsersByIDs(ctx context.Context, userIDs []model.UserID) (map[model.UserID]model.User, error) {
	if len(userIDs) == 0 {
		return map[model.UserID]model.User{}, nil
	}

	users, err := ctl.user.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, errgo.Wrap(err, "failed to get subjects")
	}

	return users, nil
}

func (ctl Ctrl) GetFriends(ctx context.Context, id model.UserID) (map[model.UserID]domain.FriendItem, error) {
	if id == 0 {
		return map[model.UserID]domain.FriendItem{}, nil
	}

	f, err := ctl.user.GetFriends(ctx, id)
	if err != nil {
		return nil, errgo.Wrap(err, "userRepo.GetFriends")
	}

	return f, nil
}
