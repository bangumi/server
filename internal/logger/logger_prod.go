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

//go:build !dev && !test

package logger

import (
	"os"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// production log config.
func getLogger(level zapcore.Level) *zap.Logger {
	prod := zapcore.EncoderConfig{
		TimeKey:        timeKey,
		NameKey:        nameKey,
		MessageKey:     messageKey,
		CallerKey:      callerKey,
		LevelKey:       levelKey,
		StacktraceKey:  traceKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   getCallerEncoder(),
	}

	return zap.New(
		zapcore.NewCore(zapcore.NewJSONEncoder(prod), zapcore.AddSync(os.Stdout), level),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
}

func FxLogger() fx.Option {
	return fx.WithLogger(func() fxevent.Logger { return &fxevent.ZapLogger{Logger: Named("fx")} })
}
