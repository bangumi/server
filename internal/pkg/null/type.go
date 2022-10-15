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

type Uint8 = Null[uint8]

func NewUint8(v uint8) Uint8 {
	return Null[uint8]{Set: true, Value: v}
}

type Uint16 = Null[uint16]

func NewUint16(v uint16) Uint16 {
	return Null[uint16]{Set: true, Value: v}
}

type Uint32 = Null[uint32]

func NewUint32(v uint32) Uint32 {
	return Null[uint32]{Set: true, Value: v}
}

type Uint64 = Null[uint64]

func NewUint64(v uint64) Uint64 {
	return Null[uint64]{Set: true, Value: v}
}

type Uint = Null[uint]

func NewUint(v uint) Uint {
	return Null[uint]{Set: true, Value: v}
}

type Int8 = Null[int8]

func NewInt8(v int8) Int8 {
	return Null[int8]{Set: true, Value: v}
}

type Int16 = Null[int16]

func NewInt16(v int16) Int16 {
	return Null[int16]{Set: true, Value: v}
}

type Int32 = Null[int32]

func NewInt32(v int32) Int32 {
	return Null[int32]{Set: true, Value: v}
}

type Int64 = Null[int64]

func NewInt64(v int64) Int64 {
	return Null[int64]{Set: true, Value: v}
}

type Int = Null[int]

func NewInt(v int) Int {
	return Null[int]{Set: true, Value: v}
}

type Float32 = Null[float32]

func NewFloat32(v float32) Float32 {
	return Null[float32]{Set: true, Value: v}
}

type Float64 = Null[float64]

func NewFloat64(v float64) Float64 {
	return Null[float64]{Set: true, Value: v}
}

type Bool = Null[bool]

func NewBool(v bool) Bool {
	return Null[bool]{Set: true, Value: v}
}

type String = Null[string]

func NewString(v string) String {
	return Null[string]{Set: true, Value: v}
}

type Bytes = Null[[]byte]

func NewBytes(v []byte) Bytes {
	return Null[[]byte]{Set: true, Value: v}
}
