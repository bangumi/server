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
	"strconv"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/bangumi/server/internal/cache"
	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
)

const TokenTypeOauthToken = 0
const TokenTypeAccessToken = 1

func NewService(repo domain.AuthRepo, user domain.UserRepo, logger *zap.Logger, c cache.Generic) domain.AuthService {
	return service{
		localCache: cache.NewMemoryCache(),
		cache:      c,
		repo:       repo,
		log:        logger.Named("auth.Service"),
		user:       user,
	}
}

type service struct {
	localCache cache.Generic
	cache      cache.Generic
	repo       domain.AuthRepo
	user       domain.UserRepo
	log        *zap.Logger
}

func (s service) GetByID(ctx context.Context, userID model.UserID) (domain.Auth, error) {
	u, err := s.user.GetByID(ctx, userID)
	if err != nil {
		return domain.Auth{}, errgo.Wrap(err, "AuthRepo.GetByToken")
	}

	p, err := s.GetPermission(ctx, u.UserGroup)
	if err != nil {
		return domain.Auth{}, err
	}

	return domain.Auth{
		RegTime:    u.RegistrationTime,
		ID:         u.ID,
		GroupID:    u.UserGroup,
		Permission: p,
	}, nil
}

func (s service) GetByIDWithCache(ctx context.Context, userID model.UserID) (domain.Auth, error) {
	var a domain.Auth

	var cacheKey = cachekey.User(userID)

	ok, err := s.cache.Get(ctx, cacheKey, &a)
	if err != nil {
		return domain.Auth{}, errgo.Wrap(err, "cache.GetByID")
	}
	if ok {
		if a.Permission, err = s.GetPermission(ctx, a.GroupID); err != nil {
			return domain.Auth{}, err
		}

		return a, nil
	}

	a, err = s.GetByID(ctx, userID)
	if err != nil {
		return domain.Auth{}, err
	}

	_ = s.cache.Set(ctx, cacheKey, a, time.Hour)

	return a, nil
}

func (s service) Login(ctx context.Context, email, password string) (domain.Auth, bool, error) {
	var a, hashedPassword, err = s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Auth{}, false, nil
		}

		return domain.Auth{}, false, errgo.Wrap(err, "repo.GetByEmail")
	}

	ok, err := s.ComparePassword(hashedPassword, password)
	if err != nil {
		s.log.Error("unexpected error when comparing password with bcrypt", zap.Error(err))
		return domain.Auth{}, false, err
	}
	if !ok {
		return domain.Auth{}, false, nil
	}

	p, err := s.GetPermission(ctx, a.GroupID)
	if err != nil {
		return domain.Auth{}, false, err
	}
	a.Permission = p

	return a, true, nil

}

// GetByTokenWithCache not sure should we cache it in service or let caller cache this.
func (s service) GetByTokenWithCache(ctx context.Context, token string) (domain.Auth, error) {
	var a domain.Auth

	var cacheKey = cachekey.Auth(token)

	ok, err := s.cache.Get(ctx, cacheKey, &a)
	if err != nil {
		return domain.Auth{}, errgo.Wrap(err, "cache.GetByID")
	}
	if ok {
		if a.Permission, err = s.GetPermission(ctx, a.GroupID); err != nil {
			return domain.Auth{}, err
		}

		return a, nil
	}

	a, err = s.GetByToken(ctx, token)
	if err != nil {
		return domain.Auth{}, err
	}

	_ = s.cache.Set(ctx, cacheKey, a, time.Hour)

	return a, nil
}

func (s service) GetByToken(ctx context.Context, token string) (domain.Auth, error) {
	a, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return domain.Auth{}, errgo.Wrap(err, "AuthRepo.GetByToken")
	}

	p, err := s.GetPermission(ctx, a.GroupID)
	if err != nil {
		return domain.Auth{}, err
	}

	a.Permission = p

	return a, nil
}

func (s service) ComparePassword(hashed []byte, password string) (bool, error) {
	p := preProcessPassword(password)

	if err := bcrypt.CompareHashAndPassword(hashed, p[:]); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}

		return false, errgo.Wrap(err, "bcrypt.CompareHashAndPassword")
	}

	return true, nil
}

func preProcessPassword(s string) [32]byte {
	// don't know why old code base use md5 to hash password first
	p := md5.Sum([]byte(s)) //nolint:gosec
	var md5Password [32]byte

	hex.Encode(md5Password[:], p[:])

	return md5Password
}

func (s service) GetPermission(ctx context.Context, id model.UserGroupID) (domain.Permission, error) {
	var p domain.Permission
	key := strconv.FormatUint(uint64(id), 10)
	ok, err := s.localCache.Get(ctx, key, &p)
	if err != nil {
		return domain.Permission{}, errgo.Wrap(err, "read cache")
	}

	if ok {
		return p, nil
	}

	p, err = s.repo.GetPermission(ctx, id)
	if err != nil {
		return domain.Permission{}, errgo.Wrap(err, "AuthRepo.GetPermission")
	}

	_ = s.localCache.Set(ctx, key, p, time.Minute)

	return p, nil
}

func (s service) CreateAccessToken(
	ctx context.Context, userID model.UserID, name string, expiration time.Duration,
) (string, error) {
	token, err := s.repo.CreateAccessToken(ctx, userID, name, expiration)
	return token, errgo.Wrap(err, "repo.CreateAccessToken")
}

func (s service) ListAccessToken(ctx context.Context, userID model.UserID) ([]domain.AccessToken, error) {
	tokens, err := s.repo.ListAccessToken(ctx, userID)
	return tokens, errgo.Wrap(err, "repo.ListAccessToken")
}

func (s service) DeleteAccessToken(ctx context.Context, id uint32) (bool, error) {
	result, err := s.repo.DeleteAccessToken(ctx, id)
	return result, errgo.Wrap(err, "repo.DeleteAccessToken")
}

func (s service) GetTokenByID(ctx context.Context, id uint32) (domain.AccessToken, error) {
	result, err := s.repo.GetTokenByID(ctx, id)
	return result, errgo.Wrap(err, "repo.GetTokenByID")
}
