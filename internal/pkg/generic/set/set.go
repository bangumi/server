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

package set

type empty struct {
}

func FromSlice[T comparable](in []T) Set[T] {
	var s = Set[T]{
		m: make(map[T]empty, len(in)),
	}

	for _, t := range in {
		s.m[t] = empty{}
	}

	return s
}

type Set[T comparable] struct {
	m map[T]empty
}

func (s Set[T]) Has(item T) bool {
	_, ok := s.m[item]
	return ok
}
