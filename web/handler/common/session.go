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
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/cookie"
	"github.com/bangumi/server/web/handler/internal/ctxkey"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/session"
	"github.com/bangumi/server/web/util"
)

func (h Common) MiddlewareSessionAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var a = accessor.Get()
		defer accessor.Put(a)
		a.FillBasicInfo(c)

		co, err := c.Cookie(session.CookieKey)
		if err != nil {
			return errgo.Wrap(err, "get cookie")
		}

		if co.Value != "" {
			s, err := h.getSession(c, co.Value)
			if err != nil {
				if errors.Is(err, session.ErrExpired) || errors.Is(err, domain.ErrNotFound) {
					cookie.Clear(c, session.CookieKey)
					goto Next
				}

				h.log.Error("failed to get session", zap.Error(err), a.Log())
				return c.JSON(http.StatusInternalServerError,
					res.Error{
						Title:       "internal server error",
						Details:     util.DetailWithErr(c, err),
						Description: "failed to read session, please try clear your browser cookies and re-try",
					})
			}

			auth, err := h.auth.GetByID(c.Request().Context(), s.UserID)
			if err != nil {
				return errgo.Wrap(err, "failed to user with permission")
			}

			a.SetAuth(auth)
		}

	Next:
		c.Set(ctxkey.User, a)

		return next(c)
	}
}

func (h Common) getSession(c echo.Context, value string) (session.Session, error) {
	s, err := h.session.Get(c.Request().Context(), value)
	if err != nil {
		return session.Session{}, errgo.Wrap(err, "sessionManager.Get")
	}

	return s, nil
}
