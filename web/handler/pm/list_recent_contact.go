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
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/res"
)

func (h PrivateMessage) ListRecentContact(c echo.Context) error {
	accessor := accessor.FromCtx(c)
	contactIDs, err := h.pmRepo.ListRecentContact(c.Request().Context(), accessor.ID)
	if err != nil {
		return res.InternalError(c, err, "failed to list recent contact")
	}
	contacts, err := h.u.GetByIDs(c.Request().Context(), contactIDs)
	if err != nil {
		return res.InternalError(c, err, "failed to get contacts")
	}
	return c.JSON(http.StatusOK, slice.MapFilter(contactIDs, func(v model.UserID) (res.User, bool) {
		if m, ok := contacts[v]; ok {
			return res.ConvertModelUser(m), ok
		}
		return res.User{}, false
	}))
}
