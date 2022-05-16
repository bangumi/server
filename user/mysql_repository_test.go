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

package user_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/test"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/user"
)

func getRepo(t *testing.T) domain.UserRepo {
	t.Helper()
	repo, err := user.NewUserRepo(query.Use(test.GetGorm(t)), zap.NewNop())
	require.NoError(t, err)

	return repo
}

func TestGetByID(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	const id uint32 = 382951

	u, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)

	require.Equal(t, id, u.ID)
	require.Equal(t, "000/38/29/382951.jpg?r=1571167246", u.Avatar)
	require.False(t, u.RegistrationTime.IsZero())
}

func TestGetByName(t *testing.T) {
	test.RequireEnv(t, "mysql")
	t.Parallel()

	repo := getRepo(t)

	const id uint32 = 382951

	u, err := repo.GetByName(context.Background(), "382951")
	require.NoError(t, err)

	require.Equal(t, id, u.ID)
	require.Equal(t, "000/38/29/382951.jpg?r=1571167246", u.Avatar)
}

func TestMysqlRepo_GetCollection(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	const id model.UIDType = 382951
	const subjectID model.SubjectIDType = 888998

	q := test.GetQuery(t)
	err := q.WithContext(context.Background()).SubjectCollection.Create(&dao.SubjectCollection{
		UID:       id,
		SubjectID: subjectID,
		Rate:      2,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err = q.WithContext(context.Background()).SubjectCollection.
			Where(q.SubjectCollection.SubjectID.Eq(subjectID), q.SubjectCollection.UID.Eq(id)).Delete()
		require.NoError(t, err)
	})

	c, err := repo.GetCollection(context.Background(), id, subjectID)
	require.NoError(t, err)

	require.Equal(t, uint8(2), c.Rate)
}
