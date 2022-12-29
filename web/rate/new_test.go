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

package rate_test

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/test"
	"github.com/bangumi/server/web/rate"
	"github.com/bangumi/server/web/rate/action"
)

func flushDB(t *testing.T, db *redis.Client) {
	t.Helper()
	test.RunAndCleanup(t, func() { require.NoError(t, db.FlushDB(context.Background()).Err()) })
}

//nolint:paralleltest
func TestRateLimitManager_action(t *testing.T) {
	test.RequireEnv(t, "redis")
	db := test.GetRedis(t)
	flushDB(t, db)

	const uid model.UserID = 6
	r := rate.New(db)

	allowed, remain, err := r.AllowAction(context.TODO(), uid, action.Unknown, rate.PerHour(10))
	require.NoError(t, err)
	require.True(t, allowed)
	require.EqualValues(t, 9, remain)
}

//nolint:paralleltest
func TestRateLimitManager_Allowed(t *testing.T) {
	test.RequireEnv(t, "redis")

	db := test.GetRedis(t)
	flushDB(t, db)

	const ip = "0.0.0.-0"

	a, err := db.Exists(context.TODO(), rate.RedisRateKeyPrefix+ip).Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), a)

	a, err = db.Exists(context.TODO(), rate.RedisBanKeyPrefix+ip).Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), a)

	rateLimiter := rate.New(db)

	allowed, remain, err := rateLimiter.Login(context.TODO(), ip)
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, 4, remain)

	allowed, remain, err = rateLimiter.Login(context.TODO(), ip)
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, 3, remain)

	allowed, remain, err = rateLimiter.Login(context.TODO(), ip)
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, 2, remain)

	allowed, remain, err = rateLimiter.Login(context.TODO(), ip)
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, 1, remain)

	allowed, remain, err = rateLimiter.Login(context.TODO(), ip)
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, 0, remain)

	allowed, remain, err = rateLimiter.Login(context.TODO(), ip)
	require.NoError(t, err)
	require.False(t, allowed)
	require.Equal(t, 0, remain)

	allowed, remain, err = rateLimiter.Login(context.TODO(), ip)
	require.NoError(t, err)
	require.False(t, allowed)
	require.Equal(t, 0, remain)

	allowed, remain, err = rateLimiter.Login(context.TODO(), ip)
	require.NoError(t, err)
	require.False(t, allowed)
	require.Equal(t, 0, remain)
}
