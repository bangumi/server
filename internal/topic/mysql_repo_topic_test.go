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

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func TestMysqlRepo_Get(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.Get(context.Background(), domain.TopicTypeSubject, 1)
	require.NoError(t, err)

	require.Equal(t, model.TopicID(1), s.ID)
}

func TestMysqlRepo_Count(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	count, err := repo.Count(context.Background(), domain.TopicTypeSubject, 1, []model.TopicDisplay{
		model.TopicDisplayNormal,
	})
	require.NoError(t, err)
	require.Equal(t, count, int64(1))
}

func TestMysqlRepo_List(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	_, err := repo.List(context.Background(), domain.TopicTypeSubject, 2, []model.TopicDisplay{
		model.TopicDisplayNormal,
	}, 0, 0)
	require.NoError(t, err)
}
