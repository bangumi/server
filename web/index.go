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

package web

import (
	_ "embed" //nolint:revive
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/bangumi/server/config/env"
)

//go:embed index.html
var indexPageHTML string

func indexPage() echo.HandlerFunc {
	if env.Production {
		return func(c echo.Context) error {
			return c.Redirect(http.StatusFound, "https://github.com/bangumi/")
		}
	}

	return func(c echo.Context) error {
		return c.HTML(http.StatusOK, indexPageHTML)
	}
}
