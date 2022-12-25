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

package slice

func Map[T any, K any, F func(item T) K](in []T, fn F) []K {
	var s = make([]K, len(in))
	for i, t := range in {
		s[i] = fn(t)
	}

	return s
}

func MapFilter[T any, K any, F func(item T) (k K, ok bool)](in []T, fn F) []K {
	var s = make([]K, 0, len(in))
	for _, t := range in {
		v, ok := fn(t)
		if ok {
			s = append(s, v)
		}
	}

	return s
}

func ToMap[K comparable, T any, F func(item T) K](in []T, fn F) map[K]T {
	var s = make(map[K]T, len(in))
	for _, t := range in {
		s[fn(t)] = t
	}

	return s
}

func Any[T any, F func(item T) bool](in []T, fn F) bool {
	for _, t := range in {
		if fn(t) {
			return true
		}
	}

	return false
}

func All[T any, F func(item T) bool](in []T, fn F) bool {
	for _, t := range in {
		if !fn(t) {
			return false
		}
	}

	return true
}

// Clone is an alias of go builtin copy.
func Clone[S ~[]E, E any](s S) S {
	// Preserve nil in case it matters.
	if s == nil {
		return nil
	}

	var out = make(S, len(s))
	copy(out, s)

	return out
}
