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

package topic_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/topic"
)

func TestMysqlRepo_CountReplies(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.CountReplies(context.Background(), topic.CommentTypeSubjectTopic, 1)
	require.NoError(t, err)

	require.EqualValues(t, 59, s, "count top comments")
}

func TestMysqlRepo_ListReplies(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.ListReplies(context.Background(), topic.CommentTypeSubjectTopic, 1, 10, 0)
	require.NoError(t, err)

	require.NotEqual(t, 0, len(s), "fetch top comments")
}

func TestMysqlRepo_ListReplies_all(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.ListReplies(context.Background(), topic.CommentTypeSubjectTopic, 1, 0, 0)
	require.NoError(t, err)

	require.NotEqual(t, 0, len(s), "fetch top comments")
}
