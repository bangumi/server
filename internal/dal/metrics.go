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
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/uber-go/tally/v4"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/pkg/errgo"
)

func setupMetrics(db *gorm.DB, conn *sql.DB, scope tally.Scope, register prometheus.Registerer) error {
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

	dbConnCount := scope.Gauge("db_open_connections_total")
	go func() {
		for {
			s := conn.Stats()
			dbConnCount.Update(float64(s.OpenConnections))
			time.Sleep(time.Second * 15)
		}
	}()

	return nil
}
