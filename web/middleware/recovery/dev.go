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

package recovery

import (
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/pkg/logger"
)

//go:embed debug.html
var _debugHTML string

// New creates a new middleware handler with debug info.
func dev() fiber.Handler {
	log := logger.Copy().WithOptions(zap.AddCallerSkip(2))
	// Return new handler
	return func(c *fiber.Ctx) (err error) { //nolint:nonamedreturns
		defer func() {
			if r := recover(); r != nil {
				c.Status(http.StatusInternalServerError).
					Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
				_, err = fmt.Fprintf(c, _debugHTML, r, takeStacktrace(2))
				log.Error("panic: " + fmt.Sprintln(r))
			}
		}()

		return c.Next()
	}
}

// this is mainly copied from uber/zap's trace reader.
func takeStacktrace(skip int) string {
	const defaultProgramCounter = 64

	programCounters := make([]uintptr, defaultProgramCounter)
	buf := bytes.NewBuffer(nil)

	var numFrames int
	for {
		numFrames = runtime.Callers(skip+2, programCounters)
		if numFrames < len(programCounters) {
			break
		}

		programCounters = make([]uintptr, len(programCounters)*2)
	}

	frames := runtime.CallersFrames(programCounters[:numFrames])

	// Note: On the last iteration, frames.Next() returns false, with a valid
	// frame, but we ignore this frame. The last frame is a a runtime frame which
	// adds noise, since it's only either runtime.main or runtime.goexit.
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		buf.WriteByte('\n')
		buf.WriteString(frame.Function)
		buf.WriteByte('\n')
		buf.WriteByte('\t')
		buf.WriteString(frame.File)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(frame.Line))
	}

	return strings.TrimSpace(buf.String())
}
