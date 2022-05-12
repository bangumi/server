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

package auth_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/auth"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/mocks"
)

func getService() domain.AuthService {
	return auth.NewService(nil, zap.NewNop(), test.NopCache())
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
	m.EXPECT().GetPermission(mock.Anything, domain.GroupID(2)).Return(domain.Permission{EpEdit: true}, nil)

	s := auth.NewService(m, zap.NewNop(), test.NopCache())

	u, err := s.GetByTokenWithCache(context.Background(), test.TreeHoleAccessToken)
	require.NoError(t, err)

	require.Equal(t, domain.GroupID(2), u.GroupID)
	require.True(t, u.Permission.EpEdit)
}

func TestService_GetByTokenWithCache_cached(t *testing.T) {
	t.Parallel()

	var c = mocks.NewGeneric(t)
	c.EXPECT().Get(mock.Anything, cachekey.Auth(test.TreeHoleAccessToken), mock.Anything).
		Run(func(ctx context.Context, key string, value interface{}) {
			vOut := reflect.ValueOf(value).Elem()
			vOut.Set(reflect.ValueOf(domain.Auth{GroupID: 2}))
		}).Return(true, nil)

	var m = mocks.NewAuthRepo(t)
	m.EXPECT().GetPermission(mock.Anything, domain.GroupID(2)).Return(domain.Permission{EpEdit: true}, nil)

	s := auth.NewService(m, zap.NewNop(), c)

	u, err := s.GetByTokenWithCache(context.Background(), test.TreeHoleAccessToken)
	require.NoError(t, err)

	require.Equal(t, domain.GroupID(2), u.GroupID)
	require.True(t, u.Permission.EpEdit)
}
