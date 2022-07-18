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

package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/uber-go/tally/v4"
	promreporter "github.com/uber-go/tally/v4/prometheus"
	"go.uber.org/fx"

	"github.com/bangumi/server/internal/pkg/errgo"
)

func SQLTimeBucket() tally.Buckets {
	return tally.DurationBuckets{
		time.Millisecond,
		time.Millisecond * 2,
		time.Millisecond * 3,
		time.Millisecond * 4,
		time.Millisecond * 5,
		time.Millisecond * 10,
		time.Millisecond * 20,
		time.Millisecond * 30,
		time.Millisecond * 40,
		time.Millisecond * 50,
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 300,
		time.Millisecond * 500,
		time.Second,
	}
}

func ResponseTimeBucket() tally.Buckets {
	return tally.DurationBuckets{
		time.Millisecond,
		time.Millisecond * 5,
		time.Millisecond * 10,
		time.Millisecond * 20,
		time.Millisecond * 50,
		time.Millisecond * 100,
		time.Millisecond * 200,
		time.Millisecond * 300,
		time.Millisecond * 500,
		time.Second,
	}
}

func NewScope(lc fx.Lifecycle) (tally.Scope, promreporter.Reporter, prometheus.Registerer) {
	r := promreporter.NewReporter(promreporter.Options{})
	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Prefix:         "chii",
		Tags:           map[string]string{},
		CachedReporter: r,
		Separator:      promreporter.DefaultSeparator,
	}, time.Second)

	lc.Append(fx.Hook{OnStop: func(ctx context.Context) error {
		return errgo.Wrap(closer.Close(), "close tally")
	}})

	return scope, r, prometheus.DefaultRegisterer
}
