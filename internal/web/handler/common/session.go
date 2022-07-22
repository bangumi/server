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

package common

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/accessor"
	"github.com/bangumi/server/internal/web/cookie"
	"github.com/bangumi/server/internal/web/handler/internal/ctxkey"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/session"
)

func (h Common) MiddlewareSessionAuth(c *fiber.Ctx) error {
	var a = accessor.Get()
	defer accessor.Put(a)
	a.FillBasicInfo(c)

	value := utils.UnsafeString(c.Context().Request.Header.Cookie(session.CookieKey))
	if value != "" {
		s, err := h.getSession(c, value)
		if err != nil {
			if errors.Is(err, session.ErrExpired) || errors.Is(err, domain.ErrNotFound) {
				cookie.Clear(c, session.CookieKey)
				return c.Next()
			}

			h.log.Error("failed to get session", zap.Error(err), a.Log())
			return res.InternalError(c, err, "failed to read session, please try clear your browser cookies and re-try")
		}

		auth, err := h.auth.GetByID(c.Context(), s.UserID)
		if err != nil {
			return h.InternalError(c, err, "failed to user with permission", a.Log(), log.UserID(s.UserID))
		}

		a.SetAuth(auth)
	}

	c.Context().SetUserValue(ctxkey.User, a)

	return c.Next()
}

func (h Common) getSession(c *fiber.Ctx, value string) (session.Session, error) {
	s, err := h.session.Get(c.Context(), value)
	if err != nil {
		return session.Session{}, errgo.Wrap(err, "sessionManager.Get")
	}

	return s, nil
}
