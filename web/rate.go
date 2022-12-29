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
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/rate"
	"github.com/bangumi/server/web/rate/action"
	"github.com/bangumi/server/web/res"
)

type baseHandler interface {
	GetHTTPAccessor(c *fiber.Ctx) *accessor.Accessor
}

// rateMiddleware require Handler.NeedLogin before this middleware.
func rateMiddleware(r rate.Manager, h baseHandler, action action.Action, limit rate.Limit) fiber.Handler {
	return func(c *fiber.Ctx) error {
		a := h.GetHTTPAccessor(c)
		if !a.Login {
			return res.Unauthorized("login required")
		}

		allowed, _, err := r.AllowAction(c.UserContext(), a.ID, action, limit)
		if err != nil {
			return errgo.Wrap(err, "rate.Manager.AllowAction")
		}
		if !allowed {
			return c.SendStatus(http.StatusTooManyRequests)
		}

		return c.Next()
	}
}
