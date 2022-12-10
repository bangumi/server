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

package pm

import (
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/web/res"
)

func (h PrivateMessage) ListRecentContact(c *fiber.Ctx) error {
	accessor := h.Common.GetHTTPAccessor(c)
	contactIDs, err := h.pmRepo.ListRecentContact(c.UserContext(), accessor.ID)
	if err != nil {
		return res.InternalError(c, err, "failed to list recent contact")
	}
	contacts, err := h.ctrl.GetUsersByIDs(c.Context(), contactIDs)
	if err != nil {
		return res.InternalError(c, err, "failed to get contacts")
	}
	return res.JSON(c, slice.MapFilter(contactIDs, func(v model.UserID) (res.User, bool) {
		if m, ok := contacts[v]; ok {
			return res.ConvertModelUser(m), ok
		}
		return res.User{}, false
	}))
}
