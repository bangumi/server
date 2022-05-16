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
	"github.com/bangumi/server/internal/random"
	"github.com/bangumi/server/model"
)

const keyLength = 32
const keyPrefix = "chii:web:session:"

var ErrExpired = errors.New("your session has been expired")

type Manager interface {
	Create(ctx context.Context, a domain.Auth) (string, Session, error)
	Get(ctx context.Context, key string) (Session, error)
	Revoke(ctx context.Context, key string) error
	RevokeUser(ctx context.Context, id model.UIDType) error
}

func New(c cache.Generic, repo Repo, log *zap.Logger) Manager {
	return manager{cache: c, repo: repo, log: log.Named("web.Session.Manager")}
}

type manager struct {
	repo  Repo
	cache cache.Generic
	log   *zap.Logger
}

func (m manager) Create(ctx context.Context, a domain.Auth) (string, Session, error) {
	var key string
	var s Session
	var err error

	for i := 0; i < 5; i++ {
		key = random.Base62String(keyLength)
		s, err = m.repo.Create(ctx, key, a.ID, a.RegTime)
		if err != nil {
			if errors.Is(err, ErrKeyConflict) {
				// key conflict, re-generate new key and retry
				key = random.Base62String(keyLength)
				continue
			}

			m.log.Error("un-expected error when creating session", zap.Error(err))
			return "", Session{}, errgo.Wrap(err, "un-expected error when creating session")
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

	now := time.Now()
	if now.After(ws.ExpiredAt) {
		return Session{}, domain.ErrNotFound
	}

	s = ws.Value
	s.ExpiredAt = ws.ExpiredAt.Unix()

	// 缓存3天或缓存者到token失效
	ttl := minDur(timex.OneDay*3, ws.ExpiredAt.Sub(now))

	if err := m.cache.Set(ctx, keyPrefix+key, s, ttl); err != nil {
		m.log.Panic("failed to set cache")
	}

	return s, nil
}

func (m manager) Revoke(ctx context.Context, key string) error {
	if err := m.repo.Revoke(ctx, key); err != nil {
		return errgo.Wrap(err, "repo.Revoke")
	}

	err := m.cache.Del(ctx, keyPrefix+key)
	return errgo.Wrap(err, "redisCache.Del")
}

func (m manager) RevokeUser(ctx context.Context, id model.UIDType) error {
	keys, err := m.repo.RevokeUser(ctx, id)
	if err != nil {
		return errgo.Wrap(err, "repo.Revoke")
	}

	for i, key := range keys {
		keys[i] = keyPrefix + key
	}

	err = m.cache.Del(ctx, keys...)
	return errgo.Wrap(err, "redisCache.Del")
}

func minDur(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
