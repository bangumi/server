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

package gstr

import (
	"strconv"

	"github.com/trim21/errgo"
)

func ParseInt8(s string) (int8, error) {
	v, err := strconv.ParseInt(s, 10, 8)

	return int8(v), errgo.Wrap(err, "strconv")
}

func ParseUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)

	return uint8(v), errgo.Wrap(err, "strconv")
}

func ParseUint16(s string) (uint16, error) {
	v, err := strconv.ParseUint(s, 10, 16)

	return uint16(v), errgo.Wrap(err, "strconv")
}

func ParseInt32(s string) (int32, error) {
	v, err := strconv.ParseInt(s, 10, 32)

	return int32(v), errgo.Wrap(err, "strconv")
}

func ParseUint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)

	return uint32(v), errgo.Wrap(err, "strconv")
}

func ParseBool(s string) (bool, error) {
	v, err := strconv.ParseBool(s)

	return v, errgo.Wrap(err, "strconv")
}
