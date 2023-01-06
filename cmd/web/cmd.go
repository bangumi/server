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

package web

import (
	"github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/dal"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/index"
	"github.com/bangumi/server/internal/notification"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/dam"
	"github.com/bangumi/server/internal/pkg/driver"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pm"
	"github.com/bangumi/server/internal/revision"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/timeline"
	"github.com/bangumi/server/internal/user"
	"github.com/bangumi/server/web"
)

var Command = &cobra.Command{
	Use:   "web",
	Short: "start web server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return start()
	},
}

func start() error {
	var e *echo.Echo
	var cfg config.AppConfig

	err := fx.New(
		fx.NopLogger,

		// driver and connector
		fx.Provide(
			config.AppConfigReader(config.AppTypeHTTP),
			driver.NewRedisClient,         // redis
			driver.NewMysqlConnectionPool, // mysql
			func() *resty.Client {
				httpClient := resty.New().SetJSONEscapeHTML(false)
				httpClient.JSONUnmarshal = sonic.Unmarshal
				httpClient.JSONMarshal = sonic.Marshal
				return httpClient
			},
		),

		dal.Module,

		fx.Provide(
			logger.Copy, cache.NewRedisCache,

			character.NewMysqlRepo,

			user.NewMysqlRepo,
			index.NewMysqlRepo, auth.NewMysqlRepo, episode.NewMysqlRepo, revision.NewMysqlRepo, collection.NewMysqlRepo,
			timeline.NewMysqlRepo, pm.NewMysqlRepo, notification.NewMysqlRepo,

			dam.New, subject.NewMysqlRepo, subject.NewCachedRepo, person.NewMysqlRepo,

			auth.NewService, person.NewService, search.New,
		),

		ctrl.Module,
		web.Module,

		fx.Populate(&e, &cfg),
	).Err()

	if err != nil {
		return err //nolint:wrapcheck
	}

	return errgo.Wrap(web.Start(cfg, e), "failed to start app")
}
