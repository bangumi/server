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

package session

import (
	"context"
	"errors"
	"time"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/gtime"
)

func NewMysqlRepo(q *query.Query, logger *zap.Logger) Repo {
	return mysqlRepo{q: q, log: logger.Named("session.mysqlRepo")}
}

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (r mysqlRepo) Create(
	ctx context.Context, userID model.UserID, regTime time.Time, keyGen func() string,
) (string, Session, error) {
	createdAt := time.Now().Unix()
	expiredAt := createdAt + gtime.OneWeekSec
	s := Session{RegTime: regTime, UserID: userID, ExpiredAt: expiredAt}
	encodedJSON, err := sonic.Marshal(s)
	if err != nil {
		return "", Session{}, errgo.Wrap(err, "sonic.MarshalWithOption")
	}

	tx := r.q.Begin()
	for i := 0; i < defaultRetry; i++ {
		key := keyGen()

		c, err := tx.WithContext(ctx).WebSession.Where(tx.WebSession.Key.Eq(key)).Count()
		if err != nil {
			return "", Session{}, errgo.Wrap(err, "tx.WebSession.Count")
		}

		if c != 0 {
			// session id conflict, re-generate key
			continue
		}

		webSession := dao.WebSession{
			Key:       key,
			UserID:    userID,
			Value:     encodedJSON,
			CreatedAt: createdAt,
			ExpiredAt: expiredAt,
		}

		err = tx.WebSession.WithContext(ctx).Create(&webSession)
		if err != nil {
			return "", Session{}, errgo.Wrap(err, "orm.Tx.Create")
		}

		err = tx.Commit()
		if err != nil {
			return "", Session{}, errgo.Wrap(err, "tx.Commit")
		}

		return key, s, nil
	}

	return "", Session{}, errTooManyRetry
}

func (r mysqlRepo) Get(ctx context.Context, key string) (Session, error) {
	record, err := r.q.WithContext(ctx).WebSession.
		Where(r.q.WebSession.Key.Eq(key), r.q.WebSession.ExpiredAt.Gte(time.Now().Unix())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Session{}, domain.ErrNotFound
		}

		return Session{}, errgo.Wrap(err, "orm.Tx.Where.First")
	}

	var s Session
	if err = sonic.Unmarshal(record.Value, &s); err != nil {
		return Session{}, errgo.Wrap(err, "sonic.Unmarshal")
	}

	s.UserID = record.UserID
	s.CreatedAt = record.CreatedAt
	s.ExpiredAt = record.ExpiredAt

	return s, nil
}

func (r mysqlRepo) Revoke(ctx context.Context, key string) error {
	_, err := r.q.WithContext(ctx).WebSession.Where(r.q.WebSession.Key.Eq(key)).
		UpdateSimple(r.q.WebSession.ExpiredAt.Value(time.Now().Unix()))
	if err != nil {
		r.log.Error("unexpected error", zap.Error(err))
		return errgo.Wrap(err, "gorm.UpdateSimple")
	}

	return nil
}

func (r mysqlRepo) RevokeUser(ctx context.Context, userID model.UserID) ([]string, error) {
	s, err := r.q.WithContext(ctx).WebSession.Where(r.q.WebSession.UserID.Eq(userID)).Find()
	if err != nil {
		r.log.Error("unexpected error", zap.Error(err))
		return nil, errgo.Wrap(err, "gorm.Find")
	}

	_, err = r.q.WithContext(ctx).WebSession.Where(r.q.WebSession.UserID.Eq(userID)).
		UpdateSimple(r.q.WebSession.ExpiredAt.Value(time.Now().Unix()))
	if err != nil {
		return nil, errgo.Wrap(err, "dal.UpdateSimple")
	}

	var keys = make([]string, len(s))
	for i, session := range s {
		keys[i] = session.Key
	}
	return keys, nil
}
