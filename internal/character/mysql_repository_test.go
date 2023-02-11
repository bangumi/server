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

package character_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) character.Repo {
	t.Helper()
	repo, err := character.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestGet(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.Get(context.Background(), 1)
	require.NoError(t, err)

	require.EqualValues(t, 1, s.ID)
}

func TestMysqlRepo_Get_err_not_found(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	_, err := repo.Get(context.Background(), 10000)
	require.ErrorIs(t, err, gerr.ErrCharacterNotFound)
}

func TestMysqlRepo_GetByIDs(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.GetByIDs(context.Background(), []model.CharacterID{1, 2})
	require.NoError(t, err)

	_, ok := s[1]
	require.True(t, ok)
	require.EqualValues(t, 1, s[1].ID)

	_, ok = s[2]
	require.True(t, ok)
	require.EqualValues(t, 2, s[2].ID)
}

func TestMysqlRepo_GetPersonRelated(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	c, err := repo.GetPersonRelated(context.TODO(), 1)
	require.NoError(t, err)

	require.Len(t, c, 272)
}

func TestMysqlRepo_GetSubjectRelated(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	c, err := repo.GetSubjectRelated(context.TODO(), 8)
	require.NoError(t, err)

	require.Len(t, c, 3)
	require.Equal(t,
		[]domain.SubjectCharacterRelation{
			{TypeID: 0x1, SubjectID: 0x8, CharacterID: 0x1},
			{TypeID: 0x1, SubjectID: 0x8, CharacterID: 0x2},
			{TypeID: 0x1, SubjectID: 0x8, CharacterID: 0x3},
		},
		c,
	)
}

func TestMysqlRepo_GetRelations(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	r, err := repo.GetSubjectRelationByIDs(context.TODO(), []character.SubjectCompositeId{
		{CharacterID: 1, SubjectID: 8},
		{CharacterID: 1, SubjectID: 793},
		{CharacterID: 1, SubjectID: 32214},
	})
	require.NoError(t, err)

	require.Len(t, r, 3)
	require.Equal(t,
		[]domain.SubjectCharacterRelation{
			{TypeID: 1, SubjectID: 8, CharacterID: 1},
			{TypeID: 1, SubjectID: 793, CharacterID: 1},
			{TypeID: 1, SubjectID: 32214, CharacterID: 1},
		},
		r,
	)
}
