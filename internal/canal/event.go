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

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/dal"
	"github.com/bangumi/server/internal/dal/query"
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
	session session.Manager
	log     *zap.Logger
	q       *query.Query
	search  search.Client
	dryRun  bool
}

func newEventHandler(
	q *query.Query,
	log *zap.Logger,
	session session.Manager,
	search search.Client,
) *eventHandler {
	dryRun, _ := strconv.ParseBool(os.Getenv("DRY_RUN"))

	return &eventHandler{
		dryRun:  dryRun,
		session: session,
		q:       q,
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
