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

package logger

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// production log config.
func jsonLogger(level zapcore.Level) *zap.Logger {
	prod := zapcore.EncoderConfig{
		TimeKey:       timeKey,
		NameKey:       nameKey,
		MessageKey:    messageKey,
		CallerKey:     callerKey,
		LevelKey:      levelKey,
		StacktraceKey: traceKey,
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime:    zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: func(duration time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString((duration / time.Millisecond * time.Millisecond).String()) //nolint:durationcheck
		},
		EncodeCaller: zapcore.FullCallerEncoder,

		NewReflectedEncoder: defaultReflectedEncoder,
	}

	return zap.New(
		zapcore.NewCore(zapcore.NewJSONEncoder(prod), zapcore.AddSync(os.Stdout), level),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}

func defaultReflectedEncoder(w io.Writer) zapcore.ReflectedEncoder {
	enc := json.NewEncoder(w)
	// For consistency with our custom BodyJSON encoder.
	enc.SetEscapeHTML(false)
	return enc
}
