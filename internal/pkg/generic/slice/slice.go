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

func Map[T any, K any](in []T, fn func(item T) K) []K {
	var s = make([]K, len(in))
	for i, t := range in {
		s[i] = fn(t)
	}

	return s
}

func MapFilter[T any, K any](in []T, fn func(item T) (k K, ok bool)) []K {
	var s = make([]K, 0, len(in))
	for _, t := range in {
		v, ok := fn(t)
		if ok {
			s = append(s, v)
		}
	}

	return s
}

func First[T any](in []T, end int) []T {
	if len(in) < end {
		end = len(in)
	}

	out := make([]T, end)
	copy(out, in)
	return out
}
