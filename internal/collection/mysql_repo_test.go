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

package collection_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/collection"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) (domain.CollectionRepo, *query.Query) {
	t.Helper()
	q := test.GetQuery(t)
	repo, err := collection.NewMysqlRepo(q, zap.NewNop())
	require.NoError(t, err)

	return repo, q
}

func TestMysqlRepo_GetCollection(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvMysql)

	const id model.UserID = 382951
	const subjectID model.SubjectID = 888998

	repo, q := getRepo(t)

	test.RunAndCleanup(t, func() {
		_, err := q.WithContext(context.Background()).SubjectCollection.
			Where(q.SubjectCollection.SubjectID.Eq(subjectID), q.SubjectCollection.UserID.Eq(id)).Delete()
		require.NoError(t, err)
	})

	err := q.WithContext(context.Background()).SubjectCollection.Create(&dao.SubjectCollection{
		UserID:    id,
		SubjectID: subjectID,
		Rate:      2,
	})
	require.NoError(t, err)

	c, err := repo.GetSubjectCollection(context.Background(), id, subjectID)
	require.NoError(t, err)

	require.Equal(t, uint8(2), c.Rate)
}

func TestMysqlRepo_CountSubjectCollections(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvMysql)

	const id model.UserID = 382951

	repo, _ := getRepo(t)

	count, err := repo.CountSubjectCollections(context.Background(), id,
		model.SubjectTypeAll, model.SubjectCollectionAll, true)
	require.NoError(t, err)
	require.EqualValues(t, 25, count)
}

func TestMysqlRepo_ListSubjectCollection(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvMysql)

	const id model.UserID = 382951

	repo, _ := getRepo(t)

	data, err := repo.ListSubjectCollection(context.Background(), id,
		model.SubjectTypeGame, model.SubjectCollectionAll, true, 5, 0)
	require.NoError(t, err)
	require.Len(t, data, 5)
}
