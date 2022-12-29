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

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pm"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h PrivateMessage) ListRelated(c *fiber.Ctx) error {
	accessor := h.Common.GetHTTPAccessor(c)
	relatedID, err := req.ParsePrivateMessageID(c.Params("id"))
	if err != nil {
		return err
	}
	list, err := h.pmRepo.ListRelated(c.Context(), accessor.ID, relatedID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return res.ErrNotFound
		case errors.Is(err, pm.ErrPmDeleted):
		case errors.Is(err, pm.ErrPmNotOwned):
			return res.BadRequest(err.Error())
		}
		return res.InternalError(c, err, "failed to list related private messages")
	}
	userIDs := []model.UserID{list[0].SenderID, list[0].ReceiverID}
	users, err := h.ctrl.GetUsersByIDs(c.Context(), userIDs)
	if err != nil {
		return res.InternalError(c, err, "failed to get users")
	}
	data := slice.Map(list, func(v pm.PrivateMessage) res.PrivateMessage {
		return res.ConvertModelPrivateMessage(v, users)
	})
	return res.JSON(c, data)
}
