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

package notification

import (
	"context"

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func NewMysqlRepo(q *query.Query, log *zap.Logger) (Repo, error) {
	return mysqlRepo{q: q, log: log.Named("notification.mysqlRepo")}, nil
}

func (r mysqlRepo) count(ctx context.Context, userID model.UserID) (int64, error) { //nolint:golint,unused
	count, err := r.q.Notification.WithContext(ctx).Where(
		r.q.Notification.ReceiverID.Eq(userID),
		r.q.Notification.Status.Eq(uint8(StatusUnread)),
	).Count()
	if err != nil {
		r.log.Error("unexpected error", zap.Error(err))
		return 0, errgo.Wrap(err, "dal")
	}
	return count, nil
}

func (r mysqlRepo) Count(ctx context.Context, userID model.UserID) (int64, error) {
	member, err := r.q.Member.WithContext(ctx).Where(
		r.q.Member.ID.Eq(userID),
	).Select(r.q.Member.ID, r.q.Member.NewNotify).First()
	if err != nil {
		r.log.Error("unexpected error", zap.Error(err))
		return 0, errgo.Wrap(err, "dal")
	}
	return int64(member.NewNotify), nil
}
