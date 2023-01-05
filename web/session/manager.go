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

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/internal/pkg/random"
)

const defaultKeyLength = 32
const CookieKey = "chiiNextSessionID"
const redisKeyPrefix = "chii:web:session:"

var ErrExpired = errors.New("your session has been expired")

type Manager interface {
	Create(ctx context.Context, a auth.Auth) (string, Session, error)
	Get(ctx context.Context, key string) (Session, error)
	Revoke(ctx context.Context, key string) error
	RevokeUser(ctx context.Context, id model.UserID) error
}

func New(c cache.RedisCache, repo Repo, log *zap.Logger) Manager {
	return manager{cache: c, repo: repo, log: log.Named("web.Session.Manager")}
}

type manager struct {
	repo  Repo
	cache cache.RedisCache
	log   *zap.Logger
}

func defaultKeyGenerator() string {
	return random.Base62String(defaultKeyLength)
}

func (m manager) Create(ctx context.Context, a auth.Auth) (string, Session, error) {
	key, s, err := m.repo.Create(ctx, a.ID, a.RegTime, defaultKeyGenerator)

	if err != nil {
		m.log.Error("un-expected error when creating session", zap.Error(err))
		return "", Session{}, errgo.Wrap(err, "un-expected error when creating session")
	}

	if err := m.cache.Set(ctx, redisKeyPrefix+key, s, gtime.OneWeek); err != nil {
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

	s, err = m.repo.Get(ctx, key)
	if err != nil {
		return Session{}, errgo.Wrap(err, "mysqlRepo.Get")
	}

	if s.ExpiredAt <= time.Now().Unix() {
		return Session{}, ErrExpired
	}

	// 缓存3天或缓存者到token失效
	ttl := generic.Min(gtime.OneDaySec*3, s.ExpiredAt)

	if err := m.cache.Set(ctx, redisKeyPrefix+key, s, gtime.Second(ttl)); err != nil {
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

	if len(keys) == 0 {
		return nil
	}

	for i, key := range keys {
		keys[i] = redisKeyPrefix + key
	}

	err = m.cache.Del(ctx, keys...)
	return errgo.Wrap(err, "redisCache.Del")
}
