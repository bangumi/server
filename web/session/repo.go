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

	"github.com/goccy/go-json"
	"github.com/gookit/goutil/timex"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
)

const Key = "sessionID"

type persistSession struct {
	CreatedAt time.Time
	ExpiredAt time.Time
	Key       string
	Value     Session
	UserID    model.UIDType
}

type Repo interface {
	Create(ctx context.Context, key string, userID model.UIDType, regTime time.Time) (Session, error)
	Get(ctx context.Context, key string) (persistSession, error)
	RevokeUser(ctx context.Context, userID model.UIDType) (keys []string, err error)
	Revoke(ctx context.Context, key string) error
}

func NewMysqlRepo(q *query.Query, logger *zap.Logger) Repo {
	return mysqlRepo{q: q, log: logger.Named("session.mysqlRepo")}
}

var ErrKeyConflict = errors.New("conflict key in database")

type mysqlRepo struct {
	q   *query.Query
	log *zap.Logger
}

func (r mysqlRepo) Create(ctx context.Context, key string, userID model.UIDType, regTime time.Time) (Session, error) {
	tx := r.q.Begin()

	_, err := tx.WithContext(ctx).WebSession.Where(tx.WebSession.Key.Eq(key)).First()
	if err == nil {
		return Session{}, ErrKeyConflict
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return Session{}, errgo.Wrap(err, "orm.Tx.Where.First")
	}

	createdAt := time.Now().Unix()
	expiredAt := createdAt + timex.OneWeekSec
	s := Session{RegTime: regTime, UserID: userID, ExpiredAt: expiredAt}

	encodedJSON, err := json.Marshal(s)
	if err != nil {
		return Session{}, errgo.Wrap(err, "json.Marshal")
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
		return Session{}, errgo.Wrap(err, "orm.Tx.Create")
	}

	err = tx.Commit()
	if err != nil {
		return Session{}, errgo.Wrap(err, "tx.Commit")
	}

	return s, nil
}

func (r mysqlRepo) Get(ctx context.Context, key string) (persistSession, error) {
	s, err := r.q.WithContext(ctx).WebSession.
		Where(r.q.WebSession.Key.Eq(key), r.q.WebSession.ExpiredAt.Gte(time.Now().Unix())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return persistSession{}, domain.ErrNotFound
		}

		return persistSession{}, errgo.Wrap(err, "orm.Tx.Where.First")
	}

	var v Session
	if err = json.Unmarshal(s.Value, &v); err != nil {
		return persistSession{}, errgo.Wrap(err, "json.Unmarshal")
	}

	return persistSession{
		Key:       s.Key,
		Value:     v,
		CreatedAt: time.Unix(s.CreatedAt, 0),
		ExpiredAt: time.Unix(s.ExpiredAt, 0),
		UserID:    s.UserID,
	}, nil
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

func (r mysqlRepo) RevokeUser(ctx context.Context, userID model.UIDType) ([]string, error) {
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
