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

//nolint:depguard
package canal

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
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
	e, err := getEventHandler()
	if err != nil {
		return err
	}

	shutdown := make(chan int)
	sigChan := make(chan os.Signal, 1)
	// register for interrupt (Ctrl+C) and SIGTERM (docker)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("receive signal, shutdown")
		e.Close()
		shutdown <- 1
	}()

	select {
	case err := <-e.start(context.Background()):
		return err
	case <-shutdown:
		return nil
	}
}

func getEventHandler() (*eventHandler, error) {
	var h *eventHandler

	err := fx.New(
		logger.FxLogger(),
		config.Module,

		dal.Module,

		// driver and connector
		fx.Provide(
			driver.NewMysqlConnectionPool, metrics.NewScope,
			driver.NewRedisClient, logger.Copy, cache.NewRedisCache,
			subject.NewMysqlRepo,
			search.New,
			session.NewMysqlRepo, session.New,

			func(c config.AppConfig) *kafka.Reader {
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
			},

			newEventHandler,
		),

		fx.Populate(&h),
	).Err()

	if err != nil {
		return nil, errgo.Wrap(err, "fx")
	}

	return h, nil
}
