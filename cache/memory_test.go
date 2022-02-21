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

package cache_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/cache"
)

type MemCacheTestItem struct {
	V int
}

func TestHappyPath(t *testing.T) {
	t.Parallel()
	const key = "K"
	var value = MemCacheTestItem{V: 1}

	m := cache.NewMemoryCache()

	require.NoError(t, m.Set(context.Background(), key, value, time.Hour))

	var result MemCacheTestItem

	ok, err := m.Get(context.Background(), key, &result)

	require.NoError(t, err)
	require.True(t, ok, "should get item")

	require.Equal(t, value.V, result.V)
}

func TestWrongType(t *testing.T) {
	t.Parallel()
	const key = "K"
	m := cache.NewMemoryCache()

	require.NoError(t, m.Set(context.Background(), key, struct{ F int }{}, time.Hour))

	var result MemCacheTestItem

	ok, err := m.Get(context.Background(), key, &result)
	require.NotNil(t, err)
	require.Regexp(t, regexp.MustCompile("not same type"), err)
	require.False(t, ok)
}

func TestExpired(t *testing.T) {
	t.Parallel()
	const key = "K"
	m := cache.NewMemoryCache()

	require.NoError(t, m.Set(context.Background(), key, MemCacheTestItem{}, time.Duration(0)))

	time.Sleep(time.Millisecond * 100)

	var result MemCacheTestItem

	ok, err := m.Get(context.Background(), key, &result)
	require.NoError(t, err)
	require.False(t, ok)
}
