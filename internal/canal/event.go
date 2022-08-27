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
	"os"
	"strconv"

	"github.com/segmentio/kafka-go"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/driver"
	"github.com/bangumi/server/internal/metrics"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/search"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/web/session"
)

type eventHandler struct {
	config  config.AppConfig
	session session.Manager
	log     *zap.Logger
	search  search.Client
	dryRun  bool
	reader  *kafka.Reader
}

func (e *eventHandler) start() <-chan error {
	errChan := make(chan error)

	go func() {
		for {
			message, err := e.reader.FetchMessage(context.Background())
			if err != nil {
				continue
			}

			err = e.onMessage(message)
			if err != nil {
				e.log.Error("failed to handle kafka message",
					zap.String("topic", message.Topic),
					zap.ByteString("key", message.Key),
					zap.ByteString("value", message.Value),
					zap.Error(err),
				)
				continue
			}

			_ = e.reader.CommitMessages(context.Background(), message)
		}
	}()

	return errChan
}

func (e *eventHandler) Close() error {
	return errgo.Wrap(e.reader.Close(), "kafka.Close")
}

func newEventHandler(
	log *zap.Logger,
	appConfig config.AppConfig,
	session session.Manager,
	reader *kafka.Reader,
	search search.Client,
) *eventHandler {
	dryRun, _ := strconv.ParseBool(os.Getenv("DRY_RUN"))

	return &eventHandler{
		dryRun:  dryRun,
		config:  appConfig,
		session: session,
		reader:  reader,
		search:  search,
		log:     log.Named("eventHandler"),
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

func (e *eventHandler) OnUserPasswordChange(id model.UserID) error {
	e.log.Info("user change password", log.UserID(id))
	if e.dryRun {
		e.log.Info("dry-run enabled, skip handler")
		return nil
	}

	if err := e.session.RevokeUser(context.Background(), id); err != nil {
		e.log.Error("failed to revoke user", log.UserID(id), zap.Error(err))
		return errgo.Wrap(err, "session.RevokeUser")
	}

	return nil
}
