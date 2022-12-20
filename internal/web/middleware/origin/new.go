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

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/config/env"
	"github.com/bangumi/server/internal/web/res"
)

func New(allowed string) fiber.Handler {
	if !env.Production {
		return dev(allowed)
	}
	return func(ctx *fiber.Ctx) error {
		if ctx.Method() == http.MethodGet {
			return ctx.Next()
		}

		origin := ctx.Get(fiber.HeaderOrigin)
		if origin == "" {
			return res.BadRequest("empty origin is not allowed")
		}
		if origin != allowed {
			return res.BadRequest("cross-site request is not allowed")
		}

		return ctx.Next()
	}
}

func dev(allowed string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Next()
	}
}
