// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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

package rate

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/gookit/goutil/timex"

	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/web/rate/redisrate"
)

const defaultAllowPerHour = 5

type Manager interface {
	// Allowed 检查是否允许登录。
	Allowed(ctx context.Context, ip string) (allowed bool, remain int, err error)
	// Reset 登录成功时应该重置计数。
	Reset(ctx context.Context, ip string) error
}

const RedisRateKeyPrefix = "chii:login:rate:"
const RedisBanKeyPrefix = "chii:ban:"

func New(rdb *redis.Client) Manager {
	return manager{
		r: rdb,
		l: redisrate.NewLimiter(rdb),
	}
}

type manager struct {
	r *redis.Client
	l *redisrate.Limiter
}

func (m manager) Allowed(ctx context.Context, ip string) (bool, int, error) {
	// TODO: replace this function with a single lua script.
	var banKey = RedisBanKeyPrefix + ip
	result, err := m.r.Exists(ctx, banKey, "1").Result()
	if err != nil {
		return false, 0, errgo.Wrap(err, "redis.Exists")
	}

	if result == 1 {
		return false, 0, nil
	}

	res, err := m.l.Allow(ctx, RedisRateKeyPrefix+ip, redisrate.PerHour(defaultAllowPerHour))
	if err != nil {
		return false, 0, errgo.Wrap(err, "Limiter.Allow")
	}

	if res.Allowed <= 0 {
		err := m.r.Set(ctx, banKey, "1", timex.OneWeek).Err()

		return false, 0, errgo.Wrap(err, "redis.Set")
	}

	return true, res.Remaining, nil
}

func (m manager) Reset(ctx context.Context, ip string) error {
	err := m.l.Reset(ctx, RedisRateKeyPrefix+ip)

	return errgo.Wrap(err, "Limiter.Allow")
}
