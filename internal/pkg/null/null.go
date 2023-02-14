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

package null

import (
	"encoding/json"
)

var _ json.Unmarshaler = (*Null[any])(nil)

// Null is a nullable type.
type Null[T any] struct {
	Value T
	Set   bool
}

func New[T any](t T) Null[T] {
	return Null[T]{
		Value: t,
		Set:   true,
	}
}

func NewFromPtr[T any](p *T) Null[T] {
	if p == nil {
		return Null[T]{}
	}

	return Null[T]{
		Value: *p,
		Set:   true,
	}
}

func (t Null[T]) Ptr() *T {
	if t.Set {
		return &t.Value
	}

	return nil
}

// Default return default value its value is Null or not Set.
func (t Null[T]) Default(v T) T {
	if t.Set {
		return t.Value
	}

	return v
}

func (t Null[T]) Interface() any {
	if t.Set {
		return t.Value
	}

	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Null[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	t.Set = true
	return json.Unmarshal(data, &t.Value) //nolint:wrapcheck
}
