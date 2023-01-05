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
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/res"
)

func (h User) GetCurrent(c echo.Context) error {
	u := accessor.GetFromCtx(c)
	if !u.Login || u.ID == 0 {
		return res.Unauthorized("need Login")
	}

	user, err := h.user.GetByID(c.Request().Context(), u.ID)
	if err != nil {
		return errgo.Wrap(err, "failed to get user")
	}

	return c.JSON(http.StatusOK, res.User{
		ID:        user.ID,
		URL:       "https://bgm.tv/user/" + user.UserName,
		Username:  user.UserName,
		Nickname:  user.NickName,
		UserGroup: user.UserGroup,
		Avatar:    res.UserAvatar(user.Avatar),
		Sign:      user.Sign,
	})
}
