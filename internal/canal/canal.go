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
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	promreporter "github.com/uber-go/tally/v4/prometheus"
	"go.uber.org/fx"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/web/session"
)

func Main() error {
	var h *eventHandler
	var reporter promreporter.Reporter

	di := fx.New(
		logger.FxLogger(),
		config.Module, dal.Module,

		// driver and connector
		fx.Provide(
			driver.NewMysqlConnectionPool, metrics.NewScope,
			driver.NewRedisClient, logger.Copy, cache.NewRedisCache,
			subject.NewMysqlRepo, search.New, session.NewMysqlRepo, session.New,

			newKafkaReader, newEventHandler,
		),

		fx.Populate(&h, &reporter),
	)

	if err := di.Err(); err != nil {
		return errgo.Wrap(err, "fx")
	}

	var errChan = make(chan error, 1)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		errChan <- http.ListenAndServe(h.config.ListenAddr(), mux) //nolint:gosec
	}()

	go func() {
		errChan <- h.start()
	}()

	// register for interrupt (Ctrl+C) and SIGTERM (docker)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	defer h.Close()

	select {
	case err := <-errChan:
		return err
	case <-sigChan:
		logger.Info("receive signal, shutdown")
		return nil
	}
}

const groupID = "my-group"

func newKafkaReader(c config.AppConfig) *kafka.Reader {
	topics := []string{
		"chii.bangumi.chii_subject_fields",
		"chii.bangumi.chii_subjects",
		"chii.bangumi.chii_members",
	}
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{c.KafkaBroker},
		GroupID:     groupID,
		GroupTopics: topics,
	})
}
