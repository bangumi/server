// Copyright (c) 2021-2022 Trim21 <trim21.me@gmail.com>
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

//go:build dev

package recovery

import (
	_ "embed"
	"fmt"
	"runtime"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/bytebufferpool"
)

//go:embed debug.html
var _debugHTML string

// New creates a new middleware handler with debug info.
func New() fiber.Handler {
	// Return new handler
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				c.Status(fiber.StatusInternalServerError).
					Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
				_, err = fmt.Fprintf(c, _debugHTML, takeStacktrace(2))
			}
		}()

		return c.Next()
	}
}

type programCounters struct {
	pcs []uintptr
}

func newProgramCounters(size int) *programCounters {
	return &programCounters{make([]uintptr, size)}
}

const defaultProgramCounter = 64

// this is mainly copied from uber/zap's trace reader.
func takeStacktrace(skip int) string {
	programCounters := newProgramCounters(defaultProgramCounter)
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)

	var numFrames int
	for {
		numFrames = runtime.Callers(skip+2, programCounters.pcs)
		if numFrames < len(programCounters.pcs) {
			break
		}

		programCounters = newProgramCounters(len(programCounters.pcs) * 2)
	}

	i := 0
	frames := runtime.CallersFrames(programCounters.pcs[:numFrames])

	// Note: On the last iteration, frames.Next() returns false, with a valid
	// frame, but we ignore this frame. The last frame is a a runtime frame which
	// adds noise, since it's only either runtime.main or runtime.goexit.
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		if i != 0 {
			buf.WriteByte('\n')
		}
		i++
		buf.WriteString(frame.Function)
		buf.WriteByte('\n')
		buf.WriteByte('\t')
		buf.WriteString(frame.File)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(frame.Line))
	}

	return buf.String()
}
