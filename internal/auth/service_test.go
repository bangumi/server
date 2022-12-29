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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/mocks"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func getService() auth.Service {
	return auth.NewService(nil, nil, zap.NewNop(), cache.NewNoop())
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

func TestService_GetByToken(t *testing.T) {
	t.Parallel()

	var m = mocks.NewAuthRepo(t)
	m.EXPECT().GetByToken(mock.Anything, test.TreeHoleAccessToken).Return(auth.UserInfo{GroupID: 2}, nil)
	m.EXPECT().GetPermission(mock.Anything, model.UserGroupID(2)).Return(auth.Permission{EpEdit: true}, nil)

	var u = mocks.NewUserRepo(t)

	s := auth.NewService(m, u, zap.NewNop(), cache.NewNoop())

	a, err := s.GetByToken(context.Background(), test.TreeHoleAccessToken)
	require.NoError(t, err)

	require.Equal(t, model.UserGroupID(2), a.GroupID)
	require.True(t, a.Permission.EpEdit)
}
