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

package canal

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"go.uber.org/fx"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/sys"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/web/session"
)

func Main() error {
	var h *eventHandler
	di := fx.New(
		fx.NopLogger,
		dal.Module,

		// driver and connector
		fx.Provide(
			config.AppConfigReader(config.AppTypeCanal),
			driver.NewMysqlConnectionPool,
			driver.NewRedisClient, logger.Copy, cache.NewRedisCache,
			subject.NewMysqlRepo, search.New, session.NewMysqlRepo, session.New,

			newKafkaReader, newEventHandler,
		),

		fx.Populate(&h),
	)

	if err := di.Err(); err != nil {
		return errgo.Wrap(err, "fx")
	}

	var errChan = make(chan error, 1)

	// metrics http reporter
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Addr: h.config.ListenAddr(), Handler: mux, ReadHeaderTimeout: time.Second}
	go func() { errChan <- errgo.Wrap(srv.ListenAndServe(), "http") }()
	defer srv.Shutdown(context.Background()) //nolint:errcheck

	go func() { errChan <- errgo.Wrap(h.start(), "kafka") }()
	defer h.Close()

	select {
	case err := <-errChan:
		return err
	case <-sys.HandleSignal():
		logger.Info("receive signal, shutdown")
		return nil
	}
}

const groupID = "my-group"

func newKafkaReader(c config.AppConfig) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{c.KafkaBroker},
		GroupID:     groupID,
		GroupTopics: c.KafkaCanalTopics,
	})
}
