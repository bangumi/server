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
	"github.com/bangumi/server/internal/web/handler/common"
	"github.com/bangumi/server/internal/web/rate"
	"github.com/bangumi/server/internal/web/rate/action"
)

// Rate require Handler.NeedLogin before this middleware.
func Rate(r rate.Manager, h common.Common, action action.Action, limit rate.Limit) fiber.Handler {
	return func(c *fiber.Ctx) error {
		a := h.GetHTTPAccessor(c)
		if !a.Login {
			panic("Rate are handing not login-ed request")
		}

		allowed, _, err := r.AllowAction(c.Context(), a.ID, action, limit)
		if err != nil {
			return errgo.Wrap(err, "rate.Manager.AllowAction")
		}
		if !allowed {
			return c.SendStatus(http.StatusTooManyRequests)
		}

		return c.Next()
	}
}
