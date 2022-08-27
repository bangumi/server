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

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/web/frontend"
)

func (h Handler) PageLogin(c *fiber.Ctx) error {
	v := h.GetHTTPAccessor(c)
	var u model.User
	if v.Login {
		var err error
		u, err = h.ctrl.GetUser(c.Context(), v.ID)

		if err != nil {
			return errgo.Wrap(err, "failed to get current user")
		}
	}

	return h.render(c, frontend.TplLogin, frontend.Login{Title: "Login", User: u})
}
