// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/web/res"
)

func (h Handler) GetCurrentUser(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)
	if !u.login || u.ID == 0 {
		return fiber.ErrUnauthorized
	}

	user, err := h.u.GetByID(c.Context(), u.ID)
	if err != nil {
		return errgo.Wrap(err, "repo")
	}

	var me = res.User{
		ID:        user.ID,
		URL:       "https://bgm.tv/user/" + user.UserName,
		Username:  user.UserName,
		Nickname:  user.NickName,
		UserGroup: user.UserGroup,
		Avatar:    res.Avatar{}.Fill(user.Avatar),
		Sign:      user.Sign,
	}

	return c.JSON(me)
}

func (h Handler) GetUser(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return fiber.NewError(http.StatusBadRequest, "missing require parameters `username`")
	}
	if len(username) >= 32 {
		return res.HTTPError(c, http.StatusBadRequest, "username is too long")
	}

	user, err := h.u.GetByName(c.Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.HTTPError(c, http.StatusNotFound, "can't find user with username "+strconv.Quote(username))
		}

		h.log.Error("unexpected error happened", zap.Error(err))
		return errgo.Wrap(err, "user.GetByName")
	}

	var r = res.User{
		ID:        user.ID,
		URL:       "https://bgm.tv/user/" + user.UserName,
		Username:  user.UserName,
		Nickname:  user.NickName,
		UserGroup: user.UserGroup,
		Avatar:    res.Avatar{}.Fill(user.Avatar),
		Sign:      user.Sign,
	}

	return c.JSON(r)
}
