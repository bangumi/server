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
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/web/frontend"
	"github.com/bangumi/server/internal/web/res"
)

// PageListAccessToken 直接调用了 `query.Query`。
// 因为这只是个demo网页，在后续重构为 api 的时候仍然应该替换为 service 。
func (h Handler) PageListAccessToken(c *fiber.Ctx) error {
	v := h.getHTTPAccessor(c)
	if !v.login {
		return c.Redirect("/demo/login")
	}

	u, err := h.u.GetByID(c.Context(), v.ID)
	if err != nil {
		return res.InternalError(c, err, "failed to get current user")
	}

	tokens, err := h.a.ListAccessToken(c.Context(), v.ID)
	if err != nil {
		return res.InternalError(c, err, "failed to fetch access tokens")
	}

	return h.render(c, frontend.TplListAccessToken, frontend.ListAccessToken{Tokens: tokens, User: u})
}

func (h Handler) PageCreateAccessToken(c *fiber.Ctx) error {
	v := h.getHTTPAccessor(c)
	if !v.login {
		return c.Redirect("/demo/login")
	}

	u, err := h.u.GetByID(c.Context(), v.ID)
	if err != nil {
		return res.InternalError(c, err, "failed to get current user")
	}

	return h.render(c, frontend.TplCreateAccessToken, frontend.CreateAccessToken{User: u})
}
