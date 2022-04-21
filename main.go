// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
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

package main

import (
	"github.com/go-resty/resty/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/goccy/go-json"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/bangumi/server/auth"
	"github.com/bangumi/server/cache"
	"github.com/bangumi/server/character"
	"github.com/bangumi/server/config"
	"github.com/bangumi/server/episode"
	"github.com/bangumi/server/index"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/person"
	"github.com/bangumi/server/revision"
	"github.com/bangumi/server/subject"
	"github.com/bangumi/server/user"
	"github.com/bangumi/server/web"
	"github.com/bangumi/server/web/captcha/hcaptcha"
	"github.com/bangumi/server/web/handler"
	"github.com/bangumi/server/web/session"
)

func main() {
	if err := mysql.SetLogger(logger.Std()); err != nil {
		logger.Panic("can't replace mysql driver's errLog", zap.Error(err))
	}

	if err := start(); err != nil {
		logger.Fatal("failed to start app", zap.Error(err))
	}
}

func start() error {
	app := fx.New(
		logger.FxLogger(),
		// driver and connector
		fx.Provide(
			driver.NewRedisClient, // redis
			dal.NewConnectionPool,
			dal.NewDB,
			func() *resty.Client {
				httpClient := resty.New().SetJSONEscapeHTML(false)
				httpClient.JSONUnmarshal = json.Unmarshal
				httpClient.JSONMarshal = json.MarshalNoEscape
				return httpClient
			},
		),

		fx.Provide(
			config.NewAppConfig, logger.Copy, metrics.NewScope,

			query.Use, cache.NewRedisCache,

			character.NewMysqlRepo, subject.NewMysqlRepo, user.NewUserRepo, person.NewMysqlRepo,
			index.NewMysqlRepo, auth.NewMysqlRepo, episode.NewMysqlRepo, revision.NewMysqlRepo,

			auth.NewService, character.NewService, subject.NewService, person.NewService,
		),

		fx.Provide(
			hcaptcha.New, session.New, handler.New, web.New,
		),

		fx.Invoke(
			web.ResistRouter, web.Listen,
		),
	)

	app.Run()

	return errgo.Wrap(app.Err(), "failed to start app")
}
