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
	"sort"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) episode.Repo {
	t.Helper()
	repo, err := episode.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestMysqlRepo_Count(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.Count(context.Background(), 253, episode.Filter{})
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

	require.Equal(t, episode.Episode{
		Airdate:   "2008-07-12",
		Name:      "ギアス 狩り",
		NameCN:    "Geass 狩猎",
		Duration:  "24m",
		Ep:        14,
		SubjectID: 8,
		Sort:      14,
		Comment:   11,
		ID:        eid,
		Type:      episode.TypeNormal,
	}, e)
}

func TestMysqlRepo_List(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	testCases := []struct {
		filter episode.Filter
		len    int
	}{
		{filter: episode.Filter{}, len: 31},
		{filter: episode.Filter{Type: null.New(episode.TypeNormal)}, len: 26},
		{filter: episode.Filter{Type: null.New(episode.TypeSpecial)}, len: 1},
		{filter: episode.Filter{Type: null.New(episode.TypeOpening)}, len: 1},
		{filter: episode.Filter{Type: null.New(episode.TypeEnding)}, len: 3},
		{filter: episode.Filter{Type: null.New(episode.TypeMad)}, len: 0},
	}

	for _, tc := range testCases {
		episodes, err := repo.List(context.TODO(), 253, tc.filter, 100, 0)
		require.NoError(t, err)

		orig := slice.Clone(episodes)
		sorted := sort.SliceIsSorted(episodes, func(i, j int) bool { return episodes[i].Less(episodes[j]) })
		require.True(t, sorted, "episode should be sorted"+spew.Sdump(orig))

		require.Len(t, episodes, tc.len)
	}
}

func TestMysqlRepo_List_Limit(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	nums := []int{0, 10, 22, 30, 100}
	expected := []int{31, 10, 22, 30, 31}

	for i, num := range nums {
		episodes, err := repo.List(context.TODO(), 253, episode.Filter{}, num, 0)
		require.NoError(t, err)
		require.Len(t, episodes, expected[i])
	}
}
