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

//go:build !dev

package referer

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/res/code"
)

func New(referer string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ref := c.Get(fiber.HeaderReferer)
		if ref == "" || strings.HasPrefix(ref, referer) {
			return c.Next()
		}

		return res.HTTPError(c, code.BadRequest, "bad referer, cross-site api request is not allowed")
	}
}
