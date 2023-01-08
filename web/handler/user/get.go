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
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/res"
)

func (h User) Get(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}
	if len(username) >= 32 {
		return res.BadRequest("username is too long")
	}

	user, err := h.user.GetByName(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("can't find user with username " + strconv.Quote(username))
		}

		return errgo.Wrap(err, "failed to get user by username")
	}

	var r = res.ConvertModelUser(user)

	return c.JSON(http.StatusOK, r)
}

func (h User) GetAvatar(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}
	if len(username) >= 32 {
		return res.BadRequest("username is too long")
	}

	user, err := h.user.GetByName(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("can't find user with username " + strconv.Quote(username))
		}

		return errgo.Wrap(err, "failed to get user by username")
	}

	l, ok := res.UserAvatar(user.Avatar).Select(c.QueryParam("type"))
	if !ok {
		return res.BadRequest("bad avatar type: " + c.QueryParam("type"))
	}

	return c.Redirect(http.StatusFound, l)
}
