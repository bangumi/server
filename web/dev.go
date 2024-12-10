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
	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/internal/pkg/random"
	"github.com/bangumi/server/web/req/cf"
)

func genFakeRequestID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		devRequestID := "fake-ray-" + random.Base62String(10)
		c.Request().Header.Set(cf.HeaderRequestID, devRequestID)
		c.Set(cf.HeaderRequestID, devRequestID)

		return next(c)
	}
}
