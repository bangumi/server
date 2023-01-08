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
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pm"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h PrivateMessage) ListRelated(c echo.Context) error {
	accessor := accessor.GetFromCtx(c)
	relatedID, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}
	list, err := h.pmRepo.ListRelated(c.Request().Context(), accessor.ID, relatedID)
	if err != nil {
		switch {
		case errors.Is(err, gerr.ErrNotFound):
			return res.ErrNotFound
		case errors.Is(err, pm.ErrPmDeleted):
		case errors.Is(err, pm.ErrPmNotOwned):
			return res.BadRequest(err.Error())
		}
		return res.InternalError(c, err, "failed to list related private messages")
	}
	userIDs := []model.UserID{list[0].SenderID, list[0].ReceiverID}
	users, err := h.u.GetByIDs(c.Request().Context(), userIDs)
	if err != nil {
		return res.InternalError(c, err, "failed to get users")
	}
	data := slice.Map(list, func(v pm.PrivateMessage) res.PrivateMessage {
		return res.ConvertModelPrivateMessage(v, users)
	})
	return c.JSON(http.StatusOK, data)
}
