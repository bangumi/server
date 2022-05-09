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

package main

import (
	"context"
	"io"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/bangumi/server/cache"
	"github.com/bangumi/server/config"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/metrics"
)

func main() {
	if err := mysql.SetLogger(logger.Std()); err != nil {
		logger.Panic("can't replace mysql driver's errLog", zap.Error(err))
	}

	start()
}

func start() {
	var q *query.Query
	err := fx.New(
		logger.FxLogger(),
		fx.Provide(
			driver.NewRedisClient, dal.NewConnectionPool, dal.NewDB,

			config.NewAppConfig, logger.Copy, metrics.NewScope,

			query.Use, cache.NewRedisCache,
		),

		fx.Populate(&q),
	).Err()

	if err != nil {
		logger.Err(err, "failed to fill deps")
	}
}

func exportSubjects(q *query.Query, w io.Writer) {
	q.WithContext(context.Background()).Subject.Where(q.Subject.ID.Gt(1))
}
