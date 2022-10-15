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

	"github.com/valyala/bytebufferpool"

	"github.com/bangumi/server/internal/pkg/generic/slice"
)

type unwrap interface {
	Unwrap() error
}

var _ unwrap = (*wrapError)(nil)
var _ unwrap = (*msgError)(nil)
var _ unwrap = (*withStackError)(nil)

type wrapError struct {
	err error
	msg string
}

func (e *wrapError) Error() string {
	return e.msg + ": " + e.err.Error()
}

func (e *wrapError) Unwrap() error {
	return e.err
}

// func (e *wrapError) Format(s fmt.State, v rune) {
// 	switch v {
// 	case 'v':
// 		switch {
// 		case s.Flag('+'):
// 			fmt.Fprintf(s, "%s: %s", e.msg, e.err)
// 			return
// 		case s.Flag('#'):
// 			fmt.Fprintf(s, "&wrapError{msg: %q, err: %#v}", e.msg, e.err)
// 			return
// 		}
// 		fallthrough
// 	case 's':
// 		_, _ = io.WriteString(s, e.Error())
// 	case 'q':
// 		_, _ = io.WriteString(s, strconv.Quote(e.Error()))
// 	}
// }

type msgError struct {
	err error
	msg string
}

func (e *msgError) Error() string {
	return e.msg
}

func (e *msgError) Unwrap() error {
	return e.err
}

type withStackError struct {
	Err   error
	Stack stack
}

func (w *withStackError) Error() string {
	return w.Err.Error()
}

// Unwrap provides compatibility for Go 1.13 error chains.
func (w *withStackError) Unwrap() error { return w.Err }

// Format implement fmt.Formatter, add trace to zap.Error(err).
func (w *withStackError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'T':
		_, _ = io.WriteString(s, "*withStackError")
	case 'v':
		// _, _ = io.WriteString(s, w.Error())
		if s.Flag('#') {
			fmt.Fprintf(s, "&errgo.withStackError{Err: %#v, Stack: ...}", w.Err)
			return
		}

		if s.Flag('+') {
			_, _ = io.WriteString(s, w.Err.Error())
			_, _ = io.WriteString(s, "\nerror stack:")
			w.Stack.Format(s, verb)
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, w.Error())
	case 'q':
		_, _ = io.WriteString(s, strconv.Quote(w.Error()))
	}
}

// MarshalJSON marshal error with stack to
//
//	{
//		"error": "context: real error",
//		"stack": [
//			"main.main  ...main.go:54",
//			"..."
//		]
//	}
func (w *withStackError) MarshalJSON() ([]byte, error) {
	if w == nil {
		return []byte("null"), nil
	}

	b := bytebufferpool.Get()
	defer bytebufferpool.Put(b)

	b.WriteString(`{"error":`)
	b.B = strconv.AppendQuote(b.B, w.Error())
	b.WriteString(",")

	b.WriteString(`"stack":[`)

	frames := runtime.CallersFrames(w.Stack)
	for {
		frame, more := frames.Next()
		b.WriteString(`"`)
		b.WriteString(frame.Function)
		b.WriteString("  ")
		b.WriteString(frame.File)
		b.WriteString(":")
		b.B = strconv.AppendInt(b.B, int64(frame.Line), 10)
		b.WriteString(`"`)
		if !more {
			break
		}
		b.WriteString(",")
	}

	b.WriteString("]}")

	return slice.Clone(b.B), nil
}
