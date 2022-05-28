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
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/gookit/goutil/timex"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/config"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/web/cookie"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/res/code"
	"github.com/bangumi/server/internal/web/session"
)

func (h Handler) RevokeSession(c *fiber.Ctx) error {
	var r req.RevokeSession
	if err := json.UnmarshalNoEscape(c.Body(), r); err != nil {
		return res.WithError(c, err, code.UnprocessableEntity, "can't parse request body as JSON")
	}

	if err := h.v.Struct(r); err != nil {
		return res.WithError(c, err, code.BadRequest, "can't validate request body")
	}

	return c.JSON("session revoked")
}

func (h Handler) PrivateLogin(c *fiber.Ctx) error {
	a := h.getHTTPAccessor(c)
	var r req.UserLogin
	if err := json.UnmarshalNoEscape(c.Body(), &r); err != nil {
		return res.WithError(c, err, code.UnprocessableEntity, "can't decode request body as json")
	}

	if err := h.v.Struct(r); err != nil {
		var validationErrors validator.ValidationErrors
		if ok := errors.As(err, &validationErrors); ok {
			var details = make([]string, len(validationErrors))
			for i, e := range validationErrors {
				// can translate each error one at a time.
				details[i] = e.Translate(h.validatorTranslation)
			}

			return res.JSON(c.Status(code.BadRequest), res.Error{
				Title:       http.StatusText(code.BadRequest),
				Description: "can't validate request body",
				Details:     details,
			})
		}

		return res.WithError(c, err, code.BadRequest, "can't validate request body")
	}

	allowed, remain, err := h.rateLimit.Allowed(c.Context(), c.Context().RemoteIP().String())
	if err != nil {
		h.log.Error("failed to apply rate limit", zap.Error(err), a.LogRequestID())
		return res.InternalError(c, err, "failed to apply rate limit")
	}

	if !allowed {
		return res.HTTPError(c, code.TooManyRequests, "Too many requests, you are not allowed to log in for a while.")
	}

	ok, err := h.captcha.Verify(c.Context(), r.HCaptchaResponse)
	if err != nil {
		return res.WithError(c, err, code.BadGateway, "Captcha verify http request error")
	}

	if !ok {
		return res.JSON(c.Status(code.BadRequest), res.Error{
			Title:       http.StatusText(code.BadRequest),
			Description: "can't validate request body",
			Details:     []string{"未通过hCaptcha验证"},
		})
	}

	return h.privateLogin(c, a, r, remain)
}

func (h Handler) privateLogin(c *fiber.Ctx, a *accessor, r req.UserLogin, remain int) error {
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
		h.log.Error("failed to create session", zap.Error(err), a.LogRequestID())
		return res.InternalError(c, err, "Unexpected Session Manager Error")
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
		SameSite: fiber.CookieSameSiteStrictMode,
	})

	user, err := h.u.GetByID(c.Context(), s.UserID)
	if err != nil {
		h.log.Error("failed to get user by user id", zap.Error(err), a.LogRequestID())
		return res.InternalError(c, err, "failed to get user")
	}

	h.log.Info("user login", log.UserID(user.ID))

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

func (h Handler) PrivateLogout(c *fiber.Ctx) error {
	if a := h.getHTTPAccessor(c); !a.login {
		return res.HTTPError(c, code.Unauthorized, "you are not logged-in")
	}

	sessionID := utils.UnsafeString(c.Context().Request.Header.Cookie(session.Key))
	if err := h.session.Revoke(c.Context(), sessionID); err != nil {
		h.log.Error("failed to revoke session", zap.Error(err), zap.String("session_id", sessionID))
		return res.InternalError(c, err, "failed to revoke session")
	}

	cookie.Clear(c, session.Key)
	c.Status(code.NoContent)

	return nil
}
