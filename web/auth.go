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

package web

import (
	"errors"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/strutil"
	"github.com/bangumi/server/web/util"
)

const headerCFRay = "Cf-Ray"
const ctxKeyUser = "access-user"

// should bump cache version every time we change domain.Auth.
const authCacheKeyPrefix = "chii:auth:1:access-token:"

func newAccessUserMiddleware(h Handler) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		a, err := h.getUser(ctx)
		if err != nil {
			return err
		}

		ctx.Context().SetUserValue(ctxKeyUser, a)

		return ctx.Next()
	}
}

func (h Handler) getUser(c *fiber.Ctx) (accessor, error) {
	a := accessor{
		Auth:  domain.Auth{},
		login: false,
		cfRay: c.Get(headerCFRay),
		ip:    util.RequestIP(c),
	}

	authorization := c.Get(fiber.HeaderAuthorization)
	if authorization == "" {
		logger.Info("access without token", a.LogField())

		return a, nil
	}

	key, value := strutil.Partition(authorization, ' ')
	if key != "bearer" {
		return a, fiber.NewError(fiber.StatusUnauthorized,
			"http Authorization header has wrong scope")
	}

	var u domain.Auth

	var cacheKey = authCacheKeyPrefix + value

	ok, err := h.cache.Get(c.Context(), cacheKey, &u)
	if err != nil {
		return a, errgo.Wrap(err, "cache")
	}

	if !ok {
		u, err = h.a.GetByToken(c.Context(), value)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return a, nil
			}

			return a, errgo.Wrap(err, "repo")
		}

		if err := h.cache.Set(c.Context(), cacheKey, u, time.Minute); err != nil {
			logger.Error("can't set cache value", zap.Error(err))
		}
	}

	a.Auth = u
	a.login = true

	return a, nil
}

type accessor struct {
	domain.Auth
	cfRay string
	ip    net.IP
	login bool
}

func (a accessor) LogField() zap.Field {
	return zap.Object("request", a)
}

func (a accessor) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	if a.ID != 0 {
		encoder.AddUint32("user_id", a.ID)
	}
	encoder.AddString("ip", a.ip.String())
	encoder.AddString("id", a.cfRay)

	return nil
}
