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

var _ json.Unmarshaler = (*Type)(nil)
var _ iface = Type{}

// Type is a nullable type.
type Type struct {
	Value GenericType
	Set   bool // if json object has this field
	Null  bool // if json field's value is `null`
}

// NewType creates a new GenericType.
func NewType(t GenericType) Type {
	return Type{
		Null:  false,
		Value: t,
		Set:   true,
	}
}

func (t Type) HasValue() bool {
	return t.Set && !t.Null
}

func (t Type) Ptr() *GenericType {
	if t.Set && !t.Null {
		return &t.Value
	}

	return nil
}

// Default return default value its value is Null or not Set.
func (t Type) Default(v GenericType) GenericType {
	if t.Set && !t.Null {
		return t.Value
	}

	return v
}

func (t Type) Interface() any {
	if t.HasValue() {
		return t.Value
	}

	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Type) UnmarshalJSON(data []byte) error {
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
