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

package handler

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/strutil"
	"github.com/bangumi/server/web/handler/cachekey"
	"github.com/bangumi/server/web/handler/ctxkey"
	"github.com/bangumi/server/web/util"
)

const headerCFRay = "Cf-Ray"
const ctxKeyVisitor = "access-user"

func (h Handler) MiddlewareAccessUser() fiber.Handler {
	pool := sync.Pool{New: func() interface{} { return &accessor{} }}

	return func(ctx *fiber.Ctx) error {
		var a = pool.Get().(*accessor) //nolint:forcetypeassert
		defer pool.Put(a)

		if err := h.fill(ctx, a); err != nil {
			return err
		}

		ctx.Context().SetUserValue(ctxKeyVisitor, a)

		return ctx.Next()
	}
}
func (h Handler) getUser(c *fiber.Ctx) *accessor {
	u, ok := c.Context().UserValue(ctxkey.User).(*accessor) // get visitor
	if !ok {
		panic("can't convert type")
	}

	return u
}

func (h Handler) fill(c *fiber.Ctx, a *accessor) error {
	a.login = false
	a.cfRay = c.Get(headerCFRay)
	a.ip = util.RequestIP(c)

	authorization := c.Get(fiber.HeaderAuthorization)
	if authorization == "" {
		return nil
	}

	key, token := strutil.Partition(authorization, ' ')
	if key != "Bearer" {
		return fiber.NewError(fiber.StatusUnauthorized, "http Authorization header has wrong scope")
	}

	var cacheKey = cachekey.Auth(token)

	ok, err := h.cache.Get(c.Context(), cacheKey, &a.Auth)
	if err != nil {
		return errgo.Wrap(err, "cache.Get")
	}
	if ok {
		a.login = true
		return nil
	}

	if a.Auth, err = h.a.GetByToken(c.Context(), token); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fiber.NewError(fiber.StatusUnauthorized, "access token has been expired or doesn't exist")
		}

		return errgo.Wrap(err, "auth.GetByToken")
	}

	a.login = true

	if err := h.cache.Set(c.Context(), cacheKey, a.Auth, time.Hour); err != nil {
		logger.Error("can't set cache value", zap.Error(err))
	}

	return nil
}

type accessor struct {
	domain.Auth
	cfRay string
	ip    net.IP
	login bool
}

func (a *accessor) AllowNSFW() bool {
	return a.login && a.Auth.AllowNSFW()
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
