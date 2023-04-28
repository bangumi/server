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

package infra_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trim21/go-phpserialize"
	"go.uber.org/zap"
	"gorm.io/gen/field"

	"github.com/bangumi/server/dal/dao"
	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/internal/collections"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/collections/infra"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) (collections.Repo, *query.Query) {
	t.Helper()
	q := test.GetQuery(t)
	repo, err := infra.NewMysqlRepo(q, zap.NewNop())
	require.NoError(t, err)

	return repo, q
}

func TestMysqlRepo_GetCollection(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvMysql)

	const id model.UserID = 30000
	const subjectID model.SubjectID = 10000

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

	const id model.UserID = 31000
	// parallel problem

	repo, q := getRepo(t)
	test.RunAndCleanup(t, func() {
		_, err := q.SubjectCollection.
			WithContext(context.Background()).
			Where(q.SubjectCollection.UserID.Eq(id)).
			Delete()
		require.NoError(t, err)
	})

	for i := 0; i < 5; i++ {
		err := q.SubjectCollection.
			WithContext(context.Background()).
			Create(&dao.SubjectCollection{
				UserID:      id,
				SubjectID:   model.SubjectID(i + 100),
				SubjectType: model.SubjectTypeAnime,
				UpdatedTime: uint32(time.Now().Unix()),
			})
		require.NoError(t, err)
	}

	count, err := repo.CountSubjectCollections(context.Background(), id,
		model.SubjectTypeAll, collection.SubjectCollectionAll, true)
	require.NoError(t, err)
	require.EqualValues(t, 5, count)
}

func TestMysqlRepo_ListSubjectCollection(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvMysql)

	const uid model.UserID = 32000

	repo, q := getRepo(t)

	var err error
	test.RunAndCleanup(t, func() {
		_, err = q.SubjectCollection.
			WithContext(context.Background()).
			Where(q.SubjectCollection.UserID.Eq(uid)).
			Delete()
		require.NoError(t, err)
	})

	data, err := repo.ListSubjectCollection(context.Background(), uid,
		model.SubjectTypeAll, collection.SubjectCollectionAll, true, 5, 0)
	require.NoError(t, err)
	require.Len(t, data, 0)

	for i := 0; i < 5; i++ {
		err = q.SubjectCollection.
			WithContext(context.Background()).
			Create(&dao.SubjectCollection{
				UserID:      uid,
				SubjectID:   model.SubjectID(100 + i),
				SubjectType: model.SubjectTypeAnime,
				UpdatedTime: uint32(time.Now().Unix()),
			})
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		err = q.SubjectCollection.
			WithContext(context.Background()).
			Create(&dao.SubjectCollection{
				UserID:      uid,
				SubjectID:   model.SubjectID(200 + i),
				SubjectType: model.SubjectTypeGame,
				UpdatedTime: uint32(time.Now().Unix()),
			})
		require.NoError(t, err)
	}

	getList := func(subjectType model.SubjectType) []collection.UserSubjectCollection {
		data, err = repo.ListSubjectCollection(context.Background(), uid,
			subjectType, collection.SubjectCollectionAll, true, 10, 0)
		require.NoError(t, err)
		return data
	}
	require.Len(t, getList(model.SubjectTypeAll), 7)
	require.Len(t, getList(model.SubjectTypeGame), 2)
	require.Len(t, getList(model.SubjectTypeAnime), 5)
	require.Len(t, getList(model.SubjectTypeBook), 0)
}

