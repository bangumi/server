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

//nolint:mnd,gochecknoinits,gochecknoglobals
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.MustRegister(RequestCount)
	prometheus.MustRegister(RequestHistogram)
	prometheus.MustRegister(SQLHistogram)
}

var RequestCount = prometheus.NewCounter(prometheus.CounterOpts{
	Subsystem: "chii",
	Name:      "request_count_total",
	Help:      "",
})

var RequestHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
	Subsystem: "chii",
	Name:      "response_time_seconds",
	Help:      "web response time",
	Buckets: []float64{
		0.001,
		0.005,
		0.010,
		0.020,
		0.050,
		0.100,
		0.200,
		0.300,
		0.500,
		1.000,
	},
})

var SQLHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
	Subsystem: "chii",
	Name:      "sql_time_seconds",
	Help:      "sql execution time",
	Buckets: []float64{
		0.001,
		0.002,
		0.003,
		0.004,
		0.005,
		0.010,
		0.020,
		0.030,
		0.040,
		0.050,
		0.100,
		0.200,
		0.300,
		0.500,
		1.000,
	},
})
