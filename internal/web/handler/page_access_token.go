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

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/frontend"
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
		return h.InternalError(c, err, "failed to get current user", log.UserID(v.ID))
	}

	tokens, err := h.a.ListAccessToken(c.Context(), v.ID)
	if err != nil {
		return h.InternalError(c, err, "failed to fetch access tokens", log.UserID(v.ID))
	}

	clients, err := h.oauth.GetClientByID(c.Context(), clientIDs(tokens)...)
	if err != nil {
		return h.InternalError(c, err, "failed to fetch access tokens", log.UserID(v.ID))
	}

	return h.render(c, frontend.TplListAccessToken, frontend.ListAccessToken{Tokens: tokens, User: u, Clients: clients})
}

func (h Handler) PageCreateAccessToken(c *fiber.Ctx) error {
	v := h.getHTTPAccessor(c)
	if !v.login {
		return c.Redirect("/demo/login")
	}

	u, err := h.u.GetByID(c.Context(), v.ID)
	if err != nil {
		return h.InternalError(c, err, "failed to get current user", log.UserID(v.ID))
	}

	return h.render(c, frontend.TplCreateAccessToken, frontend.CreateAccessToken{User: u})
}

func clientIDs(tokens []domain.AccessToken) []string {
	var clientIDs = make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		clientIDs[token.ClientID] = struct{}{}
	}

	var s = make([]string, 0, len(clientIDs))
	for token := range clientIDs {
		s = append(s, token)
	}

	return s
}
