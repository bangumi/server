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

package index_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/index"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) domain.IndexRepo {
	t.Helper()
	repo, err := index.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestMysqlRepo_Get(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	i, err := repo.Get(context.Background(), 15045)
	require.NoError(t, err)

	require.EqualValues(t, 15045, i.ID)
	require.EqualValues(t, 14127, i.CreatorID)
	require.False(t, i.NSFW)
}

func TestMysqlRepo_ListSubjects(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	subjects, err := repo.ListSubjects(context.Background(), 15045, model.SubjectTypeAll, 20, 0)
	require.NoError(t, err)
	require.Len(t, subjects, 20)
}
