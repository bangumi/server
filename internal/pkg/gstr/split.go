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
	"strings"
)

// Split a string into slice and remove empty elements.
func Split(s string, c string) []string {
	split := strings.Split(s, c)

	result := make([]string, 0, len(split))
	for _, s2 := range split {
		if s2 != "" {
			result = append(result, s2)
		}
	}

	return result
}
