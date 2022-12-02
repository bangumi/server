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

package res

import (
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

func JSON(c *fiber.Ctx, v any) error {
	data, err := sonic.Marshal(v)
	if err != nil {
		c.Status(http.StatusInternalServerError).Context().SetBodyString("failed to encode json body: " + err.Error())
		return nil
	}

	c.Context().SetContentType(fiber.MIMEApplicationJSON)
	c.Context().Response.SetBodyRaw(data)

	return nil
}
