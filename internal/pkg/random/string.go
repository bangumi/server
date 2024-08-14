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

package random

import (
	"bufio"
	"crypto/rand"
	"io"

	"github.com/bangumi/server/internal/pkg/generic/pool"
)

var p = pool.New(func() *bufio.Reader {
	return bufio.NewReader(rand.Reader)
})

// we may never need to change these values.
const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const base62CharsLength = byte(len(base62Chars))
const base62MaxByte = byte(255 - (256 % len(base62Chars)))

// Base62String generate a cryptographically secure base62 string in given length.
// Will panic if it can't read from 'crypto/rand'.
func Base62String(length int) string {
	reader := p.Get()
	defer p.Put(reader)

	b := make([]byte, length)
	// storage for random bytes.
	r := make([]byte, length+(length/4)) //nolint:gomnd
	i := 0

	for {
		n, err := io.ReadFull(reader, r)
		if err != nil {
			panic("unexpected error happened when reading from bufio.NewReader(crypto/rand.Reader)")
		}
		if n != len(r) {
			panic("partial reads occurred when reading from bufio.NewReader(crypto/rand.Reader)")
		}
		for _, rb := range r {
			if rb > base62MaxByte {
				// Skip this number to avoid modulo bias.
				continue
			}
			b[i] = base62Chars[rb%base62CharsLength]
			i++
			if i == length {
				return string(b)
			}
		}
	}
}
