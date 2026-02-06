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

package origin

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/bangumi/server/config/env"
	"github.com/bangumi/server/web/res"
)

func New(allowed string) echo.MiddlewareFunc {
	if !env.Production {
		return dev(allowed)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodGet {
				return next(c)
			}

			origin := c.Get(echo.HeaderOrigin)
			if origin == "" {
				return res.BadRequest("empty origin is not allowed")
			}
			if origin != allowed {
				return res.BadRequest("cross-site request is not allowed")
			}

			return next(c)
		}
	}
}

func dev(_ string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
