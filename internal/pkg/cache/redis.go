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
	"reflect"
	"time"

	"github.com/redis/rueidis"
	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/pkg/logger"
)

// RedisCache
//
//	var s model.Subject
//	c.Get(ctx, key, &s)
//	c.Set(ctx, key, s, time.Minute)
type RedisCache interface {
	Get(ctx context.Context, key string, value any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error

	// MGet result should be `*[]T`
	// for example
	//
	//	var s []struct{}
	//	cache.MGet(ctx, keys, &s)
	MGet(ctx context.Context, key []string, result any) error
}

// NewRedisCache create a redis backed cache.
func NewRedisCache(ru rueidis.Client) RedisCache {
	return redisCache{ru: ru}
}

type redisCache struct {
	ru rueidis.Client
}

func (c redisCache) Get(ctx context.Context, key string, value any) (bool, error) {
	result := c.ru.Do(ctx, c.ru.B().Get().Key(key).Build())
	if err := result.NonRedisError(); err != nil {
		return false, errgo.Wrap(err, "redis get")
	}

	raw, err := result.AsBytes()
	// redis.Nil
	if err != nil {
		return false, nil
	}

	err = unmarshalBytes(raw, value)
	if err != nil {
		logger.Warn("can't unmarshal redis cached data as json", zap.String("key", key))

		c.ru.Do(ctx, c.ru.B().Del().Key(key).Build())

		return false, nil
	}

	return true, nil
}

// MGet result should be `*[]T`
// for example
//
//	var s []struct{}
//	cache.MGet(ctx, keys, &s)
func (c redisCache) MGet(ctx context.Context, keys []string, result any) error {
	results, err := c.ru.Do(ctx, c.ru.B().Mget().Key(keys...).Build()).ToArray()
	if err != nil {
		//nolint:errorlint
		if err == rueidis.Nil {
			return nil
		}

		return errgo.Wrap(err, "redis mget")
	}

	// no more ideal way to do this
	// reflect.ValueOf(*[]T)
	rv := reflect.ValueOf(result)
	if !(rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Slice) {
		panic("only allow *[]T as input")
	}

	// reflect.ValueOf([]T)
	vv := rv.Elem()

	vv.Set(reflect.MakeSlice(rv.Elem().Type(), 0, len(results)))
	elementType := rv.Elem().Type().Elem()
	for _, message := range results {
		var v = reflect.New(elementType)

		e := message.DecodeJSON(v.Interface())
		if e != nil {
			logger.Warn("unexpected failure when decoding json", zap.Error(e))
			continue
		}

		reflect.Append(vv, v.Elem())
	}

	return nil
}

func (c redisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	err := c.ru.Do(ctx, c.ru.B().Set().Key(key).Value(rueidis.JSON(value)).Ex(ttl).Build()).Error()
	if err != nil {
		return errgo.Wrap(err, "redis set")
	}

	return nil
}

func (c redisCache) Del(ctx context.Context, keys ...string) error {
	err := c.ru.Do(ctx, c.ru.B().Del().Key(keys...).Build()).Error()
	return errgo.Wrap(err, "redis.Del")
}
