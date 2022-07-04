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

package generic

type Number interface {
	Integer | Float
}

type Signed interface {
	int8 | int16 | int32 | int64 | int
}

type Unsigned interface {
	uint8 | uint16 | uint32 | uint64 | uint
}

type Integer interface {
	Signed | Unsigned
}

type Float interface {
	float32 | float64
}

func Min[T Number](a, b T) T {
	if a < b {
		return a
	}

	return b
}
