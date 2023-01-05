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

package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/test"
)

type RedisCacheTestItem struct {
	S string
	I int
}

func mockedCache() (cache.RedisCache, redismock.ClientMock) {
	db, mock := redismock.NewClientMock()
	c := cache.NewRedisCache(db)

	return c, mock
}

func TestRedisCache_Set(t *testing.T) {
	t.Parallel()
	var key = t.Name() + "redis_key"
	c, mock := mockedCache()
	mock.Regexp().ExpectSet(key, `.*`, time.Hour).SetVal("OK")

	value := RedisCacheTestItem{
		S: "sss",
		I: 2,
	}

	require.NoError(t, c.Set(context.TODO(), key, value, time.Hour))

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedisCache_Get_Nil(t *testing.T) {
	t.Parallel()

	var key = t.Name() + "redis_key"
	c, mock := mockedCache()
	mock.Regexp().ExpectGet(key).RedisNil()

	var result RedisCacheTestItem

	ok, err := c.Get(context.TODO(), key, &result)
	require.NoError(t, err)
	require.False(t, ok)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedisCache_Get_Cached(t *testing.T) {
	t.Parallel()

	var key = t.Name() + "redis_key"
	value := RedisCacheTestItem{
		S: "sss",
		I: 2,
	}

	c, mock := mockedCache()
	encoded, err := sonic.Marshal(value)
	require.NoError(t, err)

	mock.Regexp().ExpectGet(key).SetVal(string(encoded))

	var result RedisCacheTestItem

	ok, err := c.Get(context.TODO(), key, &result)
	require.NoError(t, err)
	require.True(t, ok)

	require.Equal(t, value, result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedisCache_Get_Broken(t *testing.T) {
	t.Parallel()

	var key = t.Name() + "redis_key"
	c, mock := mockedCache()

	mock.Regexp().ExpectGet(key).SetVal("some random broken content")
	mock.Regexp().ExpectDel(key).SetVal(1)

	var result RedisCacheTestItem

	ok, err := c.Get(context.TODO(), key, &result)
	require.NoError(t, err)
	require.False(t, ok)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRedisCache_Real(t *testing.T) {
	t.Parallel()

	var key = t.Name() + "redis_key"

	db := test.GetRedis(t)
	db.Del(context.TODO(), key)

	c := cache.NewRedisCache(db)

	var data = RedisCacheTestItem{S: "ss", I: 5}
	require.NoError(t, c.Set(context.TODO(), key, data, time.Hour))

	var result RedisCacheTestItem
	ok, err := c.Get(context.TODO(), key, &result)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, data, result)
}

func TestRedisCache_Del(t *testing.T) {
	t.Parallel()

	var key = t.Name() + "redis_test "

	db := test.GetRedis(t)
	require.NoError(t, db.Set(context.Background(), key, "", 0).Err())

	c := cache.NewRedisCache(db)

	require.NoError(t, c.Del(context.Background(), key))

	v, err := db.Exists(context.Background(), key).Result()
	require.NoError(t, err)
	require.True(t, v == 0)
}
