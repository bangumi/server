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

package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/pkg/cache"
)

type MemCacheTestItem struct {
	V int
}

func Test_MemCache_HappyPath(t *testing.T) {
	t.Parallel()
	const key = "K"
	var value = MemCacheTestItem{V: 1}

	m := cache.NewMemoryCache[string, MemCacheTestItem]()

	m.Set(context.Background(), key, value, time.Hour)

	result, ok := m.Get(context.Background(), key)
	require.True(t, ok, "should get item")
	require.Equal(t, value.V, result.V)
}

func Test_MemCache_Expired(t *testing.T) {
	t.Parallel()
	const key = "K"
	m := cache.NewMemoryCache[string, MemCacheTestItem]()

	m.Set(context.Background(), key, MemCacheTestItem{}, time.Duration(0))

	time.Sleep(time.Millisecond * 100)

	_, ok := m.Get(context.Background(), key)
	require.False(t, ok)
}
