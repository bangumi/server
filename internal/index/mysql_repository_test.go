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

package index_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/index"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) index.Repo {
	t.Helper()
	q := query.Use(test.GetGorm(t))
	repo, err := index.NewMysqlRepo(q, zap.NewNop(), sqlx.NewDb(lo.Must(q.DB().DB()), "mysql"))
	require.NoError(t, err)

	return repo
}

func TestMysqlRepo_Get(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	i, err := repo.Get(context.Background(), 15045)
	require.NoError(t, err)

	require.EqualValues(t, 15045, i.ID)
	require.EqualValues(t, 14127, i.CreatorID)
	require.False(t, i.NSFW)
}

func TestMysqlRepo_GetPrivateIndex(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	ctx := context.Background()
	now := time.Now()

	idx := &model.Index{
		ID:          0,
		Title:       "private index",
		Description: "private visibility",
		CreatorID:   382951,
		CreatedAt:   now,
		UpdatedAt:   now,
		Privacy:     model.IndexPrivacyPrivate,
	}
	require.NoError(t, repo.New(ctx, idx))
	defer func() { _ = repo.Delete(ctx, idx.ID) }()

	got, err := repo.Get(ctx, idx.ID)
	require.NoError(t, err)
	require.Equal(t, idx.ID, got.ID)
	require.Equal(t, model.IndexPrivacyPrivate, got.Privacy)
}

func TestMysqlRepo_GetDeletedIndex(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	ctx := context.Background()
	now := time.Now()

	idx := &model.Index{
		ID:          0,
		Title:       "deleted index",
		Description: "deleted visibility",
		CreatorID:   382951,
		CreatedAt:   now,
		UpdatedAt:   now,
		Privacy:     model.IndexPrivacyPublic,
	}
	require.NoError(t, repo.New(ctx, idx))
	defer func() { _ = repo.Delete(ctx, idx.ID) }()

	require.NoError(t, repo.Delete(ctx, idx.ID))

	_, err := repo.Get(ctx, idx.ID)
	require.ErrorIs(t, err, gerr.ErrNotFound)
}

func TestMysqlRepo_ListSubjects(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	subjects, err := repo.ListSubjects(context.Background(), 15045, model.SubjectTypeAll, 20, 0)
	require.NoError(t, err)
	require.Len(t, subjects, 20)
}

func TestMysqlRepo_NewIndex(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	// 存入的时间戳是 int32 类型， nanosecond 会被忽略掉
	// TODO: 数据库时间戳是否应该改成 uint32 或者 uint64 类型
	now := time.Now()

	index := &model.Index{
		ID:          0,
		Title:       "test",
		Description: "Test Index",
		CreatorID:   382951,
		CreatedAt:   now,
		UpdatedAt:   now,
		Total:       0,
		Comments:    0,
		Collects:    0,
		NSFW:        false,
	}
	err := repo.New(context.Background(), index)
	require.NoError(t, err)
	require.NotEqualValues(t, 0, index.ID)

	i, err := repo.Get(context.Background(), index.ID)
	require.NoError(t, err)

	require.EqualValues(t, 382951, i.CreatorID)
	require.EqualValues(t, "test", i.Title)
	require.EqualValues(t, "Test Index", i.Description)
	require.EqualValues(t, now.Truncate(time.Second), i.CreatedAt)
	require.EqualValues(t, now.Truncate(time.Second), i.UpdatedAt)
}

func TestMysqlRepo_UpdateIndex(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	ctx := context.Background()

	now := time.Now()
	index := &model.Index{
		ID:          0,
		Title:       "test",
		Description: "Test Index",
		CreatorID:   382951,
		CreatedAt:   now,
		UpdatedAt:   now,
		Total:       0,
		Comments:    0,
		Collects:    0,
		NSFW:        false,
	}
	err := repo.New(ctx, index)
	require.NoError(t, err)

	// update index information
	err = repo.Update(ctx, 15045, "日本动画", "日本动画的介绍")
	require.NoError(t, err)

	// check updated index information
	i, err := repo.Get(ctx, 15045)
	require.NoError(t, err)
	require.EqualValues(t, "日本动画", i.Title)
	require.EqualValues(t, "日本动画的介绍", i.Description)

	_ = repo.Delete(ctx, index.ID)
}

func TestMysqlRepo_DeleteIndex(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	index := &model.Index{
		ID:          0,
		Title:       "test",
		Description: "Test Index",
		CreatorID:   382951,
		CreatedAt:   time.Now(),
		Total:       0,
		Comments:    0,
		Collects:    0,
		NSFW:        false,
	}
	_ = repo.New(context.Background(), index)
	require.NotEqual(t, 0, index.ID)

	i, err := repo.Get(context.Background(), index.ID)
	require.NoError(t, err)
	require.EqualValues(t, index.ID, i.ID)

	err = repo.Delete(context.Background(), index.ID)
	require.NoError(t, err)

	_, err = repo.Get(context.Background(), index.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, gerr.ErrNotFound)
}

