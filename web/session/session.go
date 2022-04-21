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
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/goccy/go-json"
	"github.com/gookit/goutil/timex"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/rand"
	"github.com/bangumi/server/model"
)

const keyLength = 32
const keyPrefix = "chii:web:session:"

type Manager interface {
	Create(ctx context.Context, a domain.Auth) (string, Session, error)
	Get(ctx context.Context, key string) (Session, error)
}

func New(r *redis.Client, q *query.Query, log *zap.Logger) Manager {
	return manager{r: r, q: q, log: log.Named("web.Session.Manager")}
}

type manager struct {
	r   *redis.Client
	q   *query.Query
	log *zap.Logger
}

func (m manager) Create(ctx context.Context, a domain.Auth) (string, Session, error) {
	key := rand.SecureRandomString(keyLength)
	s := Session{}

	err := m.q.WithContext(ctx).WebSession.Create(&dao.WebSession{
		Key:      key,
		CreateAt: time.Now().Unix(),
		Value:    "",
	})
	if err != nil {
		return "", Session{}, errgo.Wrap(err, "dal")
	}

	if err = m.r.Set(ctx, keyPrefix+keyPrefix, s, timex.OneWeek).Err(); err != nil {
		return "", Session{}, errgo.Wrap(err, "redis.Set")
	}

	return key, Session{}, nil
}

func (m manager) Get(ctx context.Context, key string) (Session, error) {
	result, err := m.r.Get(ctx, keyPrefix+keyPrefix).Bytes()
	if err != nil {
		return Session{}, errgo.Wrap(err, "redis.Set")
	}

	var s Session
	if err := json.Unmarshal(result, &s); err != nil {
		m.log.Warn("can't decode session from redis")
	}

	return Session{}, nil
}

type Session struct {
	UID model.IDType
}
