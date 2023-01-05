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

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/user"
)

func (ctl Ctrl) GetUsersByIDs(ctx context.Context, userIDs []model.UserID) (map[model.UserID]user.User, error) {
	if len(userIDs) == 0 {
		return map[model.UserID]user.User{}, nil
	}

	users, err := ctl.user.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, errgo.Wrap(err, "failed to get subjects")
	}

	return users, nil
}

func (ctl Ctrl) GetFriends(ctx context.Context, id model.UserID) (map[model.UserID]user.FriendItem, error) {
	if id == 0 {
		return map[model.UserID]user.FriendItem{}, nil
	}

	f, err := ctl.user.GetFriends(ctx, id)
	if err != nil {
		return nil, errgo.Wrap(err, "userRepo.GetFriends")
	}

	return f, nil
}
