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

type empty = struct {
}

// Set is a not thread-safe set based on built-in map.
type Set[T comparable] struct {
	m map[T]empty
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

func New[T comparable]() Set[T] {
	return Set[T]{
		m: map[T]empty{},
	}
}

func (s Set[T]) Has(item T) bool {
	_, ok := s.m[item]
	return ok
}

func (s Set[T]) Add(item T) {
	s.m[item] = empty{}
}

func (s Set[T]) Remove(item T) {
	delete(s.m, item)
}

func (s Set[T]) Removes(items ...T) Set[T] {
	for _, item := range items {
		delete(s.m, item)
	}

	return s
}

func (s Set[T]) Len() int {
	return len(s.m)
}

func (s Set[T]) ToSlice() []T {
	var out = make([]T, 0, len(s.m))
	for k := range s.m {
		out = append(out, k)
	}

	return out
}

// Union return a new set = (s1 | s2).
func (s Set[T]) Union(o Set[T]) Set[T] {
	ns := Set[T]{
		m: make(map[T]empty, len(s.m)+len(o.m)),
	}

	for e := range s.m {
		ns.m[e] = empty{}
	}

	for e := range o.m {
		ns.m[e] = empty{}
	}

	return ns
}

// Intersection return a new set = (s1 & s2).
func (s Set[T]) Intersection(o Set[T]) Set[T] {
	l := s.Len()
	if o.Len() > l {
		l = len(o.m)
	}

	ns := Set[T]{
		m: make(map[T]empty, l),
	}

	for e := range o.m {
		if s.Has(e) {
			ns.m[e] = empty{}
		}
	}

	return ns
}

func (s Set[K]) Each(fn func(key K)) {
	for k := range s.m {
		fn(k)
	}
}
