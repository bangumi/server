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

package req

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/bangumi/server/web/res"
)

func JSON(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		contentType := c.Request().Header.Get(echo.HeaderContentType)
		//nolint:staticcheck
		if contentType == echo.MIMEApplicationJSON || contentType == echo.MIMEApplicationJSONCharsetUTF8 {
			return next(c)
		}

		return c.JSON(http.StatusUnsupportedMediaType,
			res.Error{
				Title:       "Unsupported Media Type",
				Description: `request with body must set "content-type" header to "application/json"`,
			},
		)
	}
}
