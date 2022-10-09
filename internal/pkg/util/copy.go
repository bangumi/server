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

package util

import (
	"reflect"
)

func CopySameNameField(dst, src any) {
	rvs, ok := truncatePtr(reflect.ValueOf(src))
	if !ok {
		return
	}
	rts := rvs.Type()

	rvd, ok := truncatePtr(reflect.ValueOf(dst))
	if !ok {
		return
	}

	for i := 0; i < rts.NumField(); i++ {
		fieldSrc := rts.Field(i)
		fieldDst := rvd.FieldByName(fieldSrc.Name)
		if !fieldDst.IsValid() {
			continue
		}
		if !sameSimpleType(fieldDst.Type(), fieldSrc.Type) {
			continue
		}
		fieldDst.Set(rvs.Field(i))
	}
}

// only simple type and single level ptr is considered equal for [util.CopySameNameField].
//
// true:
// sameSimpleType(reflect.Type(int), reflect.Type(int))
// sameSimpleType(reflect.Type(*string), reflect.Type(*string))
// sameSimpleType(reflect.Type(*int), reflect.Type(*int))
//
// false
// sameSimpleType(reflect.Type(struct{}), ...)
// sameSimpleType(reflect.Type(map[...]...), ...)
// sameSimpleType(reflect.Type([]...), ...)
// sameSimpleType(reflect.Type(int), reflect.Type(string))
// sameSimpleType(reflect.Type(*int), reflect.Type(*string))
// sameSimpleType(reflect.Type(**int), reflect.Type(**int)).
func sameSimpleType(t1, t2 reflect.Type) bool {
	if t1.Kind() != t2.Kind() {
		return false
	}

	// same type

	switch t1.Kind() {
	case reflect.Struct, reflect.Array, reflect.Map:
		return false
	case reflect.Ptr:
		// t1 and t2 is ptr
		switch t1.Elem().Kind() {
		case reflect.Struct, reflect.Array, reflect.Map, reflect.Ptr:
			return false
		default:
			return t1.Elem().Kind() == t2.Elem().Kind()
		}
	default:
		return true
	}
}

func truncatePtr(value reflect.Value) (reflect.Value, bool) {
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return value, false
		}
		value = value.Elem()
	}
	return value, true
}
