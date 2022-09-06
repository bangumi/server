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

	"github.com/bangumi/server/internal/pkg/errgo"
)

// NewMemoryCache 不对缓存的对象进行序列化，直接用 [sync.Map] 保存在内存里。
//
// 过期的缓存不会从内存中自动回收，不能用来缓存值空间非常大的数据如条目或用户，
// 用于缓存用户组权限这样的值空间比较小的数据。
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{}
}

var errCacheNotSameType = errors.New("cached item have is not same type as expected result")

// MemoryCache store data in memory,
// will be used to cache user group permission rule.
type MemoryCache struct {
	m sync.Map
}

type cacheItem struct {
	Value any
	Dead  time.Time
}

func (c *MemoryCache) Get(_ context.Context, key string, value any) (bool, error) {
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

func (c *MemoryCache) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	c.m.Store(key, cacheItem{
		Value: value,
		Dead:  time.Now().Add(ttl),
	})

	return nil
}
