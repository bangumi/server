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

// Type is a nullable type.
type Type struct {
	Value GenericType
	Set   bool // if json object has this field
}

// NewType creates a new GenericType.
func NewType(v GenericType) Type {
	return Type{
		Value: v,
		Set:   true,
	}
}

func (t Type) Ptr() *GenericType {
	if t.Set {
		return &t.Value
	}

	return nil
}

// Default return default value its value is Null or not Set.
func (t Type) Default(v GenericType) GenericType {
	if t.Set {
		return t.Value
	}

	return v
}

func (t Type) Interface() any {
	if t.Set {
		return t.Value
	}

	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Type) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	t.Set = true
	return json.UnmarshalNoEscape(data, &t.Value) //nolint:wrapcheck
}
