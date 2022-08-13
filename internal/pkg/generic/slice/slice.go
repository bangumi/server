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

import "database/sql/driver"

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

func ToUint8[T interface{ ~uint8 }](in []T) []uint8 {
	var s = make([]uint8, len(in))
	for i, t := range in {
		s[i] = uint8(t)
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

func ToValuer[T driver.Valuer](in []T) []driver.Valuer {
	var s = make([]driver.Valuer, len(in))
	for i, t := range in {
		s[i] = t
	}

	return s
}

func Flat[T any](in [][]T) []T {
	var c int
	for _, ts := range in {
		c += len(ts)
	}

	var out = make([]T, 0, c)
	for _, ts := range in {
		out = append(out, ts...)
	}

	return out
}

func UniqueUnsorted[T comparable](in []T) []T {
	var m = make(map[T]struct{}, len(in))

	for _, t := range in {
		m[t] = struct{}{}
	}

	var out = make([]T, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	return out
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

func Contains[T comparable](s []T, item T) bool {
	for _, t := range s {
		if t == item {
			return true
		}
	}

	return false
}
