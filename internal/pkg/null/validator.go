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
	"reflect"

	"github.com/go-playground/validator/v10"
)

var _ validator.CustomTypeFunc = Validator

// Validator implements validator.CustomTypeFunc.
func Validator(field reflect.Value) interface{} {
	if valuer, ok := field.Interface().(interface{ Interface() interface{} }); ok {
		return valuer.Interface()
	}

	return nil
}
