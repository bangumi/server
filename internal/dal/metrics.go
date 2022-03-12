// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
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

package dal

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/uber-go/tally/v4"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/metrics"
)

func newMetricsLog(log gormLogger.Interface, scope tally.Scope) gormLogger.Interface {
	return metricsLog{
		Interface: log,
		h:         scope.Histogram("sql_time", metrics.SQLTimeBucket()),
	}
}

type metricsLog struct {
	gormLogger.Interface
	h tally.Histogram
}

func (l metricsLog) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64), err error,
) {
	fc()
	l.h.RecordDuration(time.Since(begin))
}

func setupMetrics(db *gorm.DB, scope tally.Scope, register prometheus.Registerer) error {
	db.Logger = newMetricsLog(db.Logger, scope)

	var DatabaseQuery = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chii_db_execute_total",
			Help: "Number of executing sql.",
		},
		[]string{"table"},
	)

	// uber/tally doesn't like dynamic tag value.
	err := db.Callback().Query().Before("gorm:select").Register("metrics:select", func(db *gorm.DB) {
		DatabaseQuery.WithLabelValues(db.Statement.Table).Inc()
	})

	if err != nil {
		return errgo.Wrap(err, "gorm callback")
	}

	register.MustRegister(DatabaseQuery)

	return nil
}
