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

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/pkg/timex"
	"github.com/bangumi/server/internal/web/cookie"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/session"
)

func (h Handler) RevokeSession(c *fiber.Ctx) error {
	var r req.RevokeSession
	if err := json.UnmarshalNoEscape(c.Body(), r); err != nil {
		return res.JSONError(c, err)
	}

	if err := h.v.Struct(r); err != nil {
		return h.ValidationError(c, err)
	}

	return c.JSON("session revoked")
}

func (h Handler) PrivateLogin(c *fiber.Ctx) error {
	var r req.UserLogin
	if err := json.UnmarshalNoEscape(c.Body(), &r); err != nil {
		return res.JSONError(c, err)
	}

	if err := h.v.Struct(r); err != nil {
		return h.ValidationError(c, err)
	}

	ok, err := h.captcha.Verify(c.Context(), r.HCaptchaResponse)
	if err != nil {
		return res.FromError(c, err, http.StatusBadGateway, "Failed to connect to hCaptcha server")
	}

	if !ok {
		return res.JSON(c.Status(http.StatusBadRequest), res.Error{
			Title:       utils.StatusMessage(http.StatusBadRequest),
			Description: "can't validate request body",
			Details:     []string{"未通过hCaptcha验证"},
		})
	}

	a := h.getHTTPAccessor(c)
	allowed, remain, err := h.rateLimit.Allowed(c.Context(), a.ip.String())
	if err != nil {
		return h.InternalError(c, err, "failed to apply rate limit", a.LogRequestID())
	}

	if !allowed {
		return res.NewError(http.StatusTooManyRequests, "Too many requests, you are not allowed to log in for a while.")
	}

	return h.privateLogin(c, a, r, remain)
}

func (h Handler) privateLogin(c *fiber.Ctx, a *accessor, r req.UserLogin, remain int) error {
	login, ok, err := h.a.Login(c.Context(), r.Email, r.Password)
	if err != nil {
		return h.InternalError(c, err, "Unexpected error when logging in")
	}

	if !ok {
		return res.JSON(c.Status(http.StatusUnauthorized), res.Error{
			Title:       "Unauthorized",
			Description: "Email or Password is not correct",
			Details:     res.LoginRemain{Remain: remain},
		})
	}

	key, s, err := h.session.Create(c.Context(), login)
	if err != nil {
		return h.InternalError(c, err, "failed to create session", a.LogRequestID())
	}

	if err = h.rateLimit.Reset(c.Context(), c.Context().RemoteIP().String()); err != nil {
		h.log.Error("failed to reset login rate limit", zap.Error(err), a.LogRequestID())
	}

	c.Cookie(&fiber.Cookie{
		Name:     session.Key,
		Value:    key,
		MaxAge:   timex.OneWeekSec * 2,
		Secure:   !config.Development,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	})

	user, err := h.u.GetByID(c.Context(), s.UserID)
	if err != nil {
		return h.InternalError(c, err, "failed to get user by user id", a.LogRequestID())
	}

	h.log.Info("user login", log.UserID(user.ID))

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
	if a := h.getHTTPAccessor(c); !a.login {
		return res.Unauthorized("you are not logged-in")
	}

	sessionID := utils.UnsafeString(c.Context().Request.Header.Cookie(session.Key))
	if err := h.session.Revoke(c.Context(), sessionID); err != nil {
		return h.InternalError(c, err, "failed to revoke session", zap.String("session_id", sessionID))
	}

	cookie.Clear(c, session.Key)
	c.Status(http.StatusNoContent)

	return nil
}
