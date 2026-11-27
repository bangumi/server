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

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/user"
)

const TokenTypeOauthToken = 0
const TokenTypeAccessToken = 1

const tokenCacheTTL = 10 * time.Minute

// cacheEntry marks a token as revoked so a concurrent cache miss can't write it back.
type cacheEntry struct {
	Revoked bool `json:"revoked,omitempty"`
	UserInfo
}

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
	var cacheKey = cachekey.AuthToken(token)
	var entry cacheEntry

	ok, err := s.cache.Get(ctx, cacheKey, &entry)
	if err != nil {
		return Auth{}, errgo.Wrap(err, "cache.Get")
	}

	if ok && entry.Revoked {
		return Auth{}, gerr.ErrNotFound
	}

	a = entry.UserInfo
	// A legacy cache entry has no defined fields, while a decoded empty ACL has an empty slice.
	if !ok || a.Permission.DefinedFields == nil {
		a, err = s.repo.GetByToken(ctx, token)
		if err != nil {
			return Auth{}, errgo.Wrap(err, "AuthRepo.GetByToken")
		}

		// Legacy entries expire on their original TTL. Cache fills must not overwrite revocation markers.
		if !ok {
			_, _ = s.cache.SetNX(ctx, cacheKey, cacheEntry{UserInfo: a}, tokenCacheTTL)
		}
	}

	groupPermission, err := s.getPermission(ctx, a.GroupID)
	if err != nil {
		return Auth{}, err
	}
	permission := groupPermission.Merge(a.Permission)

	return Auth{
		Login:      true,
		RegTime:    a.RegTime,
		ID:         a.ID,
		GroupID:    a.GroupID,
		Permission: permission,
		Scope:      a.Scope,
		Legacy:     a.Legacy,
	}, nil
}

// RevokeToken invalidates cached authorization for a token that has been expired or deleted in database.
// The revoked marker is written unconditionally, and cache fill always uses SetNX, so it can't be overwritten.
func (s service) RevokeToken(ctx context.Context, token string) error {
	err := s.cache.Set(ctx, cachekey.AuthToken(token), cacheEntry{Revoked: true}, tokenCacheTTL)
	return errgo.Wrap(err, "cache.Set")
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
