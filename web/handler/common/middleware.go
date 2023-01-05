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
	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/res"
)

var errNeedLogin = res.Unauthorized("this API need authorization")

func (h Common) NeedLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if u := accessor.FromCtx(c); !u.Login {
			return errNeedLogin
		}

		return next(c)
	}
}
