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
	"bytes"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/web/res"
)

var jsonType = []byte(fiber.MIMEApplicationJSON) //nolint:gochecknoglobals

func JSON(c *fiber.Ctx) error {
	if !bytes.Equal(c.Request().Header.ContentType(), jsonType) {
		return res.BadRequest("need content-type to be 'application/json'")
	}

	return c.Next()
}
