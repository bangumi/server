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

import "reflect"

func CopySameNameField(dst interface{}, src interface{}) {
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
		fieldDst.Set(rvs.Field(i))
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
