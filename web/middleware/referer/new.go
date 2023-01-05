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

package referer

import (
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/config/env"
	"github.com/bangumi/server/web/res"
)

const HeaderReferer = "Referer"

func New(referer string) echo.MiddlewareFunc {
	if env.Production {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				ref := c.Request().Header.Get(HeaderReferer)
				if ref == "" || strings.HasPrefix(ref, referer) {
					return next(c)
				}

				return res.BadRequest("bad referer, cross-site api request is not allowed")
			}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			return next(ctx)
		}
	}
}
