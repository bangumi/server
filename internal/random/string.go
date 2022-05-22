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
	"crypto/rand"

	"github.com/gofiber/fiber/v2/utils"
)

const availableCharBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const availableCharLength = uint8(len(availableCharBytes))
const maxByte = 247 // 255 - (256 % availableCharLength)

// Base62String generate a cryptographically secure base62 string in given length.
// Will panic if it can't read from 'crypto/rand'.
func Base62String(length int) string {
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) //nolint:gomnd    // storage for random bytes.
	i := 0
	for {
		if _, err := rand.Read(r); err != nil {
			panic("unexpected error happened when reading from 'crypto/rand'")
		}
		for _, rb := range r {
			if rb > maxByte {
				// Skip this number to avoid modulo bias.
				continue
			}
			b[i] = availableCharBytes[rb%availableCharLength]
			i++
			if i == length {
				return utils.UnsafeString(b)
			}
		}
	}
}
