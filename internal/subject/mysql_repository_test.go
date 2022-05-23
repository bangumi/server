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

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/test"
)

func getRepo(t *testing.T) domain.SubjectRepo {
	t.Helper()
	repo, err := subject.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestGet(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.Get(context.Background(), 1)
	require.NoError(t, err)

	require.Equal(t, model.SubjectIDType(1), s.ID)
}

func TestMysqlRepo_GetByIDs(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	s, err := repo.GetByIDs(context.Background(), 1, 2)
	require.NoError(t, err)

	_, ok := s[1]
	require.True(t, ok)
	require.Equal(t, model.SubjectIDType(1), s[1].ID)

	_, ok = s[2]
	require.True(t, ok)
	require.Equal(t, model.SubjectIDType(2), s[2].ID)
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
