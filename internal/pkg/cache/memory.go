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
	"sync"
	"time"
)

// NewMemoryCache 不对缓存的对象进行序列化，直接用 [sync.Map] 保存在内存里。
//
// 过期的缓存不会从内存中自动回收，不能用来缓存值空间非常大的数据如条目或用户，
// 用于缓存用户组权限这样的值空间比较小的数据。
func NewMemoryCache[K comparable, V any]() *MemoryCache[K, V] {
	return &MemoryCache[K, V]{}
}

// MemoryCache store data in memory,
// will be used to cache user group permission rule.
type MemoryCache[K comparable, T any] struct {
	m sync.Map
}

type cacheItem[T any] struct {
	Value T
	Dead  time.Time
}

func (c *MemoryCache[K, T]) Get(_ context.Context, key K) (T, bool) {
	v, ok := c.m.Load(key)
	if !ok {
		return c.zero(), false
	}

	item, ok := v.(cacheItem[T])
	if !ok {
		panic("can't cast MemCache cache item")
	}

	if time.Now().After(item.Dead) {
		c.m.Delete(key)
		return c.zero(), false
	}

	return item.Value, true
}

func (c *MemoryCache[K, T]) Set(_ context.Context, key K, value T, ttl time.Duration) {
	c.m.Store(key, cacheItem[T]{
		Value: value,
		Dead:  time.Now().Add(ttl),
	})
}

func (c *MemoryCache[K, T]) zero() T {
	var t T
	return t
}
