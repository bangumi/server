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
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
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

func Merge[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) M1 {
	for k, v := range src {
		dst[k] = v
	}

	return dst
}

func Has[M ~map[K]V, K comparable, V any](m M, key K) bool {
	_, ok := m[key]
	return ok
}

func Map[S ~map[SK]SV, SK comparable, SV any, F ~func(SK, SV) (DK, DV), DK comparable, DV any](src S, fn F) map[DK]DV {
	var dst = make(map[DK]DV, len(src))
	for key, value := range src {
		dstKey, dstValue := fn(key, value)
		dst[dstKey] = dstValue
	}

	return dst
}
