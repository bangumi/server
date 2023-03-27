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

package driver

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/internal/pkg/logger"
)

const defaultRedisPoolSize = 4

// NewRedisClient create a redis client
// use [test.GetRedis] in tests.
func NewRedisClient(c config.AppConfig) (*redis.Client, error) {
	redisOptions, err := redis.ParseURL(c.RedisURL)
	if err != nil {
		logger.Fatal("redis: failed to parse redis url", zap.String("url", c.RedisURL))
	}

	if redisOptions.PoolSize == 0 {
		redisOptions.PoolSize = defaultRedisPoolSize
	}

	cli := redis.NewClient(redisOptions)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, errgo.Wrap(err, "redis: failed to ping")
	}

	cli.AddHook(metrics.RedisHook(redisOptions.Addr))

	return cli, nil
}
