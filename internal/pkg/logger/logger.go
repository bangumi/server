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

// Package logger config a zap logger, functions have same signature with `zap.logger`.
// Can be configured by env `LOG_LEVEL`.
package logger

import (
	"fmt"
	stdLog "log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log = setup()

const (
	timeKey    = "time"
	nameKey    = "logger"
	messageKey = "msg"
	callerKey  = "caller"
	levelKey   = "level"
	traceKey   = "trace"
)

func setup() *zap.Logger {
	var level zapcore.Level
	var err error

	if level, err = zapcore.ParseLevel(os.Getenv("LOG_LEVEL")); err != nil {
		level = zapcore.InfoLevel
	}

	l := getLogger(level)
	zap.RedirectStdLog(l)

	zap.ReplaceGlobals(l)

	return l
}

// Std return a stdlib logger with zap logger underlying.
func Std() *stdLog.Logger {
	return zap.NewStdLog(log.WithOptions(zap.AddCallerSkip(-1)))
}

func Copy() *zap.Logger {
	return log.WithOptions(zap.AddCallerSkip(-1))
}

// Named create a named logger.
func Named(name string) *zap.Logger {
	return log.Named(name).WithOptions(zap.AddCallerSkip(-1))
}

// Debug level logging.
func Debug(msg string, fields ...zapcore.Field) {
	log.Debug(msg, fields...)
}

// Info level logging.
func Info(msg string, fields ...zapcore.Field) {
	log.Info(msg, fields...)
}

// Infoln log as info level with fmt.Sprintln.
func Infoln(args ...any) {
	// remove \n from msg.
	msg := fmt.Sprintln(args...)
	log.Info(msg[:len(msg)-1])
}

// Warn level logging.
func Warn(msg string, fields ...zapcore.Field) {
	log.Warn(msg, fields...)
}

// Err log a error message with error.
func Err(err error, msg string, fields ...zapcore.Field) {
	log.With(zap.Error(err)).Error(msg, fields...)
}

// Error level logging.
func Error(msg string, fields ...zapcore.Field) {
	log.Error(msg, fields...)
}

// DPanic will panic in development and log error message at production env.
func DPanic(msg string, fields ...zapcore.Field) {
	log.DPanic(msg, fields...)
}

// Panic will log at panic level then panic.
func Panic(msg string, fields ...zapcore.Field) {
	log.Panic(msg, fields...)
}

// Fatal level message and call `os.Exit(1)`.
func Fatal(msg string, fields ...zapcore.Field) {
	log.Fatal(msg, fields...)
}
