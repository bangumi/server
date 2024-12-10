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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/test"
)

type RedisCacheTestItem struct {
	S string
	I int
}

func TestRedisCache_Real(t *testing.T) {
	t.Parallel()

	var key = t.Name() + "redis_key"

	r := test.GetRedis(t)
	require.NoError(t, r.Do(context.TODO(), r.B().Del().Key(key).Build()).Error())

	c := cache.NewRedisCache(r)

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

	var key = fmt.Sprintln(t.Name(), "redis_test", time.Now())

	r := test.GetRedis(t)
	require.NoError(t, r.Do(context.TODO(), r.B().Set().Key(key).Value("").Build()).Error())

	c := cache.NewRedisCache(r)

	require.NoError(t, c.Del(context.Background(), key))

	exist, err := r.Do(context.TODO(), r.B().Exists().Key(key).Build()).AsBool()
	require.NoError(t, err)
	require.False(t, exist)
}
