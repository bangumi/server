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

package episode_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) domain.EpisodeRepo {
	t.Helper()
	repo, err := episode.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestMysqlRepo_Count(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.Count(context.Background(), 253)
	require.NoError(t, err)

	require.Equal(t, int64(31), s)
}

func TestMysqlRepo_Get(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	const eid model.EpisodeID = 2

	e, err := repo.Get(context.TODO(), eid)
	require.NoError(t, err)

	require.Equal(t, model.Episode{
		Airdate:   "2008-07-12",
		Name:      "ギアス 狩り",
		NameCN:    "Geass 狩猎",
		Duration:  "24m",
		Ep:        14,
		SubjectID: 8,
		Sort:      14,
		Comment:   11,
		ID:        eid,
		Type:      model.EpTypeNormal,
	}, e)
}
