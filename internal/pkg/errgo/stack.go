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
	"io"
	"runtime"
	"strconv"
)

// stack represents a stack of program counters.
type stack []uintptr

func (s stack) Format(st fmt.State, verb rune) {
	if verb == 'v' && st.Flag('+') {
		frames := runtime.CallersFrames(s)

		for {
			frame, more := frames.Next()
			_, _ = io.WriteString(st, "\n")
			_, _ = io.WriteString(st, frame.Function)
			_, _ = io.WriteString(st, "\n\t")
			_, _ = io.WriteString(st, frame.File)
			_, _ = io.WriteString(st, ":")
			_, _ = io.WriteString(st, strconv.Itoa(frame.Line))
			if !more {
				break
			}
		}
	}
}

func callers() stack {
	const depth = 16
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	return pcs[:n]
}
