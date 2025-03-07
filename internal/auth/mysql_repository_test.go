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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) (auth.Repo, *query.Query) {
	t.Helper()
	q := query.Use(test.GetGorm(t))
	repo := auth.NewMysqlRepo(q, zap.NewNop(), sqlx.NewDb(lo.Must(q.RawDB().DB()), "mysql"))

	return repo, q
}

func TestMysqlRepo_GetByToken_NotFound(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo, _ := getRepo(t)

	_, err := repo.GetByToken(context.Background(), "not exist token")
	require.ErrorIs(t, err, gerr.ErrNotFound)
}

func TestMysqlRepo_GetByToken(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo, _ := getRepo(t)

	u, err := repo.GetByToken(context.Background(), "a_development_access_token")
	require.NoError(t, err)

	require.EqualValues(t, 382951, u.ID)
}

func TestMysqlRepo_GetByToken_case_sensitive(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo, _ := getRepo(t)

	_, err := repo.GetByToken(context.Background(), strings.ToUpper("a_development_access_token"))
	require.ErrorIs(t, err, gerr.ErrNotFound)
}

func TestMysqlRepo_GetByToken_expired(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo, _ := getRepo(t)

	_, err := repo.GetByToken(context.Background(), "a_expired_token")
	require.ErrorIs(t, err, gerr.ErrNotFound)
}

func TestMysqlRepo_CreateAccessToken(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo, q := getRepo(t)
	t.Cleanup(func() {
		_, err := q.AccessToken.WithContext(context.TODO()).Where(q.AccessToken.UserID.Eq("1")).Delete()
		require.NoError(t, err)
	})

	token, err := repo.CreateAccessToken(context.Background(), 1, "token name", gtime.OneWeek)
	require.NoError(t, err)
	require.Len(t, token, 40)
}

func TestMysqlRepo_DeleteAccessToken(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	const id = 100
	repo, q := getRepo(t)

	cleanup := func() {
		_, err := q.AccessToken.WithContext(context.TODO()).Where(q.AccessToken.ID.Eq(id)).Delete()
		require.NoError(t, err)
	}
	t.Cleanup(cleanup)

	err := q.AccessToken.WithContext(context.Background()).Create(&dao.AccessToken{
		ID:          id,
		Type:        auth.TokenTypeAccessToken,
		AccessToken: t.Name(),
		ClientID:    "access token",
		UserID:      "2",
		ExpiredAt:   time.Now().Add(gtime.OneWeek),
		Scope:       nil,
		Info:        []byte{},
	})
	require.NoError(t, err)

	ok, err := repo.DeleteAccessToken(context.Background(), id)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestMysqlRepo_ListAccessToken(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo, q := getRepo(t)

	test.RunAndCleanup(t, func() {
		_, err := q.AccessToken.WithContext(context.TODO()).Where(q.AccessToken.UserID.Eq("3")).Delete()
		require.NoError(t, err)
	})

	for i := 1; i < 5; i++ {
		err := q.AccessToken.WithContext(context.Background()).Create(&dao.AccessToken{
			Type:        auth.TokenTypeAccessToken,
			AccessToken: t.Name() + strconv.Itoa(i),
			ClientID:    "access token",
			UserID:      "3",
			ExpiredAt:   time.Now().Add(gtime.OneWeek),
			Scope:       nil,
			Info:        []byte{},
		})
		require.NoError(t, err)
	}

	tokens, err := repo.ListAccessToken(context.Background(), 3)
	require.NoError(t, err)
	require.Len(t, tokens, 4)
}
