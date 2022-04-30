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
	"net/http"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gookit/goutil/timex"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/internal/strutil"
	"github.com/bangumi/server/web/handler/cachekey"
	"github.com/bangumi/server/web/handler/ctxkey"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
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
func (h Handler) getHTTPAccessor(c *fiber.Ctx) *accessor {
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
	cfRay string
	ip    net.IP
	domain.Auth
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

func (h Handler) RevokeSession(c *fiber.Ctx) error {
	var r req.RevokeSession
	if err := c.BodyParser(r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(res.Error{
			Title:       "Bad Request",
			Details:     util.ErrDetail(c, err),
			Description: "can't decode request body",
		})
	}

	if err := h.v.Struct(r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(res.Error{
			Title:       "Bad Request",
			Details:     util.ErrDetail(c, err),
			Description: "can't validate request body",
		})
	}

	return c.JSON("session revoked")
}

func (h Handler) PrivateLogin(c *fiber.Ctx) error {
	contentType := utils.UnsafeString(c.Request().Header.ContentType())
	if contentType != fiber.MIMEApplicationJSON {
		return c.Status(fiber.StatusBadRequest).JSON(res.Error{
			Title:       "Bad Request",
			Description: "Must use 'application/json' as request content-type.",
		})
	}

	allowed, remain, err := h.rateLimit.Allowed(c.Context(), c.Context().RemoteIP().String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.Error{
			Title:   "Unexpected Error",
			Details: util.ErrDetail(c, err),
		})
	}

	if !allowed {
		return c.Status(fiber.StatusBadRequest).JSON(res.Error{
			Title:       "Bad Request",
			Description: "Too many requests, you are not allowed to log in for a while.",
		})
	}

	var r req.UserLogin
	if err = json.Unmarshal(c.Body(), &r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(res.Error{
			Title:       "Bad Request",
			Details:     util.ErrDetail(c, err),
			Description: "can't decode request body as json",
		})
	}

	if err = h.v.Struct(r); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(res.Error{
			Title:       "Bad Request",
			Details:     util.ErrDetail(c, err),
			Description: "can't validate request body",
		})
	}

	ok, err := h.captcha.Verify(c.Context(), r.HCaptchaResponse)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(res.Error{
			Title:       "Bad Gateway",
			Details:     util.ErrDetail(c, err),
			Description: "Captcha verify http request error",
		})
	}

	if !ok {
		return c.Status(fiber.StatusBadGateway).SendString("failed to pass captcha verify, please re-do")
	}

	return h.privateLogin(c, r, remain)
}

func (h Handler) privateLogin(c *fiber.Ctx, r req.UserLogin, remain int) error {
	login, ok, err := h.a.Login(c.Context(), r.Email, r.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.Error{
			Title:   "Unexpected Error",
			Details: util.ErrDetail(c, err),
		})
	}

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(res.Error{
			Title:       "Unauthorized",
			Description: "Email or Password is not correct",
			Details:     fiber.Map{"remain": remain},
		})
	}

	key, _, err := h.session.Create(c.Context(), login)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(res.Error{
			Title:   "Unexpected Session Manager Error",
			Details: util.ErrDetail(c, err),
		})
	}

	if err := h.rateLimit.Reset(c.Context(), c.Context().RemoteIP().String()); err != nil {
		h.log.Error("failed to reset login rate limit", zap.Error(err))
	}

	c.Cookie(&fiber.Cookie{
		Name:     "sessionID",
		Value:    key,
		Path:     "/",
		Domain:   "next.bgm.tv",
		MaxAge:   timex.OneWeekSec * 2,
		Secure:   true,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})

	return c.JSON("login")
}

// OldServerRevoke 旧服务器会发一个 HTTP 请求，revoke 掉某个用户所有的 token.
func (h Handler) OldServerRevoke(c *fiber.Ctx) error {
	userIDRaw := c.Params("user_id")
	userID, err := strparse.UserID(userIDRaw)
	if err != nil {
		return res.HTTPError(c, http.StatusBadRequest, err.Error())
	}

	err = h.session.RevokeUser(c.Context(), userID)

	return errgo.Wrap(err, "session.RevokeUser")
}
