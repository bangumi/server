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

package comment_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/comment"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) domain.CommentRepo {
	t.Helper()
	repo, err := comment.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestGet(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.Get(context.Background(), domain.CommentIndex, 1038)
	require.NoError(t, err)
	require.Equal(t, model.CommentID(1038), s.ID)

	_, err = repo.Get(context.Background(), domain.CommentIndex, 1)
	require.Error(t, err)
}

func TestMysqlRepo_GetByRelateIDs(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.GetByRelateIDs(context.Background(), domain.CommentTypeSubjectTopic, 47948)
	require.NoError(t, err)

	require.True(t, len(s) == 4, "fetch related comments")
}

func TestMysqlRepo_Count(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.Count(context.Background(), domain.CommentTypeSubjectTopic, 1)
	require.NoError(t, err)

	require.True(t, s == 60, "count top comments")
}

func TestMysqlRepo_List(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.List(context.Background(), domain.CommentTypeSubjectTopic, 1, 0, 0)
	require.NoError(t, err)

	require.True(t, len(s) != 0, "fetch top comments")
}
