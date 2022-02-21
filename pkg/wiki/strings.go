// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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

func trim(s string, cut byte) string {
	i, j := 0, len(s)-1
	for ; i < j; i++ {
		if s[i] != cut {
			break
		}
	}
	for ; i <= j; j-- {
		if s[j] != cut {
			break
		}
	}

	return s[i : j+1]
}

func trimRight(s string, cut byte) string {
	j := len(s) - 1
	for ; 0 <= j; j-- {
		if s[j] != cut {
			break
		}
	}

	return s[:j+1]
}

func trimLeft(s string, cut byte) string {
	i, j := 0, len(s)-1
	for ; i < j; i++ {
		if s[i] != cut {
			break
		}
	}

	return s[i:]
}

// zero alloc version `strings.Trim(s, " ")`
// make wiki.Parse 2x faster.
func trimSpace(s string) string {
	return trim(s, ' ')
}

func trimLeftSpace(s string) string {
	return trimLeft(s, ' ')
}

func trimRightSpace(s string) string {
	return trimRight(s, ' ')
}

func unifyCharacter(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")

	return s
}

func processInput(s string) (string, int) {
	offset := 2
	s = unifyCharacter(s)

	for _, c := range s {
		switch c {
		case '\n':
			offset++
		case ' ':
			continue
		default:
			return strings.TrimSpace(s), offset
		}
	}

	return strings.TrimSpace(s), offset
}
