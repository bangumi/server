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

package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
)

type RedisCache interface {
	Get(ctx context.Context, key string, value any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	// SetMany(ctx context.Context, keys string, values []any, ttl time.Duration) error
}

// NewRedisCache create a redis backed cache.
func NewRedisCache(cli *redis.Client) RedisCache {
	return redisCache{r: cli}
}

type redisCache struct {
	r *redis.Client
}

func (c redisCache) Get(ctx context.Context, key string, value any) (bool, error) {
	raw, err := c.r.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}

		return false, errgo.Wrap(err, "redis get")
	}

	err = json.Unmarshal(raw, value)
	if err != nil {
		logger.Warn("can't unmarshal redis cached data as json", zap.String("key", key))
		c.r.Del(ctx, key)

		return false, nil
	}

	return true, nil
}

func (c redisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	b, err := json.MarshalWithOption(value, json.DisableHTMLEscape())
	if err != nil {
		return errgo.Wrap(err, "json")
	}

	if err := c.r.Set(ctx, key, b, ttl).Err(); err != nil {
		return errgo.Wrap(err, "redis set")
	}

	return nil
}

func (c redisCache) Del(ctx context.Context, keys ...string) error {
	err := c.r.Del(ctx, keys...).Err()
	return errgo.Wrap(err, "redis.Del")
}
