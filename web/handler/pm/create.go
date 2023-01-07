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
	"github.com/samber/lo"

	"github.com/bangumi/server/ctrl"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/pm"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h PrivateMessage) Create(c echo.Context) error {
	accessor := accessor.GetFromCtx(c)
	var r req.PrivateMessageCreate
	if err := c.Echo().JSONSerializer.Deserialize(c, &r); err != nil {
		return res.JSONError(c, err)
	}

	if err := h.Common.V.Struct(r); err != nil {
		return h.ValidationError(c, err)
	}
	receiverIDs := slice.Map(r.ReceiverIDs, func(v uint32) model.UserID { return v })

	msgs, err := h.ctrl.CreatePrivateMessage(
		c.Request().Context(),
		accessor.ID,
		receiverIDs,
		pm.IDFilter{Type: null.NewFromPtr(r.RelatedID)},
		r.Title,
		r.Content)
	if err != nil {
		switch {
		case errors.Is(err, ctrl.ErrPmBlocked):
		case errors.Is(err, ctrl.ErrPmNotAFriend):
		case errors.Is(err, ctrl.ErrPmNotAllReceiversExist):
		case errors.Is(err, ctrl.ErrPmReceiverReject):
		case errors.Is(err, pm.ErrPmRelatedNotExists):
		case errors.Is(err, pm.ErrPmInvalidOperation):
			return res.BadRequest(err.Error())
		}
		return res.InternalError(c, err, "failed to create private message(s)")
	}
	userIDs := make([]model.UserID, len(r.ReceiverIDs)+1)
	copy(userIDs, receiverIDs)
	userIDs[len(userIDs)-1] = accessor.ID
	users, err := h.u.GetByIDs(c.Request().Context(), lo.Uniq(userIDs))
	if err != nil {
		return res.InternalError(c, err, "failed to get users")
	}
	return c.JSON(http.StatusOK, slice.Map(msgs, func(v pm.PrivateMessage) res.PrivateMessage {
		return res.ConvertModelPrivateMessage(v, users)
	}))
}
