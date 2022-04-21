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

package auth

import (
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/bangumi/server/cache"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
)

func NewService(repo domain.AuthRepo) domain.AuthService {
	return service{
		c:    cache.NewMemoryCache(),
		repo: repo,
	}
}

type service struct {
	c    cache.Generic
	repo domain.AuthRepo
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

	if !ok || err != nil {
		return domain.Auth{}, false, errgo.Wrap(err, "ComparePassword")
	}

	p, err := s.GetPermission(ctx, a.GroupID)
	if err != nil {
		return domain.Auth{}, false, err
	}
	a.Permission = p

	return a, true, nil

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

	err := bcrypt.CompareHashAndPassword(hashed, p[:])
	if err != nil {
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

func (s service) GetPermission(ctx context.Context, id uint8) (domain.Permission, error) {
	var p domain.Permission
	key := strconv.FormatUint(uint64(id), 10)
	ok, err := s.c.Get(ctx, key, &p)
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

	_ = s.c.Set(ctx, key, p, time.Minute)

	return p, nil
}
