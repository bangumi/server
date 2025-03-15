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
	"encoding/json"

	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
	"github.com/segmentio/kafka-go"
	"github.com/spf13/cobra"
	"github.com/trim21/errgo"
	"go.uber.org/fx"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/dal"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/collections/infra"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/index"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/dam"
	"github.com/bangumi/server/internal/pkg/driver"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/revision"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/tag"
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
			driver.NewRueidisClient, // redis
			driver.NewMysqlSqlDB,    // mysql
			func() *resty.Client {
				httpClient := resty.New().SetJSONEscapeHTML(false)
				httpClient.JSONUnmarshal = json.Unmarshal
				httpClient.JSONMarshal = json.Marshal
				return httpClient
			},
		),

		fx.Invoke(dal.SetupMetrics),

		dal.Module,

		fx.Provide(
			logger.Copy, cache.NewRedisCache,

			user.NewMysqlRepo,
			index.NewMysqlRepo, auth.NewMysqlRepo, episode.NewMysqlRepo, revision.NewMysqlRepo, infra.NewMysqlRepo,
			timeline.NewSrv,

			dam.New, subject.NewMysqlRepo, subject.NewCachedRepo,
			character.NewMysqlRepo, person.NewMysqlRepo,

			tag.NewCachedRepo, tag.NewMysqlRepo,

			auth.NewService, person.NewService, search.New,
		),

		fx.Provide(
			func(cfg config.AppConfig) *kafka.Writer {
				logger.Info("new kafka stream broker")
				return kafka.NewWriter(kafka.WriterConfig{
					Brokers: []string{cfg.Canal.KafkaBroker},
				})
			},
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
