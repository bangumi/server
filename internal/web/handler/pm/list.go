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
	"github.com/samber/lo"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h PrivateMessage) List(c *fiber.Ctx) error {
	accessor := h.Common.GetHTTPAccessor(c)
	folder, err := req.ParsePrivateMessageFolder(c.Query("folder"))
	if err != nil {
		return err
	}
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}
	ctx := c.UserContext()
	count, err := h.pmRepo.CountByFolder(ctx, accessor.ID, folder)
	if err != nil {
		return res.InternalError(c, err, "failed to count private messages")
	}
	list, err := h.pmRepo.List(ctx, accessor.ID, folder, page.Offset, page.Limit)
	if err != nil {
		return res.InternalError(c, err, "failed to list private messages")
	}
	if len(list) == 0 {
		return res.JSON(c, res.Paged{
			Data:   make([]res.PrivateMessage, 0),
			Total:  count,
			Limit:  page.Limit,
			Offset: page.Offset,
		})
	}
	userIDs := make([]model.UserID, len(list)+1)
	for i := range list {
		if folder == model.PrivateMessageFolderTypeInbox {
			userIDs[i] = list[i].Self.SenderID
		} else {
			userIDs[i] = list[i].Self.ReceiverID
		}
	}
	userIDs[len(userIDs)-1] = accessor.ID
	userIDs = lo.Uniq(userIDs)
	users, err := h.ctrl.GetUsersByIDs(c.Context(), userIDs)
	if err != nil {
		return res.InternalError(c, err, "failed to get users")
	}
	data := slice.Map(list, func(v model.PrivateMessageListItem) res.PrivateMessage {
		return res.ConvertModelPrivateMessageListItem(v, users)
	})
	return res.JSON(c, res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}
