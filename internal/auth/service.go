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
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"errors"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/auth/internal/cachekey"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/cache"
	"github.com/bangumi/server/internal/pkg/errgo"
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

func (s service) GetByID(ctx context.Context, userID model.UserID) (Auth, error) {
	var cacheKey = cachekey.User(userID)

	var a UserInfo
	ok, err := s.cache.Get(ctx, cacheKey, &a)
	if err != nil {
		return Auth{}, errgo.Wrap(err, "cache.Get")
	}

	if !ok {
		var u user.User
		u, err = s.user.GetByID(ctx, userID)
		if err != nil {
			return Auth{}, errgo.Wrap(err, "AuthRepo.GetByID")
		}

		a = UserInfo{
			RegTime: u.RegistrationTime,
			ID:      u.ID,
			GroupID: u.UserGroup,
		}

		_ = s.cache.Set(ctx, cacheKey, a, time.Hour)
	}

	permission, err := s.getPermission(ctx, a.GroupID)
	if err != nil {
		return Auth{}, err
	}

	return Auth{
		RegTime:    a.RegTime,
		ID:         a.ID,
		GroupID:    a.GroupID,
		Permission: permission,
	}, nil
}

func (s service) Login(ctx context.Context, email, password string) (Auth, bool, error) {
	var a, hashedPassword, err = s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return Auth{}, false, nil
		}

		return Auth{}, false, errgo.Wrap(err, "repo.GetByEmail")
	}

	ok, err := s.ComparePassword(hashedPassword, password)
	if err != nil {
		s.log.Error("unexpected error when comparing password with bcrypt", zap.Error(err))
		return Auth{}, false, err
	}
	if !ok {
		return Auth{}, false, nil
	}

	p, err := s.getPermission(ctx, a.GroupID)
	if err != nil {
		return Auth{}, false, err
	}

	return Auth{
		RegTime:    a.RegTime,
		ID:         a.ID,
		GroupID:    a.GroupID,
		Permission: p,
	}, true, nil
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

		_ = s.cache.Set(ctx, cacheKey, a, time.Hour)
	}

	permission, err := s.getPermission(ctx, a.GroupID)
	if err != nil {
		return Auth{}, err
	}

	return Auth{
		RegTime:    a.RegTime,
		ID:         a.ID,
		GroupID:    a.GroupID,
		Permission: permission,
	}, nil
}

func (s service) ComparePassword(hashed []byte, password string) (bool, error) {
	p := preProcessPassword(password)

	if err := bcrypt.CompareHashAndPassword(hashed, p); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}

		return false, errgo.Wrap(err, "bcrypt.CompareHashAndPassword")
	}

	return true, nil
}

func preProcessPassword(s string) []byte {
	// don't know why old code base use md5 to hash password first
	p := md5.Sum([]byte(s)) //nolint:gosec

	return []byte(hex.EncodeToString(p[:]))
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

func (s service) CreateAccessToken(
	ctx context.Context, userID model.UserID, name string, expiration time.Duration,
) (string, error) {
	token, err := s.repo.CreateAccessToken(ctx, userID, name, expiration)
	return token, errgo.Wrap(err, "repo.CreateAccessToken")
}

func (s service) ListAccessToken(ctx context.Context, userID model.UserID) ([]AccessToken, error) {
	tokens, err := s.repo.ListAccessToken(ctx, userID)
	return tokens, errgo.Wrap(err, "repo.ListAccessToken")
}

func (s service) DeleteAccessToken(ctx context.Context, id uint32) (bool, error) {
	result, err := s.repo.DeleteAccessToken(ctx, id)
	return result, errgo.Wrap(err, "repo.DeleteAccessToken")
}

func (s service) GetTokenByID(ctx context.Context, id uint32) (AccessToken, error) {
	result, err := s.repo.GetTokenByID(ctx, id)
	return result, errgo.Wrap(err, "repo.GetTokenByID")
}
