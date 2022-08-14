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

package slice_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bangumi/server/internal/pkg/generic/slice"
)

func TestMap(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"1", "2", "3", "4"}, slice.Map([]int{1, 2, 3, 4}, strconv.Itoa))
}

func TestToMap(t *testing.T) {
	t.Parallel()

	require.Equal(t, map[string]int{"1": 1, "2": 2, "3": 3, "4": 4}, slice.ToMap([]int{1, 2, 3, 4}, strconv.Itoa))
}

func TestMapFilter(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"2", "4"}, slice.MapFilter([]int{1, 2, 3, 4}, func(x int) (string, bool) {
		return strconv.Itoa(x), x%2 == 0
	}))

	require.Equal(t, []string{}, slice.MapFilter([]int{1, 2, 3, 4}, func(x int) (string, bool) {
		return "", false
	}))
}

func TestReduce(t *testing.T) {
	t.Parallel()

	require.Equal(t, []int{1, 2, 3, 3, 6}, slice.Flat([][]int{
		{1},
		{2},
		{3, 3},
		{6},
	}))
}

func TestAll(t *testing.T) {
	t.Parallel()

	require.True(t, slice.All([]int{1, 2, 3, 3, 6}, func(i int) bool {
		return true
	}))

	require.False(t, slice.All([]int{0, 1, 2, 3, 4}, func(i int) bool {
		return []bool{0: true, 1: false}[i]
	}))
}

func TestAny(t *testing.T) {
	t.Parallel()

	require.False(t, slice.Any([]int{1, 2, 3, 3, 6}, func(i int) bool {
		return false
	}))

	require.True(t, slice.Any([]int{0, 1, 2, 3, 4}, func(i int) bool {
		return []bool{0: true}[i]
	}))
}

func TestContain(t *testing.T) {
	t.Parallel()

	require.False(t, slice.Contain([]int{1, 0, 0, 0}, 2))

	require.True(t, slice.Contain([]int{1, 0, 0, 0}, 1))

	require.True(t, slice.Contain([]int{1, 2, 3, 4}, 4))
}
