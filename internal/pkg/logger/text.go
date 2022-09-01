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

//go:build test || dev

package logger

import (
	"strings"

	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func textLogger(level zapcore.Level) *zap.Logger {
	consoleEncoding := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        timeKey,
		NameKey:        nameKey,
		MessageKey:     messageKey,
		CallerKey:      callerKey,
		LevelKey:       levelKey,
		StacktraceKey:  traceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller: func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {
			const prefix = "github.com/bangumi/server"
			p := caller.String()
			if strings.HasPrefix(p, prefix) {
				encoder.AppendString("." + strings.TrimPrefix(p, prefix))
				return
			}
			encoder.AppendString(p)
		},
	})

	return zap.New(
		zapcore.NewCore(
			consoleEncoding, zapcore.AddSync(colorable.NewColorableStdout()), level,
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.WarnLevel),
		zap.Development(),
	)
}
