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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/trim21/errgo"
	"gorm.io/gorm"
)

func SetupMetrics(db *gorm.DB, conn *sql.DB) error {
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

	prometheus.MustRegister(DatabaseQuery)
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Subsystem: "chii",
			Name:      "db_open_connections",
			Help:      "opened connections",
		},
		func() float64 {
			s := conn.Stats()
			return float64(s.OpenConnections)
		},
	))

	return nil
}
