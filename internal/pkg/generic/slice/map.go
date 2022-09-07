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

package slice

import (
	"database/sql/driver"
)

func ToUint8[S ~[]T, T ~uint8](s S) []uint8 {
	var out = make([]uint8, len(s))
	for i, t := range s {
		out[i] = uint8(t)
	}

	return out
}

func ToValuer[S ~[]T, T driver.Valuer](s S) []driver.Valuer {
	var out = make([]driver.Valuer, len(s))
	for i, t := range s {
		out[i] = t
	}

	return out
}
