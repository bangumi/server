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

package wiki

import (
	"strings"
)

const spaceStr = " \t"

func trimSpace(s string) string {
	return strings.Trim(s, spaceStr)
}

func trimLeftSpace(s string) string {
	return strings.TrimLeft(s, spaceStr)
}

func trimRightSpace(s string) string {
	return strings.TrimRight(s, spaceStr)
}

func unifyEOL(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")

	return s
}

func processInput(s string) (string, int) {
	offset := 2
	s = unifyEOL(s)

	for _, c := range s {
		switch c {
		case '\n':
			offset++
		case ' ', '\t':
			continue
		default:
			return strings.TrimSpace(s), offset
		}
	}

	return strings.TrimSpace(s), offset
}
