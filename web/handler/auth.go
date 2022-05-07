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

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gookit/goutil/timex"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/strutil"
	"github.com/bangumi/server/web/handler/cachekey"
	"github.com/bangumi/server/web/handler/ctxkey"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/res/code"
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

func (h Handler) RevokeSession(c *fiber.Ctx) error {
	var r req.RevokeSession
	if err := json.UnmarshalNoEscape(c.Body(), r); err != nil {
		return res.WithError(c, err, code.BadRequest, "can't validate request body")
	}

	if err := h.v.Struct(r); err != nil {
		return res.WithError(c, err, code.BadRequest, "can't validate request body")
	}

	return c.JSON("session revoked")
}

func (h Handler) PrivateLogin(c *fiber.Ctx) error {
	allowed, remain, err := h.rateLimit.Allowed(c.Context(), c.Context().RemoteIP().String())
	if err != nil {
		return res.InternalError(c, err, "failed to apply rate limit")
	}

	if !allowed {
		return res.HTTPError(c, code.TooManyRequests, "Too many requests, you are not allowed to log in for a while.")
	}

	var r req.UserLogin
	if err = json.UnmarshalNoEscape(c.Body(), &r); err != nil {
		return res.WithError(c, err, code.UnprocessableEntity, "can't decode request body as json")
	}

	if err = h.v.Struct(r); err != nil {
		return res.WithError(c, err, code.BadRequest, "can't validate request body")
	}

	ok, err := h.captcha.Verify(c.Context(), r.HCaptchaResponse)
	if err != nil {
		return res.WithError(c, err, code.BadRequest, "Captcha verify http request error")
	}

	if !ok {
		return res.HTTPError(c, code.BadGateway, "failed to pass captcha verify, please re-do")
	}

	return h.privateLogin(c, r, remain)
}

func (h Handler) privateLogin(c *fiber.Ctx, r req.UserLogin, remain int) error {
	login, ok, err := h.a.Login(c.Context(), r.Email, r.Password)
	if err != nil {
		return res.WithError(c, err, code.InternalServerError, "Unexpected error happened when trying to log in")
	}

	if !ok {
		return res.JSON(c.Status(fiber.StatusUnauthorized), res.Error{
			Title:       "Unauthorized",
			Description: "Email or Password is not correct",
			Details:     res.LoginRemain{Remain: remain},
		})
	}

	key, s, err := h.session.Create(c.Context(), login)
	if err != nil {
		return res.InternalError(c, err, "Unexpected Session Manager Error")
	}

	if err = h.rateLimit.Reset(c.Context(), c.Context().RemoteIP().String()); err != nil {
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

	user, err := h.u.GetByID(c.Context(), s.UserID)
	if err != nil {
		return res.InternalError(c, err, "failed to get user")
	}

	return res.JSON(c, res.User{
		ID:        user.ID,
		URL:       "https://bgm.tv/user/" + user.UserName,
		Username:  user.UserName,
		Nickname:  user.NickName,
		UserGroup: user.UserGroup,
		Avatar:    res.Avatar{}.Fill(user.Avatar),
		Sign:      user.Sign,
	})
}
