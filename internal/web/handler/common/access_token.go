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
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/web/accessor"
	"github.com/bangumi/server/internal/web/cookie"
	"github.com/bangumi/server/internal/web/handler/internal/ctxkey"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/session"
)

func (h Common) MiddlewareAccessTokenAuth(ctx *fiber.Ctx) error {
	var a = accessor.Get()
	defer accessor.Put(a)
	a.FillBasicInfo(ctx)

	authorization := ctx.Get(fiber.HeaderAuthorization)
	if authorization == "" {
		ctx.Context().SetUserValue(ctxkey.User, a)
		return ctx.Next()
	}

	key, token, found := strings.Cut(authorization, " ")
	if !found {
		return res.Unauthorized("invalid http Authorization header, missing scope or missing token")
	}

	if key != "Bearer" {
		return res.Unauthorized("http Authorization header has wrong scope")
	}

	auth, err := h.auth.GetByToken(ctx.Context(), token)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || errors.Is(err, session.ErrExpired) {
			cookie.Clear(ctx, session.CookieKey)
			return res.Unauthorized("access token has been expired or doesn't exist")
		}

		return errgo.Wrap(err, "auth.GetByToken")
	}

	a.SetAuth(auth)

	ctx.Context().SetUserValue(ctxkey.User, a)
	return ctx.Next()
}
