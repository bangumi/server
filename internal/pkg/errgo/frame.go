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
	"path"
	"runtime"
	"strconv"
)

// frame represents a program counter inside a stack frame.
// For historical reasons if frame is interpreted as a uintptr
// its value represents the program counter + 1.
type frame uintptr

// pc returns the program counter for this frame;
// multiple frames may have the same PC value.
func (f frame) pc() uintptr { return uintptr(f) - 1 }

// src returns the full path to the file and file line
// that contains the function for this frame's pc.
func (f frame) src() (string, string, int) {
	fn := runtime.FuncForPC(f.pc())
	if fn == nil {
		return unknownText, unknownText, 0
	}

	file, no := fn.FileLine(f.pc())
	return fn.Name(), file, no
}

// Format formats the frame according to the fmt.Formatter interface.
//
//    %s    source file
//    %v    equivalent to %s:%d
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//    %+s   function name and path of source file relative to the compile time
//          GOPATH separated by \n\t (<funcname>\n\t<path>)
//    %+v   equivalent to %+s:%d
func (f frame) Format(s fmt.State, verb rune) {
	name, file, line := f.src()
	switch verb {
	case 's':
		switch {
		case s.Flag('+'):
			_, _ = io.WriteString(s, name)
			_, _ = io.WriteString(s, "\n\t")
			_, _ = io.WriteString(s, file)
		default:
			_, _ = io.WriteString(s, path.Base(file))
		}
	case 'v':
		f.Format(s, 's')
		_, _ = io.WriteString(s, ":")
		_, _ = io.WriteString(s, strconv.Itoa(line))
	default:
		_, _ = io.WriteString(s, strconv.FormatUint(uint64(f), 10))
	}
}

// MarshalText formats a stacktrace frame as a text string. The output is the
// same as that of fmt.Sprintf("%+v", f), but without newlines or tabs.
func (f frame) MarshalText() ([]byte, error) {
	name, file, line := f.src()
	if name == unknownText {
		return []byte(name), nil
	}
	return []byte(fmt.Sprintf("%s %s:%d", name, file, line)), nil
}
