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
		return res.InternalError(c, err, "failed to list related private messages")
	}
	if len(list) > 0 {
		userIDs := []model.UserID{list[0].SenderID, list[0].ReceiverID}
		users, err := h.ctrl.GetUsersByIDs(c.Context(), userIDs)
		if err != nil {
			return res.InternalError(c, err, "failed to get users")
		}
		data := slice.Map(list, func(v model.PrivateMessage) res.PrivateMessage {
			return res.ConvertModelPrivateMessage(v, users)
		})
		return res.JSON(c, data)
	}
	return res.NotFound("related private messages not found")
}
