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

package person_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) domain.PersonRepo {
	t.Helper()
	return person.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
}

func TestGet(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.Get(context.Background(), 1)
	require.NoError(t, err)

	require.EqualValues(t, 1, s.ID)
}

func TestMysqlRepo_GetByIDs(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.GetByIDs(context.Background(), []model.PersonID{1, 2})
	require.NoError(t, err)

	_, ok := s[1]
	require.True(t, ok)
	require.Equal(t, model.PersonID(1), s[1].ID)

	_, ok = s[2]
	require.True(t, ok)
	require.Equal(t, model.PersonID(2), s[2].ID)
}

func TestMysqlRepo_GetSubjectRelated(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.GetSubjectRelated(context.Background(), 4)
	require.NoError(t, err)

	for _, relation := range s {
		require.NotEmpty(t, relation.TypeID)
	}
}