// 删除目录会把所属的 subject 全部删掉
func TestMysqlRepo_DeleteIndex2(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	index := &model.Index{
		ID:          0,
		Title:       "test",
		Description: "Test Index",
		CreatorID:   382951,
		CreatedAt:   time.Now(),
		Total:       0,
		Comments:    0,
		Collects:    0,
		NSFW:        false,
	}

	ctx := context.Background()

	err := repo.New(ctx, index)
	require.NoError(t, err)

	for i := uint32(10); i < 20; i++ {
		_, err = repo.AddOrUpdateIndexSubject(ctx, index.ID, i, i, fmt.Sprintf("comment %d", i))
		require.NoError(t, err)
	}

	i, err := repo.Get(ctx, index.ID)
	require.NoError(t, err)
	require.EqualValues(t, 10, i.Total)

	subjects, err := repo.ListSubjects(context.Background(), index.ID, model.SubjectTypeAll, 20, 0)
	require.NoError(t, err)
	require.Len(t, subjects, 10)

	err = repo.Delete(ctx, index.ID)
	require.NoError(t, err)

	i, err = repo.Get(ctx, index.ID)
	require.Equal(t, err, gerr.ErrNotFound)

	subjects, err = repo.ListSubjects(context.Background(), index.ID, model.SubjectTypeAll, 20, 0)
	require.ErrorIs(t, err, gerr.ErrNotFound)
	require.Nil(t, subjects)

	// 确保不会影响到其他目录
	subjects, err = repo.ListSubjects(context.Background(), 15045, model.SubjectTypeAll, 20, 0)
	require.NoError(t, err)
	require.Len(t, subjects, 20)
}

func TestMysqlRepo_AddOrUpdateIndexSubject(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	idx := &model.Index{
		ID:          0,
		Title:       "test",
		Description: "Test Index",
		CreatorID:   382951,
		CreatedAt:   time.Now(),
		Total:       0,
		Comments:    0,
		Collects:    0,
		NSFW:        false,
	}

	ctx := context.Background()

	err := repo.New(ctx, idx)
	require.NotEqual(t, 0, idx.ID)
	require.NoError(t, err)

	_, err = repo.AddOrUpdateIndexSubject(ctx, idx.ID, 3, 1, "comment 1")
	require.NoError(t, err)

	_, err = repo.AddOrUpdateIndexSubject(ctx, idx.ID, 4, 3, "comment 3")
	require.NoError(t, err)

	i, err := repo.Get(ctx, idx.ID)
	require.NoError(t, err)
	require.EqualValues(t, idx.ID, i.ID)

	require.EqualValues(t, 2, i.Total)

	subjects, err := repo.ListSubjects(context.Background(), idx.ID, model.SubjectTypeAll, 20, 0)
	require.NoError(t, err)
	require.Len(t, subjects, 2)

	cache := map[model.SubjectID]index.Subject{}
	for _, s := range subjects {
		cache[s.Subject.ID] = s
	}
	require.EqualValues(t, cache[3].Comment, "comment 1")
	require.EqualValues(t, cache[4].Comment, "comment 3")

	err = repo.Delete(ctx, idx.ID)
	require.NoError(t, err)
}

func TestMysqlRepo_DeleteIndexSubject(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	index := &model.Index{
		ID:          0,
		Title:       "test",
		Description: "Test Index",
		CreatorID:   382951,
		CreatedAt:   time.Now(),
		Total:       0,
		Comments:    0,
		Collects:    0,
		NSFW:        false,
	}

	ctx := context.Background()

	err := repo.New(ctx, index)
	require.NotEqual(t, 0, index.ID)
	require.NoError(t, err)

	for i := uint32(10); i < 20; i++ {
		_, err = repo.AddOrUpdateIndexSubject(ctx, index.ID, i, i, fmt.Sprintf("comment %d", i))
		require.NoError(t, err)
	}

	i, err := repo.Get(ctx, index.ID)
	require.NoError(t, err)
	require.EqualValues(t, 10, i.Total)

	subjects, err := repo.ListSubjects(context.Background(), index.ID, model.SubjectTypeAll, 20, 0)
	require.NoError(t, err)
	require.Len(t, subjects, 10)

	err = repo.DeleteIndexSubject(ctx, index.ID, 15)
	require.NoError(t, err)

	i, err = repo.Get(ctx, index.ID)
	require.NoError(t, err)
	require.EqualValues(t, 9, i.Total)

	subjects, err = repo.ListSubjects(context.Background(), index.ID, model.SubjectTypeAll, 20, 0)
	require.NoError(t, err)
	require.Len(t, subjects, 9)

	for _, v := range subjects {
		require.NotEqualValues(t, v.Subject.ID, 15)
	}

	err = repo.Delete(ctx, index.ID)
	require.NoError(t, err)
}

