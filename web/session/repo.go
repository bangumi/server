// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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
	"gorm.io/gorm"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
)

type Repo interface {
	Create(ctx context.Context, key string, userID model.IDType, s Session) error
	Get(ctx context.Context, key string) (dao.WebSession, error)
}

func NewMysqlRepo(q *query.Query) Repo {
	return mysqlRepo{q}
}

var ErrKeyConflict = errors.New("conflict key in database")

type mysqlRepo struct {
	q *query.Query
}

func (r mysqlRepo) Create(ctx context.Context, key string, userID model.IDType, s Session) error {
	tx := r.q.Begin()

	_, err := tx.WithContext(ctx).WebSession.Where(tx.WebSession.Key.Eq(key)).First()
	if err == nil {
		return ErrKeyConflict
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errgo.Wrap(err, "orm.Tx.Where.First")
	}

	createAt := time.Now().Unix()

	encodedJSON, err := json.Marshal(s)
	if err != nil {
		return errgo.Wrap(err, "json.Marshal")
	}

	webSession := dao.WebSession{
		Key:       key,
		UserID:    userID,
		Value:     encodedJSON,
		CreateAt:  createAt,
		ExpiredAt: createAt + timex.OneWeekSec,
	}

	err = tx.WebSession.WithContext(ctx).Create(&webSession)
	if err != nil {
		return errgo.Wrap(err, "orm.Tx.Create")
	}

	err = tx.Commit()
	if err != nil {
		return errgo.Wrap(err, "tx.Commit")
	}

	return nil
}

func (r mysqlRepo) Get(ctx context.Context, key string) (dao.WebSession, error) {
	s, err := r.q.WithContext(ctx).WebSession.Where(r.q.WebSession.Key.Eq(key)).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return dao.WebSession{}, errgo.Wrap(err, "orm.Tx.Where.First")
		}

		return dao.WebSession{}, domain.ErrNotFound
	}

	return *s, nil
}
