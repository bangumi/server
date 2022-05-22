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
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/bangumi/server/internal/errgo"
)

// NewMemoryCache return an in-memory cache.
// This cache backend should be used to cache limited-sized entries like user group permission rule.
func NewMemoryCache() Generic {
	return &memCache{}
}

var errCacheNotSameType = errors.New("cached item have is not same type as expected result")

// memCache store data in memory,
// will be used to cache user group permission rule.
type memCache struct {
	m sync.Map
}

type cacheItem struct {
	Value interface{}
	Dead  time.Time
}

func (c *memCache) Get(_ context.Context, key string, value interface{}) (bool, error) {
	v, ok := c.m.Load(key)
	if !ok {
		return ok, nil
	}

	item, ok := v.(cacheItem)
	if !ok {
		panic("can't cast MemCache cache item")
	}

	if time.Now().After(item.Dead) {
		c.m.Delete(key)

		return false, nil
	}

	vOut := reflect.ValueOf(value).Elem()
	vCache := reflect.ValueOf(item.Value)

	// vOut.Set(vCache) will panic if we don't check their type here.
	if !vCache.Type().AssignableTo(vOut.Type()) {
		return false, errgo.Wrap(errCacheNotSameType,
			fmt.Sprintf(
				"cached item is %s, but receiver it ptr of type %s", vOut.Type(), vCache.Type(),
			))
	}

	vOut.Set(vCache)

	return true, nil
}

func (c *memCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	c.m.Store(key, cacheItem{
		Value: value,
		Dead:  time.Now().Add(ttl),
	})

	return nil
}

func (c *memCache) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		c.m.Delete(key)
	}

	return nil
}
