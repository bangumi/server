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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/index"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
)

func getRepo(t *testing.T) domain.IndexRepo {
	t.Helper()
	repo, err := index.NewMysqlRepo(query.Use(test.GetGorm(t)), zap.NewNop())
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
		UpdateAt:    now,
		Total:       0,
		Comments:    0,
		Collects:    0,
		Ban:         false,
		NSFW:        false,
	}
	err := repo.New(context.Background(), index)
	require.NoError(t, err)
	require.NotEqualValues(t, 0, index.ID)

	i, err := repo.Get(context.Background(), index.ID)
	require.NoError(t, err)

	require.EqualValues(t, 382951, i.CreatorID)
	require.EqualValues(t, "test", i.Title)
	require.EqualValues(t, now.Truncate(time.Second), i.CreatedAt)
	require.EqualValues(t, now.Truncate(time.Second), i.UpdateAt)
}

func TestMysqlRepo_UpdateIndex(t *testing.T) {
	test.RequireEnv(t, test.EnvMysql)
	t.Parallel()

	repo := getRepo(t)
	i, err := repo.Get(context.Background(), 15045)

	defaultTitle := "日本动画最高收视率TOP100"
	defaultDesc := "[url]http://www.tudou.com/programs/view/" +
		"W6eIoxnHs6g/[/url]\r\n有美国动画混入，所以准确的说是在日本播放的" +
		"动画最高收视率（而且是关东地区的\r\n基本大部分是70年代的，那个年代娱乐贫乏优势真大"

	// check current
	require.NoError(t, err)
	require.EqualValues(t, defaultTitle, i.Title)
	require.EqualValues(t, defaultDesc, i.Description)

	// update index information
	err = repo.Update(context.Background(), 15045, "日本动画", "日本动画的介绍")
	require.NoError(t, err)

	// check updated index information
	i, err = repo.Get(context.Background(), 15045)
	require.NoError(t, err)
	require.EqualValues(t, "日本动画", i.Title)
	require.EqualValues(t, "日本动画的介绍", i.Description)

	// revert update
	err = repo.Update(context.Background(), 15045, defaultTitle, defaultDesc)
	require.NoError(t, err)
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
		Ban:         false,
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
	require.ErrorIs(t, err, domain.ErrNotFound)

	// TODO: all subjects in the index should be removed as well
}
