// Code generated by ./internal/cmd/gen/null/main.go. DO NOT EDIT.

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

var _ json.Unmarshaler = (*Uint8)(nil)
var _ iface = Uint8{}

// Uint8 is a nullable type.
type Uint8 struct {
	Value uint8
	Set   bool // if json object has this field
	Null  bool // if json field's value is `null`
}

// NewUint8 creates a new uint8.
func NewUint8(t uint8) Uint8 {
	return Uint8{
		Null:  false,
		Value: t,
		Set:   true,
	}
}

func (t Uint8) HasValue() bool {
	return t.Set && !t.Null
}

func (t Uint8) Ptr() *uint8 {
	if t.Set && !t.Null {
		return &t.Value
	}

	return nil
}

// Default return default value its value is Null or not Set.
func (t Uint8) Default(v uint8) uint8 {
	if t.Set && !t.Null {
		return t.Value
	}

	return v
}

func (t Uint8) Interface() any {
	if t.HasValue() {
		return t.Value
	}

	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *Uint8) UnmarshalJSON(data []byte) error {
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