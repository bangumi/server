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
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	"github.com/bangumi/server/internal/config/env"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/gtime"
	"github.com/bangumi/server/web/cookie"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/session"
)

func (h Handler) PrivateLogin(c *fiber.Ctx) error {
	var r req.UserLogin
	if err := sonic.Unmarshal(c.Body(), &r); err != nil {
		return res.JSONError(c, err)
	}

	if err := h.Common.V.Struct(r); err != nil {
		return h.ValidationError(c, err)
	}

	ok, err := h.captcha.Verify(c.UserContext(), r.CaptchaResponse)
	if err != nil {
		return res.FromError(c, err, http.StatusBadGateway, "Failed to connect to hCaptcha server")
	}

	if !ok {
		return res.JSON(c.Status(http.StatusBadRequest), res.Error{
			Title:       utils.StatusMessage(http.StatusBadRequest),
			Description: "can't validate request body",
			Details:     []string{"未通过captcha验证"},
		})
	}

	a := h.GetHTTPAccessor(c)
	allowed, remain, err := h.rateLimit.Login(c.UserContext(), a.IP.String())
	if err != nil {
		return errgo.Wrap(err, "failed to apply rate limit")
	}

	if !allowed {
		return res.NewError(http.StatusTooManyRequests, "Too many requests, you are not allowed to log in for a while.")
	}

	return h.privateLogin(c, r, remain)
}

func (h Handler) privateLogin(c *fiber.Ctx, r req.UserLogin, remain int) error {
	login, ok, err := h.a.Login(c.UserContext(), r.Email, r.Password)
	if err != nil {
		return errgo.Wrap(err, "Unexpected error when logging in")
	}

	if !ok {
		return res.JSON(c.Status(http.StatusUnauthorized), res.Error{
			Title:       "Unauthorized",
			Description: "Email or Password is not correct",
			Details:     res.LoginRemain{Remain: remain},
		})
	}

	key, s, err := h.session.Create(c.UserContext(), login)
	if err != nil {
		return errgo.Wrap(err, "failed to create session")
	}

	c.Cookie(&fiber.Cookie{
		Name:     session.CookieKey,
		Value:    key,
		MaxAge:   gtime.OneWeekSec * 2,
		Secure:   !env.Development,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	})

	user, err := h.ctrl.GetUser(c.UserContext(), s.UserID)
	if err != nil {
		return errgo.Wrap(err, "failed to get user by user id")
	}

	h.log.Info("user Login", user.ID.Zap())

	return res.JSON(c, res.User{
		ID:        user.ID,
		URL:       "https://bgm.tv/user/" + user.UserName,
		Username:  user.UserName,
		Nickname:  user.NickName,
		UserGroup: user.UserGroup,
		Avatar:    res.UserAvatar(user.Avatar),
		Sign:      user.Sign,
	})
}

func (h Handler) PrivateLogout(c *fiber.Ctx) error {
	if a := h.GetHTTPAccessor(c); !a.Login {
		return res.Unauthorized("you are not logged-in")
	}

	sessionID := utils.UnsafeString(c.Context().Request.Header.Cookie(session.CookieKey))
	if err := h.session.Revoke(c.UserContext(), sessionID); err != nil {
		return errgo.Wrap(err, "failed to revoke session")
	}

	cookie.Clear(c, session.CookieKey)
	c.Status(http.StatusNoContent)

	return nil
}
