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

package set_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/bangumi/server/internal/pkg/generic/set"
)

func checkeq[K comparable](t *testing.T, s set.Set[K], get func(k K) bool) {
	t.Helper()
	s.Each(func(key K) {
		if !get(key) {
			t.Fatalf("value %v should be in the set", key)
		}
	})
}

func TestCrossCheck(t *testing.T) {
	t.Parallel()
	stdm := make(map[int]bool)
	set := set.New[int]()

	const nops = 1000
	for i := 0; i < nops; i++ {
		op := rand.Intn(2)
		switch op {
		case 0:
			key := rand.Int()
			stdm[key] = true
			set.Add(key)
		case 1:
			var del int
			for k := range stdm {
				del = k
				break
			}
			delete(stdm, del)
			set.Remove(del)
		}

		checkeq(t, set, func(k int) bool {
			_, ok := stdm[k]
			return ok
		})
	}
}

func TestOf(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name  string
		input []string
	}{
		{"init with several items", []string{"foo", "bar", "baz"}},
		{"init without values", []string{}},
	}
	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			s := set.FromSlice[string](tc.input)

			if len(tc.input) != s.Len() {
				t.Fatalf("expected %d elements in set, got %d", len(tc.input), s.Len())
			}
			for _, val := range tc.input {
				if !s.Has(val) {
					t.Fatalf("expected to find val '%s' in set but did not", val)
				}
			}
		})
	}
}

func Example() {
	s := set.New[string]()
	s.Add("foo")
	s.Add("bar")
	s.Add("baz")

	fmt.Println(s.Has("foo"))
	fmt.Println(s.Has("quux"))
	// Output:
	// true
	// false
}
