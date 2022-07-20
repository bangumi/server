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

package gmap

// Keys return []key of map, random ordered.
func Keys[M map[K]V, K comparable, V any](m M) []K {
	var s = make([]K, 0, len(m))
	for k := range m {
		s = append(s, k)
	}

	return s
}

func Values[M map[K]V, K comparable, V any](m M) []V {
	var s = make([]V, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}

	return s
}

func Copy[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) {
	for k, v := range src {
		dst[k] = v
	}
}

func Map[K1, K2 comparable, V1, V2 any, M1 ~map[K1]V1, M2 map[K2]V2, F func(K1, V1) (K2, V2)](src M1, fn F) M2 {
	var m = make(M2, len(src))

	for k, v := range src {
		key, value := fn(k, v)
		m[key] = value
	}

	return m
}
