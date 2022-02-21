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

package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/uber-go/tally/v4"
	promreporter "github.com/uber-go/tally/v4/prometheus"
	"go.uber.org/fx"

	"github.com/bangumi/server/internal/errgo"
)

// DefaultDurationBucket currently return a bucket
// [0, .01, .025, .05, .075, .1, .2, .3, .4, .5, .6, .8, 1, 2, 5, +Inf]
// leave it for future adjustment.
func DefaultDurationBucket() tally.Buckets {
	return tally.DefaultBuckets
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
