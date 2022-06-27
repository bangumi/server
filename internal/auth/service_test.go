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
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/test"
)

func getService() domain.AuthService {
	return auth.NewService(nil, nil, zap.NewNop(), test.NopCache())
}

func TestService_ComparePassword(t *testing.T) {
	t.Parallel()
	s := getService()
	var hashed = []byte("$2a$12$GA5Pr9GhsyLJcSPoTpYBY.JqTzYZb2nfgSeZ1EK38bfgk/Rykkvuq")
	var input = "lovemeplease"

	eq, err := s.ComparePassword(hashed, input)
	require.NoError(t, err)
	require.True(t, eq)
}

func TestService_GetByTokenWithCache(t *testing.T) {
	t.Parallel()

	var m = mocks.NewAuthRepo(t)
	m.EXPECT().GetByToken(mock.Anything, test.TreeHoleAccessToken).Return(domain.Auth{GroupID: 2}, nil)
	m.EXPECT().GetPermission(mock.Anything, model.UserGroupID(2)).Return(domain.Permission{EpEdit: true}, nil)

	var u = mocks.NewUserRepo(t)

	s := auth.NewService(m, u, zap.NewNop(), test.NopCache())

	a, err := s.GetByTokenWithCache(context.Background(), test.TreeHoleAccessToken)
	require.NoError(t, err)

	require.Equal(t, model.UserGroupID(2), a.GroupID)
	require.True(t, a.Permission.EpEdit)
}

func TestService_GetByTokenWithCache_cached(t *testing.T) {
	t.Parallel()

	var c = mocks.NewCache(t)
	c.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).
		Run(func(ctx context.Context, key string, value interface{}) {
			vOut := reflect.ValueOf(value).Elem()
			vOut.Set(reflect.ValueOf(domain.Auth{GroupID: 2}))
		}).Return(true, nil)

	var m = mocks.NewAuthRepo(t)
	m.EXPECT().GetPermission(mock.Anything, model.UserGroupID(2)).Return(domain.Permission{EpEdit: true}, nil)
	var u = mocks.NewUserRepo(t)

	s := auth.NewService(m, u, zap.NewNop(), c)

	a, err := s.GetByTokenWithCache(context.Background(), test.TreeHoleAccessToken)
	require.NoError(t, err)

	require.Equal(t, model.UserGroupID(2), a.GroupID)
	require.True(t, a.Permission.EpEdit)
}
