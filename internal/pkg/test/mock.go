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

package test

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/user"
)

func AnyUserMock() user.Repo {
	mocker := &mocks.UserRepo{}
	mocker.EXPECT().GetByID(mock.Anything, mock.Anything).Return(user.User{}, nil)
	mocker.On("GetByIDs", mock.Anything, mock.Anything).
		Return(func(ctx context.Context, ids []model.UserID) map[model.UserID]user.User {
			var ret = make(map[model.UserID]user.User, len(ids))
			for _, id := range ids {
				ret[id] = user.User{}
			}
			return ret
		}, func(ctx context.Context, ids []model.UserID) error {
			return nil
		})
	mocker.On("GetFieldsByIDs", mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, ids []model.UserID) map[model.UserID]user.Fields {
			var ret = make(map[model.UserID]user.Fields, len(ids))
			for _, id := range ids {
				ret[id] = user.Fields{}
			}
			return ret
		}, func(ctx context.Context, ids []model.UserID) error {
			return nil
		})

	return mocker
}
