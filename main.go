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
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/collection"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/ctrl"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/dam"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/group"
	"github.com/bangumi/server/internal/index"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/internal/oauth"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/revision"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/topic"
	"github.com/bangumi/server/internal/user"
	"github.com/bangumi/server/internal/web"
)

func main() {
	if err := start(); err != nil {
		logger.Fatal("failed to start app", zap.Error(err))
	}
}

func start() error {
	var f *fiber.App
	var cfg config.AppConfig

	err := fx.New(
		logger.FxLogger(),
		config.Module,

		// driver and connector
		fx.Provide(
			driver.NewRedisClient,         // redis
			driver.NewMysqlConnectionPool, // mysql
			func() *resty.Client {
				httpClient := resty.New().SetJSONEscapeHTML(false)
				httpClient.JSONUnmarshal = json.Unmarshal
				httpClient.JSONMarshal = json.Marshal
				return httpClient
			},
		),

		dal.Module,

		fx.Provide(
			logger.Copy, metrics.NewScope, cache.NewRedisCache,

			oauth.NewMysqlRepo,

			character.NewMysqlRepo, subject.NewMysqlRepo, user.NewUserRepo, person.NewMysqlRepo,
			index.NewMysqlRepo, auth.NewMysqlRepo, episode.NewMysqlRepo, revision.NewMysqlRepo, collection.NewMysqlRepo,
			topic.NewMysqlRepo,

			dam.New,

			auth.NewService, person.NewService, group.NewMysqlRepo,
		),

		ctrl.Module,
		web.Module,

		fx.Populate(&f, &cfg),
	).Err()

	if err != nil {
		return errgo.Wrap(err, "fx")
	}

	return errgo.Wrap(web.Start(cfg, f), "failed to start app")
}
