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

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/random"
	"github.com/bangumi/server/internal/pkg/timex"
)

const defaultKeyLength = 32
const CookieKey = "sessionID"
const redisKeyPrefix = "chii:web:session:"

var ErrExpired = errors.New("your session has been expired")

type Manager interface {
	Create(ctx context.Context, a domain.Auth) (string, Session, error)
	Get(ctx context.Context, key string) (Session, error)
	Revoke(ctx context.Context, key string) error
	RevokeUser(ctx context.Context, id model.UserID) error
}

func New(c cache.Generic, repo Repo, log *zap.Logger) Manager {
	return manager{cache: c, repo: repo, log: log.Named("web.Session.Manager")}
}

type manager struct {
	repo  Repo
	cache cache.Generic
	log   *zap.Logger
}

func defaultKeyGenerator() string {
	return random.Base62String(defaultKeyLength)
}

func (m manager) Create(ctx context.Context, a domain.Auth) (string, Session, error) {
	key, s, err := m.repo.Create(ctx, a.ID, a.RegTime, defaultKeyGenerator)

	if err != nil {
		m.log.Error("un-expected error when creating session", zap.Error(err))
		return "", Session{}, errgo.Wrap(err, "un-expected error when creating session")
	}

	if err := m.cache.Set(ctx, redisKeyPrefix+key, s, timex.OneWeek); err != nil {
		return "", Session{}, errgo.Wrap(err, "redis.Set")
	}

	return key, s, nil
}

func (m manager) Get(ctx context.Context, key string) (Session, error) {
	var s Session

	ok, err := m.cache.Get(ctx, redisKeyPrefix+key, &s)
	if err != nil {
		return Session{}, errgo.Wrap(err, "redis.Get")
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
		return Session{}, ErrExpired
	}

	s = ws.Value
	s.ExpiredAt = ws.ExpiredAt.Unix()

	// 缓存3天或缓存者到token失效
	ttl := minDur(timex.OneDay*3, ws.ExpiredAt.Sub(now))

	if err := m.cache.Set(ctx, redisKeyPrefix+key, s, ttl); err != nil {
		m.log.Panic("failed to set cache")
	}

	return s, nil
}

func (m manager) Revoke(ctx context.Context, key string) error {
	if err := m.repo.Revoke(ctx, key); err != nil {
		return errgo.Wrap(err, "repo.Revoke")
	}

	err := m.cache.Del(ctx, redisKeyPrefix+key)
	return errgo.Wrap(err, "redisCache.Del")
}

func (m manager) RevokeUser(ctx context.Context, id model.UserID) error {
	keys, err := m.repo.RevokeUser(ctx, id)
	if err != nil {
		return errgo.Wrap(err, "repo.Revoke")
	}

	for i, key := range keys {
		keys[i] = redisKeyPrefix + key
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
