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

	"github.com/gookit/goutil/timex"
	"go.uber.org/zap"

	"github.com/bangumi/server/cache"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/rand"
)

const keyLength = 32
const keyPrefix = "chii:web:session:"

var ErrSessionExpired = errors.New("your session has been expired")

type Manager interface {
	Create(ctx context.Context, a domain.Auth) (string, Session, error)
	Get(ctx context.Context, key string) (Session, error)
	Revoke(ctx context.Context, key string) error
}

func New(r cache.Generic, repo Repo, log *zap.Logger) Manager {
	return manager{cache: r, repo: repo, log: log.Named("web.Session.Manager")}
}

type manager struct {
	repo  Repo
	cache cache.Generic
	log   *zap.Logger
}

func (m manager) Create(ctx context.Context, a domain.Auth) (string, Session, error) {
	var key string
	var err error
	s := Session{UID: a.ID}

	for i := 0; i < 5; i++ {
		key = rand.Base62String(keyLength)
		err = m.repo.Create(ctx, key, a.ID, s)
		if err != nil {
			if !errors.Is(err, ErrKeyConflict) {
				return "", Session{}, errgo.Wrap(err, "un-expected error when creating session")
			}
			// key conflict, re-generate new key and retry
			key = rand.Base62String(keyLength)
		}
	}

	if err := m.cache.Set(ctx, keyPrefix+key, s, timex.OneWeek); err != nil {
		return "", Session{}, errgo.Wrap(err, "redis.Set")
	}

	return key, s, nil
}

func (m manager) Get(ctx context.Context, key string) (Session, error) {
	var s Session

	ok, err := m.cache.Get(ctx, keyPrefix+key, &s)
	if err != nil {
		return Session{}, errgo.Wrap(err, "redis.Set")
	}
	if ok {
		return s, nil
	}

	ws, err := m.repo.Get(ctx, key)
	if err != nil {
		return Session{}, errgo.Wrap(err, "mysqlRepo.Get")
	}

	if time.Now().After(ws.ExpiredAt) {
		return Session{}, ErrSessionExpired
	}

	s = ws.Value

	if err := m.cache.Set(ctx, keyPrefix+key, s, timex.OneDay*3); err != nil {
		m.log.Panic("failed to set cache")
	}

	return s, nil
}

func (m manager) Revoke(ctx context.Context, key string) error {
	// TODO implement me
	panic("implement me")
}
