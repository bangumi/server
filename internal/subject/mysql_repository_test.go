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

package subject_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/internal/subject"
)

func getRepo(t *testing.T) subject.Repo {
	t.Helper()
	repo, err := subject.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestGet(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.Get(context.Background(), 1, subject.Filter{})
	require.NoError(t, err)
	require.Equal(t, model.SubjectID(1), s.ID)

	s, err = repo.Get(context.Background(), 16, subject.Filter{})
	require.NoError(t, err)
	require.Equal(t, model.SubjectID(16), s.ID)
}

func TestMysqlRepo_Get_filter(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	_, err := repo.Get(context.Background(), 16, subject.Filter{NSFW: null.New(false)})
	require.ErrorIs(t, err, gerr.ErrNotFound)
}

func TestBrowse(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	filter := subject.BrowseFilter{
		Type: 2,
	}
	s, err := repo.Browse(context.Background(), filter, 30, 0)
	require.NoError(t, err)
	require.Equal(t, 20, len(s))

	filter = subject.BrowseFilter{
		Type:     1,
		Category: null.New(uint16(1003)),
	}
	s, err = repo.Browse(context.Background(), filter, 30, 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(s))

	filter = subject.BrowseFilter{
		Type:   1,
		Series: null.New(true),
	}
	s, err = repo.Browse(context.Background(), filter, 30, 0)
	require.NoError(t, err)
	require.Equal(t, 4, len(s))

	filter = subject.BrowseFilter{
		Type: 2,
		Year: null.New(int32(2008)),
	}
	s, err = repo.Browse(context.Background(), filter, 30, 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(s))

	filter = subject.BrowseFilter{
		Type: 4,
		Sort: null.New("rank"),
	}
	s, err = repo.Browse(context.Background(), filter, 30, 0)
	require.NoError(t, err)
	require.Equal(t, 7, len(s))
	require.Equal(t, model.SubjectID(5), s[0].ID)
	require.Equal(t, model.SubjectID(13), s[1].ID)
	require.Equal(t, model.SubjectID(14), s[2].ID)
	require.Equal(t, model.SubjectID(9), s[3].ID)
	require.Equal(t, model.SubjectID(6), s[4].ID)
	require.Equal(t, model.SubjectID(7), s[5].ID)
	require.Equal(t, model.SubjectID(4), s[6].ID)
}

func TestMysqlRepo_GetByIDs(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.GetByIDs(context.Background(), []model.SubjectID{1, 2}, subject.Filter{})
	require.NoError(t, err)

	_, ok := s[1]
	require.True(t, ok)
	require.Equal(t, model.SubjectID(1), s[1].ID)

	_, ok = s[2]
	require.True(t, ok)
	require.Equal(t, model.SubjectID(2), s[2].ID)

	s, err = repo.GetByIDs(context.Background(), []model.SubjectID{16}, subject.Filter{NSFW: null.New(false)})
	require.NoError(t, err)
	require.Len(t, s, 0)
}

func TestMysqlRepo_GetCharacterRelated(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.GetCharacterRelated(context.Background(), 1)
	require.NoError(t, err)

	var found bool
	for _, relation := range s {
		if relation.SubjectID == 8 {
			found = true
		}
	}

	require.True(t, found, "character 1 should be related to subject 8")
}
