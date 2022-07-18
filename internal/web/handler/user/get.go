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

package user

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/web/res"
)

func (h User) Get(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}
	if len(username) >= 32 {
		return res.BadRequest("username is too long")
	}

	user, err := h.user.GetByName(c.Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("can't find user with username " + strconv.Quote(username))
		}

		return h.InternalError(c, err, "failed to get user by username", zap.String("username", username))
	}

	var r = res.ConvertModelUser(user)

	return res.JSON(c, r)
}

func (h User) GetAvatar(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}
	if len(username) >= 32 {
		return res.BadRequest("username is too long")
	}

	user, err := h.user.GetByName(c.Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("can't find user with username " + strconv.Quote(username))
		}

		return h.InternalError(c, err, "failed to get user by username")
	}

	l, ok := res.UserAvatar(user.Avatar).Select(c.Query("type"))
	if !ok {
		return res.BadRequest("bad avatar type: " + c.Query("type"))
	}

	return c.Redirect(l)
}
