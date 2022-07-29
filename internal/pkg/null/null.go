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
	"github.com/goccy/go-json"
)

var _ json.Unmarshaler = (*Null[bool])(nil)

// Null is a nullable type.
type Null[T any] struct {
	Value T
	Set   bool // if json object has this field
	Null  bool // if json field's value is `null`
}

func (t Null[T]) HasValue() bool {
	return t.Set && !t.Null
}

func (t Null[T]) Ptr() *T {
	if t.Set && !t.Null {
		return &t.Value
	}

	return nil
}

func (t Null[T]) Interface() any {
	if t.Set && !t.Null {
		return &t.Value
	}

	return nil
}

// Default return default value its value is Null or not Set.
func (t Null[T]) Default(v T) T {
	if t.Null && t.Set {
		return t.Value
	}

	return v
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Null[T]) UnmarshalJSON(data []byte) error {
	t.Set = true

	if string(data) == "null" {
		t.Null = true
		return nil
	}

	if err := json.UnmarshalNoEscape(data, &t.Value); err != nil {
		return err //nolint:wrapcheck
	}

	return nil
}
