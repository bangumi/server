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
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/collection"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/test"
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
		model.SubjectTypeAll, model.CollectionTypeAll, true)
	require.NoError(t, err)
	require.EqualValues(t, 5, count)
}

func TestMysqlRepo_ListSubjectCollection(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvMysql)

	const id model.UserID = 382951

	repo, _ := getRepo(t)

	data, err := repo.ListSubjectCollection(context.Background(), id,
		model.SubjectTypeGame, model.CollectionTypeAll, true, 5, 0)
	require.NoError(t, err)
	require.Len(t, data, 2)
}

// $ task test-all -- -run TestMysqlRepo_UpdateCollection_Create
func TestMysqlRepo_UpdateCollection_Create(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvMysql)

	const userID model.UserID = 382951
	const subjectID model.SubjectID = 888999

	repo, q := getRepo(t)

	test.RunAndCleanup(t, func() {
		_, err := q.SubjectCollection.WithContext(context.Background()).
			Where(q.SubjectCollection.SubjectID.Eq(subjectID), q.SubjectCollection.UserID.Eq(userID)).Delete()
		require.NoError(t, err)
	})

	now := time.Now()
	err := repo.UpdateSubjectCollection(context.Background(), userID, subjectID, model.SubjectCollectionUpdate{
		UpdatedAt: time.Now(),
	})
	require.NoError(t, err)

	c, err := repo.GetSubjectCollection(context.Background(), userID, subjectID)
	require.NoError(t, err)

	require.Equal(t, now.Unix(), c.UpdatedAt.Unix())
}

// $ task test-all -- -run TestMysqlRepo_GetEpisodeCollection
func TestMysqlRepo_GetEpisodeCollection(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvMysql)

	const userID model.UserID = 287622
	const subjectID model.SubjectID = 372487

	repo, _ := getRepo(t)

	ep, err := repo.GetEpisodeCollection(context.Background(), userID, subjectID)
	require.NoError(t, err)

	require.NotZero(t, len(ep))

	for id, item := range ep {
		require.NotZero(t, id)
		require.NotZero(t, item.ID)
		require.NotZero(t, item.Type)
	}
}