func TestMysqlRepo_DeleteNonExistsIndexSubject(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	ctx := context.Background()

	_ = repo.Delete(ctx, 99999999)

	err := repo.DeleteIndexSubject(ctx, 99999999, 15)
	require.Error(t, err)
	require.Error(t, err, gerr.ErrNotFound)
}

func TestMysqlRepo_FailedAddedToNonExists(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	ctx := context.Background()
	_ = repo.Delete(ctx, 99999999) // in case index(99999999) exists

	_, err := repo.AddOrUpdateIndexSubject(ctx, 99999999, 5, 5, "test")
	require.Error(t, err)
	require.Equal(t, err, gerr.ErrNotFound)
}

func TestMysqlRepo_UpdateSubjectInfo(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	index := &model.Index{
		ID:          0,
		Title:       "test",
		Description: "Test Index",
		CreatorID:   382951,
		CreatedAt:   time.Now(),
		Total:       0,
		Comments:    0,
		Collects:    0,
		NSFW:        false,
	}
	ctx := context.Background()

	err := repo.New(ctx, index)
	require.NoError(t, err)

	_, err = repo.AddOrUpdateIndexSubject(ctx, index.ID, 5, 5, "test")
	require.NoError(t, err)
	subjects, err := repo.ListSubjects(context.Background(), index.ID, model.SubjectTypeAll, 20, 0)
	require.NoError(t, err)
	require.Len(t, subjects, 1)
	require.EqualValues(t, subjects[0].Subject.ID, 5)
	require.EqualValues(t, subjects[0].Comment, "test")

	_, err = repo.AddOrUpdateIndexSubject(ctx, index.ID, 5, 5, "test22222")
	require.NoError(t, err)

	subjects, err = repo.ListSubjects(context.Background(), index.ID, model.SubjectTypeAll, 20, 0)
	require.NoError(t, err)
	require.Len(t, subjects, 1)
	require.EqualValues(t, subjects[0].Subject.ID, 5)
	require.EqualValues(t, subjects[0].Comment, "test22222")
}

func TestMysqlRepo_AddExists(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	index := &model.Index{
		ID:          0,
		Title:       "test",
		Description: "Test Index",
		CreatorID:   382951,
		CreatedAt:   time.Now(),
		Total:       0,
		Comments:    0,
		Collects:    0,
		NSFW:        false,
	}
	ctx := context.Background()

	_ = repo.New(ctx, index)

	subject, err := repo.AddOrUpdateIndexSubject(ctx, index.ID, 5, 5, "test")
	require.NoError(t, err)
	require.EqualValues(t, subject.Comment, "test")

	subject, err = repo.AddOrUpdateIndexSubject(ctx, index.ID, 5, 5, "test2")
	require.NoError(t, err)
	require.EqualValues(t, subject.Comment, "test2")
}

func TestMysqlRepo_AddNoneExistsSubject(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	ctx := context.Background()

	_, err := repo.AddOrUpdateIndexSubject(ctx, 15045, 999999999, 5, "test")
	require.Error(t, err)
	require.Equal(t, err, gerr.ErrSubjectNotFound)
}

func TestMysqlRepo_AddIndexCollect(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	ctx := context.Background()

	err := repo.AddIndexCollect(ctx, 15465, 233)
	require.NoError(t, err)

	// test if it added
	i, err := repo.GetIndexCollect(ctx, 15465, 233)
	require.NoError(t, err)

	require.EqualValues(t, 15465, i.IndexID)
	require.EqualValues(t, 233, i.UserID)
}

func TestMysqlRepo_GetIndexCollect(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	err := repo.AddIndexCollect(context.Background(), 15465, 2233)
	require.NoError(t, err)

	i, err := repo.GetIndexCollect(context.Background(), 15465, 2233)
	require.NoError(t, err)

	require.EqualValues(t, 15465, i.IndexID)
	require.EqualValues(t, 2233, i.UserID)
}

func TestMysqlRepo_DeleteIndexCollect(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)

	ctx := context.Background()

	err := repo.AddIndexCollect(ctx, 15465, 322)
	require.NoError(t, err)

	err = repo.DeleteIndexCollect(ctx, 15465, 322)
	require.NoError(t, err)
}
