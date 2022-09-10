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
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/cache"
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
	encoded, err := json.Marshal(value)
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
	test.RequireEnv(t, "redis")

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
	test.RequireEnv(t, test.EnvRedis)

	var key = t.Name() + "redis_test "

	db := test.GetRedis(t)
	require.NoError(t, db.Set(context.Background(), key, "", 0).Err())

	c := cache.NewRedisCache(db)

	require.NoError(t, c.Del(context.Background(), key))

	v, err := db.Exists(context.Background(), key).Result()
	require.NoError(t, err)
	require.True(t, v == 0)
}

func TestRedisCache_GetMany(t *testing.T) {
	test.RequireEnv(t, "redis")
	t.Parallel()

	db := test.GetRedis(t)
	clearKeys(t, db, t.Name())

	c := cache.NewRedisCache(db)

	require.NoError(t, db.Set(context.Background(), t.Name()+"1",
		marshal(t, RedisCacheTestItem{I: 1}), time.Minute).Err())

	require.NoError(t, db.Set(context.Background(), t.Name()+"2",
		marshal(t, RedisCacheTestItem{I: 2}), time.Minute).Err())

	result := c.GetMany(context.TODO(), []string{t.Name() + "1", t.Name() + "2", t.Name() + "3"})
	require.NoError(t, result.Err)

	var unmarshalled map[int]RedisCacheTestItem
	unmarshalled, err := cache.UnmarshalMany(result, func(i RedisCacheTestItem) int { return i.I })
	require.NoError(t, err)
	require.Contains(t, unmarshalled, 1)
	require.Contains(t, unmarshalled, 2)

	require.Contains(t, result.Result, t.Name()+"1")
	require.Contains(t, result.Result, t.Name()+"2")
	require.NotContains(t, result.Result, t.Name()+"3")

	for i := 1; i <= 2; i++ {
		var r RedisCacheTestItem
		key := t.Name() + strconv.Itoa(i)
		require.Contains(t, result.Result, key)
		bytes := result.Result[key]
		require.NoError(t, json.Unmarshal(bytes, &r))
		require.Equal(t, i, r.I)
	}
}

func TestRedisCache_SetMany(t *testing.T) {
	test.RequireEnv(t, "redis")
	t.Parallel()

	db := test.GetRedis(t)
	clearKeys(t, db, t.Name())

	c := cache.NewRedisCache(db)

	require.NoError(t, c.SetMany(context.TODO(), map[string]any{
		t.Name() + "1": RedisCacheTestItem{I: 1},
		t.Name() + "2": RedisCacheTestItem{I: 2},
		t.Name() + "3": RedisCacheTestItem{I: 3},
		t.Name() + "4": RedisCacheTestItem{I: 4},
	}, time.Minute))

	for i := 1; i <= 4; i++ {
		var result RedisCacheTestItem
		ok, err := c.Get(context.TODO(), t.Name()+strconv.Itoa(i), &result)
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, i, result.I)
	}
}

func clearKeys(t *testing.T, r *redis.Client, prefix string) {
	t.Helper()
	test.RunAndCleanup(t, func() {
		keys, err := r.Keys(context.TODO(), prefix+"*").Result()
		require.NoError(t, err)

		if len(keys) == 0 {
			return
		}

		err = r.Del(context.TODO(), keys...).Err()
		require.NoError(t, err)
	})
}

func marshal(t *testing.T, v any) []byte {
	t.Helper()

	p, err := json.Marshal(v)
	require.NoError(t, err)

	return p
}
