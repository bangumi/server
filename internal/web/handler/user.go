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
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) GetCurrentUser(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)
	if !u.Login || u.ID == 0 {
		return res.Unauthorized("need Login")
	}

	user, err := h.u.GetByID(c.Context(), u.ID)
	if err != nil {
		return h.InternalError(c, err, "failed to get user")
	}

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

func (h Handler) GetUser(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}
	if len(username) >= 32 {
		return res.BadRequest("username is too long")
	}

	user, err := h.u.GetByName(c.Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("can't find user with username " + strconv.Quote(username))
		}

		return h.InternalError(c, err, "failed to get user by username", zap.String("username", username))
	}

	var r = convertModelUser(user)
	return res.JSON(c, r)
}

func (h Handler) GetUserAvatar(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}
	if len(username) >= 32 {
		return res.BadRequest("username is too long")
	}

	user, err := h.u.GetByName(c.Context(), username)
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

func convertModelUser(u model.User) res.User {
	return res.User{
		ID:        u.ID,
		URL:       "https://bgm.tv/user/" + u.UserName,
		Username:  u.UserName,
		Nickname:  u.NickName,
		UserGroup: u.UserGroup,
		Avatar:    res.UserAvatar(u.Avatar),
		Sign:      u.Sign,
	}
}
