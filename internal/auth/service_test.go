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

package auth_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/user"
)

func TestService_GetByToken(t *testing.T) {
	t.Parallel()

	var m = mocks.NewAuthRepo(t)
	m.EXPECT().GetByToken(mock.Anything, test.TreeHoleAccessToken).
		Return(auth.UserInfo{GroupID: 2, Permission: auth.Permission{EpEdit: true}}, nil)
	m.EXPECT().GetPermission(mock.Anything, user.GroupID(2)).Return(auth.Permission{EpEdit: true}, nil)

	var u = mocks.NewUserRepo(t)

	s := auth.NewService(m, u, zap.NewNop(), cache.NewNoop())

	a, err := s.GetByToken(context.Background(), test.TreeHoleAccessToken)
	require.NoError(t, err)

	require.Equal(t, user.GroupID(2), a.GroupID)
	require.True(t, a.Permission.EpEdit)
}

func TestService_GetByTokenCacheMissUsesSetNX(t *testing.T) {
	t.Parallel()

	var c = mocks.NewRedisCache(t)
	c.EXPECT().Get(mock.Anything, cachekey.AuthToken(test.TreeHoleAccessToken), mock.Anything).
		Return(false, nil)
	c.EXPECT().SetNX(mock.Anything, cachekey.AuthToken(test.TreeHoleAccessToken), mock.Anything, 10*time.Minute).
		Return(true, nil)

	var m = mocks.NewAuthRepo(t)
	m.EXPECT().GetByToken(mock.Anything, test.TreeHoleAccessToken).
		Return(auth.UserInfo{GroupID: 2}, nil)
	m.EXPECT().GetPermission(mock.Anything, user.GroupID(2)).Return(auth.Permission{}, nil)

	s := auth.NewService(m, mocks.NewUserRepo(t), zap.NewNop(), c)

	_, err := s.GetByToken(context.Background(), test.TreeHoleAccessToken)
	require.NoError(t, err)
}

func TestService_GetByTokenRevoked(t *testing.T) {
	t.Parallel()

	var c = mocks.NewRedisCache(t)
	c.EXPECT().Get(mock.Anything, cachekey.AuthToken(test.TreeHoleAccessToken), mock.Anything).
		RunAndReturn(func(_ context.Context, _ string, value any) (bool, error) {
			return true, json.Unmarshal([]byte(`{"revoked":true}`), value)
		})

	s := auth.NewService(mocks.NewAuthRepo(t), mocks.NewUserRepo(t), zap.NewNop(), c)

	_, err := s.GetByToken(context.Background(), test.TreeHoleAccessToken)
	require.ErrorIs(t, err, gerr.ErrNotFound)
}

func TestService_RevokeToken(t *testing.T) {
	t.Parallel()

	var cached []byte

	var c = mocks.NewRedisCache(t)
	c.EXPECT().Set(mock.Anything, cachekey.AuthToken(test.TreeHoleAccessToken), mock.Anything, 10*time.Minute).
		Run(func(_ context.Context, _ string, value any, _ time.Duration) {
			cached, _ = json.Marshal(value)
		}).
		Return(nil)

	s := auth.NewService(mocks.NewAuthRepo(t), mocks.NewUserRepo(t), zap.NewNop(), c)

	require.NoError(t, s.RevokeToken(context.Background(), test.TreeHoleAccessToken))

	var entry map[string]any
	require.NoError(t, json.Unmarshal(cached, &entry))
	require.Equal(t, true, entry["revoked"])
}
