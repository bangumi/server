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

package errgo

import (
	"fmt"
	"runtime"
)

const unknownText = "(unknown)"

// stack represents a stack of program counters.
type stack []uintptr

func (s stack) Format(st fmt.State, verb rune) {
	if verb == 'v' && st.Flag('+') {
		for _, pc := range s {
			st.Write([]byte("\n"))
			frame(pc).Format(st, 'v')
			// fmt.Fprintf(st, "\n%+v", frame(pc))
		}
	}
}

func toFrames(s []uintptr) []frame {
	f := make([]frame, len(s))
	for i, u := range s {
		f[i] = frame(u)
	}
	return f
}

func callers() stack {
	const depth = 16
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[0:n]
	return st
}

type encodeToJSONError struct {
	Message string  `json:"msg"`
	Stace   []frame `json:"stace"`
}
