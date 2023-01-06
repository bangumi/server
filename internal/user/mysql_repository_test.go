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

package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/user"
)

func getRepo(t *testing.T) user.Repo {
	t.Helper()
	repo, err := user.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestGetByID(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	const id model.UserID = 382951

	u, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)

	require.Equal(t, id, u.ID)
	require.Equal(t, "000/38/29/382951.jpg?r=1571167246", u.Avatar)
	require.False(t, u.RegistrationTime.IsZero())
}

func TestGetByID_notfound(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	const id model.UserID = 382951000

	_, err := repo.GetByID(context.Background(), id)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestGetByName(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	const id model.UserID = 382951

	u, err := repo.GetByName(context.Background(), "382951")
	require.NoError(t, err)

	require.Equal(t, id, u.ID)
	require.Equal(t, "000/38/29/382951.jpg?r=1571167246", u.Avatar)
}

func TestGetByName_notfound(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	_, err := repo.GetByName(context.Background(), "382951000")
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestMysqlRepo_GetFriends(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	const id model.UserID = 287622

	friends, err := repo.GetFriends(context.Background(), id)
	require.NoError(t, err)

	require.Len(t, friends, 1)

	_, ok := friends[427613]
	require.True(t, ok, "map should contain user")
}

func TestMysqlRepo_GetByIDs(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	users, err := repo.GetByIDs(context.Background(), []model.UserID{1, 382951})
	require.NoError(t, err)

	require.Len(t, users, 2)

	require.Equal(t, model.UserID(1), users[1].ID)
	require.Equal(t, model.UserID(382951), users[382951].ID)
}

func TestGetFieldsByIDs(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	fields, err := repo.GetFieldsByIDs(context.Background(), []model.UserID{model.UserID(1)})
	require.NoError(t, err)
	require.Len(t, fields, 1)
}
