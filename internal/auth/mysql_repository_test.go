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
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) auth.Repo {
	t.Helper()
	q := query.Use(test.GetGorm(t))
	repo := auth.NewMysqlRepo(q, zap.NewNop(), sqlx.NewDb(lo.Must(q.DB().DB()), "mysql"))

	return repo
}

func TestMysqlRepo_GetByToken_NotFound(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	_, err := repo.GetByToken(context.Background(), "not exist token")
	require.ErrorIs(t, err, gerr.ErrNotFound)
}

func TestMysqlRepo_GetByToken(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	u, err := repo.GetByToken(context.Background(), "a_development_access_token")
	require.NoError(t, err)

	require.EqualValues(t, 382951, u.ID)
}

func TestMysqlRepo_GetByToken_case_sensitive(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	_, err := repo.GetByToken(context.Background(), strings.ToUpper("a_development_access_token"))
	require.ErrorIs(t, err, gerr.ErrNotFound)
}

func TestMysqlRepo_GetByToken_expired(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	_, err := repo.GetByToken(context.Background(), "a_expired_token")
	require.ErrorIs(t, err, gerr.ErrNotFound)
}
