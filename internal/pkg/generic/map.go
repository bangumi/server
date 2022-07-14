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

package generic

// MapKeys return []key of map, random ordered.
func MapKeys[K comparable, V any](m map[K]V) []K {
	var s = make([]K, 0, len(m))
	for k := range m {
		s = append(s, k)
	}

	return s
}

func MapValues[K comparable, V any](m map[K]V) []V {
	var s = make([]V, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}

	return s
}

func MapIter[K comparable, V any, T any](m map[K]V, fn func(key K, value V) T) []T {
	var s = make([]T, 0, len(m))
	for k, v := range m {
		s = append(s, fn(k, v))
	}

	return s
}
