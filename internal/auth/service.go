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

package auth

import (
	"context" //nolint:gosec
	"time"

	"github.com/trim21/errgo"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/auth/internal/cachekey"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/user"
)

const TokenTypeOauthToken = 0
const TokenTypeAccessToken = 1

func NewService(repo Repo, u user.Repo, logger *zap.Logger, c cache.RedisCache) Service {
	return service{
		permCache: cache.NewMemoryCache[user.GroupID, Permission](),
		cache:     c,
		repo:      repo,
		log:       logger.Named("auth.Service"),
		user:      u,
	}
}

type service struct {
	permCache *cache.MemoryCache[user.GroupID, Permission]
	cache     cache.RedisCache
	repo      Repo
	user      user.Repo
	log       *zap.Logger
}

func (s service) GetByToken(ctx context.Context, token string) (Auth, error) {
	var a UserInfo
	var cacheKey = cachekey.Auth(token)

	ok, err := s.cache.Get(ctx, cacheKey, &a)
	if err != nil {
		return Auth{}, errgo.Wrap(err, "cache.Get")
	}

	if !ok {
		a, err = s.repo.GetByToken(ctx, token)
		if err != nil {
			return Auth{}, errgo.Wrap(err, "AuthRepo.GetByID")
		}

		_ = s.cache.Set(ctx, cacheKey, a, time.Minute*10)
	}

	permission, err := s.getPermission(ctx, a.GroupID)
	if err != nil {
		return Auth{}, err
	}

	return Auth{
		Login:      true,
		RegTime:    a.RegTime,
		ID:         a.ID,
		GroupID:    a.GroupID,
		Permission: permission.Merge(a.Permission),
	}, nil
}

func (s service) getPermission(ctx context.Context, id user.GroupID) (Permission, error) {
	p, ok := s.permCache.Get(ctx, id)

	if ok {
		return p, nil
	}

	p, err := s.repo.GetPermission(ctx, id)
	if err != nil {
		return Permission{}, errgo.Wrap(err, "AuthRepo.GetPermission")
	}

	s.permCache.Set(ctx, id, p, time.Minute)

	return p, nil
}
