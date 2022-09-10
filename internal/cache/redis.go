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
	_ "embed"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
)

//go:embed mset.lua
var setManyLua string

var setManyScript = redis.NewScript(setManyLua) //nolint:gochecknoglobals

type GetManyResult struct {
	cache  RedisCache
	Result map[string][]byte
	Err    error
}

// RedisCache
//
//	var s model.Subject
//	c.Get(ctx, key, &s)
//	c.Set(ctx, key, s, time.Minute)
//
// SetMany...
//
//	err := c.SetMany(ctx, cache.MarshalMany(notCachedSubjects, cachekey.Subject), time.Minute)
//
// GetMany...
//
//	result := c.GetMany(ctx, slice.Map(subjectIDs, cachekey.Subject))
//	var cached map[model.SubjectID]model.Subject
//	cached, err = cache.UnmarshalMany(result, model.Subject.GetID)
type RedisCache interface {
	Get(ctx context.Context, key string, value any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	SetMany(ctx context.Context, data map[string]any, ttl time.Duration) error
	GetMany(ctx context.Context, keys []string) GetManyResult
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

	err = unmarshalBytes(raw, value)
	if err != nil {
		logger.Warn("can't unmarshal redis cached data as json", zap.String("key", key))
		c.r.Del(ctx, key)

		return false, nil
	}

	return true, nil
}

func (c redisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	b, err := marshalBytes(value)
	if err != nil {
		return err
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

func (c redisCache) GetMany(ctx context.Context, keys []string) GetManyResult {
	values, err := c.r.MGet(ctx, keys...).Result()
	if err != nil {
		return GetManyResult{Err: errgo.Wrap(err, "redis set")}
	}

	var result = make(map[string][]byte, len(keys))
	for i, value := range values {
		if value == nil {
			continue
		}

		switch v := value.(type) {
		case string:
			result[keys[i]] = []byte(v)
		default:
			return GetManyResult{
				Err: fmt.Errorf("BUG: unexpected redis response type %T %+v", value, value), //nolint:goerr113
			}
		}
	}

	return GetManyResult{Result: result, cache: c}
}

func (c redisCache) SetMany(ctx context.Context, data map[string]any, ttl time.Duration) error {
	var keys = make([]string, 0, len(data)) // [     key1,   key2,   key3,   ...]
	var args = make([]any, 0, len(data)+1)  // [ttl, bytes1, bytes2, bytes3, ...]

	args = append(args, int64(ttl.Seconds()))

	for key, value := range data {
		b, err := marshalBytes(value)
		if err != nil {
			return err
		}

		keys = append(keys, key)
		args = append(args, b)
	}

	if err := setManyScript.Run(ctx, c.r, keys, args...).Err(); err != nil {
		return errgo.Wrap(err, "redis set")
	}

	return nil
}

func MarshalMany[M ~map[K]V, K comparable, V any, F func(t K) string](data M, fn F) map[string]any {
	var out = make(map[string]any, len(data))

	for key, value := range data {
		out[fn(key)] = value
	}

	return out
}

func UnmarshalMany[T any, ID comparable, F func(t T) ID](result GetManyResult, fn F) (map[ID]T, error) {
	if result.Err != nil {
		return nil, result.Err
	}

	var out = make(map[ID]T, len(result.Result))

	var badKeys = make([]string, 0, len(result.Result))

	for key, bytes := range result.Result {
		var t T
		err := unmarshalBytes(bytes, &t)
		if err != nil {
			logger.Warn("bad cached bytes", zap.String("key", key), zap.ByteString("value", bytes))
			badKeys = append(badKeys, key)
			continue
		}

		out[fn(t)] = t
	}

	if len(badKeys) != 0 && (result.cache != nil) {
		go func() {
			err := result.cache.Del(context.Background(), badKeys...)
			if err != nil {
				logger.Error("failed to delete bad cache from redis in the background", zap.Error(err))
			}
		}()
	}

	return out, nil
}
