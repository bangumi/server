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

package dal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/internal/pkg/logger"
)

// production gorm logger log do sql time monitoring and error logging to zap logger.
func newProdLog(cfg config.AppConfig) gormLogger.Interface {
	return &metricsLog{
		slowThreshold: cfg.Mysql.SlowSQLDuration,
		log:           logger.Named("gorm").WithOptions(zap.AddStacktrace(zap.DPanicLevel)),
		h:             metrics.SQLHistogram,
	}
}

type metricsLog struct {
	slowThreshold time.Duration
	h             prometheus.Histogram
	log           *zap.Logger
}

func (l *metricsLog) Info(_ context.Context, s string, i ...any) {
	l.log.Info(fmt.Sprintln(s, i))
}

func (l *metricsLog) Warn(_ context.Context, s string, i ...any) {
	l.log.Warn(fmt.Sprintln(s, i))
}

func (l *metricsLog) Error(_ context.Context, s string, i ...any) {
	l.log.Error(fmt.Sprintln(s, i))
}

func (l *metricsLog) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	log := l.log
	switch level {
	case gormLogger.Silent:
		log = zap.NewNop()
	case gormLogger.Info:
		log.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
	case gormLogger.Warn:
		log.WithOptions(zap.IncreaseLevel(zap.WarnLevel))
	case gormLogger.Error:
		log.WithOptions(zap.IncreaseLevel(zap.ErrorLevel))
	}

	return &metricsLog{
		slowThreshold: l.slowThreshold,
		h:             l.h,
		log:           log,
	}
}

func (l *metricsLog) Trace(_ context.Context, begin time.Time, fc func() (sql string, rows int64), err error) {
	elapsed := time.Since(begin)
	l.h.Observe(elapsed.Seconds())

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		l.log.Error("gorm error", zap.String("sql", sql), zap.Error(err),
			zap.Duration("duration", elapsed), zap.Int64("rows", rows))
	case l.slowThreshold > 0 && elapsed >= l.slowThreshold:
		sql, rows := fc()
		l.log.Warn("slow SQL", zap.String("sql", sql), zap.Duration("duration", elapsed), zap.Int64("rows", rows))
	}
}