func TestMysqlRepo_GetEpisodeCollection(t *testing.T) {
	t.Parallel()
	test.RequireEnv(t, test.EnvMysql)

	const uid model.UserID = 33000
	const sid model.SubjectID = 11000

	repo, q := getRepo(t)
	table := q.SubjectCollection

	test.RunAndCleanup(t, func() {
		_, err := table.WithContext(context.TODO()).Where(field.Or(table.SubjectID.Eq(sid), table.UserID.Eq(uid))).Delete()
		require.NoError(t, err)
	})

	now := time.Now()

	_, err := repo.UpdateEpisodeCollection(context.Background(),
		uid, sid, []model.EpisodeID{1, 2}, collection.EpisodeCollectionDone, now)
	require.NoError(t, err)

	ep, err := repo.GetSubjectEpisodesCollection(context.Background(), uid, sid)
	require.NoError(t, err)

	require.NotZero(t, len(ep))

	for id, item := range ep {
		require.NotZero(t, id)
		require.NotZero(t, item.ID)
		require.NotZero(t, item.Type)
	}
}

func TestMysqlRepo_UpdateOrCreateSubjectCollection(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	const uid model.UserID = 34000
	const sid model.SubjectID = 12000

	repo, q := getRepo(t)
	table := q.SubjectCollection

	test.RunAndCleanup(t, func() {
		_, err := table.WithContext(context.TODO()).Where(field.Or(table.SubjectID.Eq(sid), table.UserID.Eq(uid))).Delete()
		require.NoError(t, err)
	})

	err := table.WithContext(context.Background()).Create(
		&dao.SubjectCollection{
			UserID: uid, SubjectID: sid + 1, Rate: 8, Type: uint8(collection.SubjectCollectionDoing),
		},
		&dao.SubjectCollection{
			UserID: uid + 1, SubjectID: sid, Rate: 8, Type: uint8(collection.SubjectCollectionDoing),
		},
	)
	require.NoError(t, err)

	now := time.Now()

	// DB 里没有数据
	_, err = table.WithContext(context.TODO()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Take()
	require.Error(t, err)

	// 创建
	err = repo.UpdateOrCreateSubjectCollection(context.Background(), uid, sid, now, "",
		func(ctx context.Context, s *collection.Subject) (*collection.Subject, error) {
			return s, nil
		})
	require.NoError(t, err)

	// DB 里有数据
	_, err = table.WithContext(context.TODO()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Take()
	require.NoError(t, err)

	// 更新
	err = repo.UpdateOrCreateSubjectCollection(context.Background(), uid, sid, now, "",
		func(ctx context.Context, s *collection.Subject) (*collection.Subject, error) {
			require.NoError(t, s.UpdateComment("c"))
			require.NoError(t, s.UpdateRate(1, collection.SubjectCollectionDropped))
			s.UpdateType(collection.SubjectCollectionDropped)
			return s, nil
		})
	require.NoError(t, err)

	r, err := table.WithContext(context.TODO()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Take()
	require.NoError(t, err)

	require.EqualValues(t, now.Unix(), r.UpdatedTime)
	require.True(t, r.HasComment)
	require.Equal(t, "c", string(r.Comment))
	require.Equal(t, uint8(1), r.Rate)
	require.EqualValues(t, now.Unix(), r.DroppedTime)
	require.Zero(t, r.WishTime)
	require.Zero(t, r.DoingTime)
	require.Zero(t, r.DoneTime)
	require.Zero(t, r.OnHoldTime)

	// When update to wish state
	err = repo.UpdateOrCreateSubjectCollection(context.Background(), uid, sid, now, "",
		func(ctx context.Context, s *collection.Subject) (*collection.Subject, error) {
			require.NoError(t, s.UpdateRate(1, collection.SubjectCollectionWish))
			s.UpdateType(collection.SubjectCollectionWish)
			return s, nil
		})
	require.NoError(t, err)

	r, err = table.WithContext(context.TODO()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Take()
	require.NoError(t, err)
	require.Equal(t, uint8(0), r.Rate)

	// 确认不会影响到其他用户或 subject
	r, err = table.WithContext(context.Background()).Where(table.SubjectID.Eq(sid+1), table.UserID.Eq(uid)).Take()
	require.NoError(t, err)

	require.EqualValues(t, 8, r.Rate)

	r, err = table.WithContext(context.Background()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid+1)).Take()
	require.NoError(t, err)

	require.EqualValues(t, 8, r.Rate)
}

func TestMysqlRepo_UpdateSubjectCollection(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	const uid model.UserID = 35000
	const sid model.SubjectID = 13000

	repo, q := getRepo(t)
	table := q.SubjectCollection

	test.RunAndCleanup(t, func() {
		_, err := table.WithContext(context.TODO()).Where(field.Or(table.SubjectID.Eq(sid), table.UserID.Eq(uid))).Delete()
		require.NoError(t, err)
	})

	err := table.WithContext(context.Background()).Create(
		&dao.SubjectCollection{
			UserID: uid, SubjectID: sid, Rate: 8, Type: uint8(collection.SubjectCollectionDoing),
		},
		&dao.SubjectCollection{
			UserID: uid, SubjectID: sid + 1, Rate: 8, Type: uint8(collection.SubjectCollectionDoing),
		},
		&dao.SubjectCollection{
			UserID: uid + 1, SubjectID: sid, Rate: 8, Type: uint8(collection.SubjectCollectionDoing),
		},
	)
	require.NoError(t, err)

	now := time.Now()

	err = repo.UpdateSubjectCollection(context.Background(), uid, sid, now, "",
		func(ctx context.Context, s *collection.Subject) (*collection.Subject, error) {
			require.NoError(t, s.UpdateComment("c"))
			require.NoError(t, s.UpdateRate(1, collection.SubjectCollectionDropped))
			s.UpdateType(collection.SubjectCollectionDropped)
			return s, nil
		})
	require.NoError(t, err)

	r, err := table.WithContext(context.TODO()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Take()
	require.NoError(t, err)

	require.EqualValues(t, now.Unix(), r.UpdatedTime)
	require.True(t, r.HasComment)
	require.Equal(t, "c", string(r.Comment))
	require.Equal(t, uint8(1), r.Rate)
	require.EqualValues(t, now.Unix(), r.DroppedTime)
	require.Zero(t, r.WishTime)
	require.Zero(t, r.DoingTime)
	require.Zero(t, r.DoneTime)
	require.Zero(t, r.OnHoldTime)

	r, err = table.WithContext(context.Background()).Where(table.SubjectID.Eq(sid+1), table.UserID.Eq(uid)).Take()
	require.NoError(t, err)

	require.EqualValues(t, 8, r.Rate)

	r, err = table.WithContext(context.Background()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid+1)).Take()
	require.NoError(t, err)

	require.EqualValues(t, 8, r.Rate)
}

func TestMysqlRepo_UpdateSubjectCollectionType(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	const uid model.UserID = 36000
	const sid model.SubjectID = 14000

	repo, q := getRepo(t)
	table := q.SubjectCollection

	test.RunAndCleanup(t, func() {
		_, err := table.WithContext(context.TODO()).Where(field.Or(table.SubjectID.Eq(sid), table.UserID.Eq(uid))).Delete()
		require.NoError(t, err)
	})

	err := table.WithContext(context.Background()).Create(
		&dao.SubjectCollection{
			UserID: uid, SubjectID: sid, Rate: 8, Type: uint8(collection.SubjectCollectionDoing),
		},
	)
	require.NoError(t, err)

	now := time.Now()

	err = repo.UpdateSubjectCollection(context.Background(), uid, sid, now, "",
		func(ctx context.Context, s *collection.Subject) (*collection.Subject, error) {
			s.UpdateType(collection.SubjectCollectionDropped)
			return s, nil
		})
	require.NoError(t, err)

	r, err := table.WithContext(context.TODO()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Take()
	require.NoError(t, err)

	require.EqualValues(t, uint32(now.Unix()), r.DroppedTime)
	require.Zero(t, r.WishTime)
	require.Zero(t, r.DoingTime)
	require.Zero(t, r.DoneTime)
	require.Zero(t, r.OnHoldTime)

	t2 := now.Add(time.Duration(10) * time.Second)

	err = repo.UpdateSubjectCollection(context.Background(), uid, sid, t2, "",
		func(ctx context.Context, s *collection.Subject) (*collection.Subject, error) {
			s.UpdateType(collection.SubjectCollectionDoing)
			return s, nil
		})
	require.NoError(t, err)

	r, err = table.WithContext(context.TODO()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Take()
	require.NoError(t, err)

	require.EqualValues(t, uint32(now.Unix()), r.DroppedTime)
	require.EqualValues(t, uint32(t2.Unix()), r.DoingTime)
	require.Zero(t, r.WishTime)
	require.Zero(t, r.DoneTime)
	require.Zero(t, r.OnHoldTime)
}
func TestMysqlRepo_UpdateEpisodeCollection(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	const uid model.UserID = 37000
	const sid model.SubjectID = 15000

	repo, q := getRepo(t)
	table := q.EpCollection
	test.RunAndCleanup(t, func() {
		_, err := table.WithContext(context.Background()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Delete()
		require.NoError(t, err)
	})

	err := table.WithContext(context.TODO()).Create(&dao.EpCollection{
		UserID:    uid,
		SubjectID: sid,
		Status:    []byte("a:0:{}"),
	})
	require.NoError(t, err)

	now := time.Now()

	_, err = repo.UpdateEpisodeCollection(context.Background(),
		uid, sid, []model.EpisodeID{1, 2}, collection.EpisodeCollectionDone, now)
	require.NoError(t, err)

	r, err := table.WithContext(context.Background()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Take()
	require.NoError(t, err)

	require.EqualValues(t, now.Unix(), r.UpdatedTime)

	var m map[uint32]struct {
		Type int `php:"type"`
	}
	require.NoError(t, phpserialize.Unmarshal(r.Status, &m))
	require.Len(t, m, 2)
	require.Contains(t, m, uint32(1))
	require.EqualValues(t, collection.EpisodeCollectionDone, m[1].Type)
	require.Contains(t, m, uint32(2))
	require.EqualValues(t, collection.EpisodeCollectionDone, m[2].Type)

	// testing remove episode collection
	_, err = repo.UpdateEpisodeCollection(context.Background(),
		uid, sid, []model.EpisodeID{1, 2}, collection.EpisodeCollectionNone, now)
	require.NoError(t, err)

	r, err = table.WithContext(context.Background()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Take()
	require.NoError(t, err)

	var m2 map[uint32]struct {
		Type int `php:"type"`
	}
	require.NoError(t, phpserialize.Unmarshal(r.Status, &m2))
	require.Len(t, m2, 0)
}

// 旧站的条目收藏只会创建条目收藏条目，章节收藏纪录是用到时才创建的。
// 这个测试是针对这种情况
func TestMysqlRepo_UpdateEpisodeCollection_create_ep_status(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	const uid model.UserID = 38000
	const sid model.SubjectID = 16000

	repo, q := getRepo(t)
	table := q.EpCollection
	test.RunAndCleanup(t, func() {
		_, err := table.WithContext(context.Background()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Delete()
		require.NoError(t, err)
	})

	now := time.Now()

	_, err := repo.UpdateEpisodeCollection(context.Background(),
		uid, sid, []model.EpisodeID{1, 2}, collection.EpisodeCollectionDone, now)
	require.NoError(t, err)

	r, err := table.WithContext(context.Background()).Where(table.SubjectID.Eq(sid), table.UserID.Eq(uid)).Take()
	require.NoError(t, err)

	var m map[uint32]struct {
		Type int `php:"type"`
	}
	require.NoError(t, phpserialize.Unmarshal(r.Status, &m))
	require.Len(t, m, 2)
	require.Contains(t, m, uint32(1))
	require.EqualValues(t, collection.EpisodeCollectionDone, m[1].Type)
	require.Contains(t, m, uint32(2))
	require.EqualValues(t, collection.EpisodeCollectionDone, m[2].Type)
}
