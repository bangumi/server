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

package driver

import (
	"context"

	"github.com/go-redis/redis/v8"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/errgo"
)

const defaultRedisPoolSize = 4

func NewRedisClient(c config.AppConfig) (*redis.Client, error) {
	if c.RedisOptions.PoolSize == 0 {
		c.RedisOptions.PoolSize = defaultRedisPoolSize
	}

	cli := redis.NewClient(c.RedisOptions)

	if err := cli.Ping(context.Background()).Err(); err != nil {
		return nil, errgo.Wrap(err, "failed to connect to redis")
	}

	return cli, nil
}
