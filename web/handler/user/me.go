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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/res"
)

type CurrentUser struct {
	Avatar           res.Avatar   `json:"avatar"`
	Sign             string       `json:"sign"`
	URL              string       `json:"url"`
	Username         string       `json:"username"`
	Nickname         string       `json:"nickname"`
	ID               model.UserID `json:"id"`
	UserGroup        uint8        `json:"user_group"`
	RegistrationTime time.Time    `json:"reg_time"`
	Email            string       `json:"email"`
	TimeOffset       int8         `json:"time_offset"`
}

func (h User) GetCurrent(c echo.Context) error {
	u := accessor.GetFromCtx(c)
	if !u.Login || u.ID == 0 {
		return res.Unauthorized("need Login")
	}

	user, err := h.user.GetFullUser(c.Request().Context(), u.ID)
	if err != nil {
		return errgo.Wrap(err, "failed to get user")
	}

	return c.JSON(http.StatusOK, CurrentUser{
		ID:               user.ID,
		URL:              "https://bgm.tv/user/" + user.UserName,
		Username:         user.UserName,
		Nickname:         user.NickName,
		UserGroup:        user.UserGroup,
		Avatar:           res.UserAvatar(user.Avatar),
		Sign:             user.Sign,
		RegistrationTime: user.RegistrationTime,
		Email:            user.Email,
		TimeOffset:       user.TimeOffset,
	})
}
