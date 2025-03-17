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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/trim21/errgo"
	"go.uber.org/fx"
	"golang.org/x/sync/errgroup"

	"github.com/bangumi/server/config"
	"github.com/bangumi/server/dal"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/person"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/driver"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/sys"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/tag"
	"github.com/bangumi/server/web/session"
)

const groupID = "go-canal"

var errNoTopic = fmt.Errorf("missing search events topic")

// nolint: funlen
func Main() error {
	cfg, err := config.AppConfigReader(config.AppTypeCanal)()
	if err != nil {
		return errgo.Trace(err)
	}

	if len(cfg.Canal.Topics) == 0 {
		return errNoTopic
	}

	var h *eventHandler
	di := fx.New(
		fx.NopLogger,
		dal.Module,

		fx.Provide(func() config.AppConfig { return cfg }),

		// driver and connector
		fx.Provide(
			driver.NewMysqlSqlDB,
			driver.NewRueidisClient, logger.Copy, cache.NewRedisCache,
			subject.NewMysqlRepo, character.NewMysqlRepo, person.NewMysqlRepo,
			search.New, session.NewMysqlRepo, session.New,
			driver.NewS3,
			tag.NewCachedRepo,
			tag.NewMysqlRepo,

			newEventHandler,
		),

		fx.Provide(newKafkaStream),

		fx.Populate(&h),
	)

	if err := di.Err(); err != nil {
		return errgo.Wrap(err, "fx")
	}

	// metrics http reporter
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Addr: h.config.ListenAddr(), Handler: mux, ReadHeaderTimeout: time.Second}

	var wg errgroup.Group

	defer srv.Shutdown(context.Background()) //nolint:errcheck
	wg.Go(func() error {
		return errgo.Wrap(srv.ListenAndServe(), "http")
	})

	defer h.Close()
	wg.Go(func() error {
		return errgo.Wrap(h.start(), "start")
	})

	wg.Go(func() error {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		select {
		case <-sys.HandleSignal():
			logger.Info("receive signal, shutdown")
			return fmt.Errorf("receive signal")
		}
	})

	return wg.Wait()
}
