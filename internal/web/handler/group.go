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

package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/res"
)

const groupIconPrefix = "https://lain.bgm.tv/pic/icon/l/"

func (h Handler) GetGroupByName(c *fiber.Ctx) error {
	groupName := c.Params("name")
	g, err := h.g.GetByName(c.Context(), groupName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("group not found")
		}

		return h.InternalError(c, err, "failed to find group", zap.String("group_name", groupName))
	}

	members, err := h.g.ListMembersByID(c.Context(), g.ID, domain.GroupMemberAll, 10, 0)
	if err != nil {
		return h.InternalError(c, err, "failed to list recent members", log.GroupID(g.ID))
	}

	userIDs := make([]model.UserID, len(members))
	for i, member := range members {
		userIDs[i] = member.UserID
	}
	userMap, err := h.u.GetByIDs(c.Context(), userIDs...)
	if err != nil {
		return h.InternalError(c, err, "failed to get recent member user info")
	}

	return res.JSON(c, res.PrivateGroupProfile{
		ID:           g.ID,
		CreatedAt:    g.CreatedAt,
		Name:         g.Name,
		Title:        g.Title,
		Description:  g.Description,
		Icon:         groupIconPrefix + g.Icon,
		TotalMembers: g.MemberCount,
		NewMembers:   convertGroupMembers(members, userMap),
	})
}

func convertGroupMembers(members []model.GroupMember, userMap map[model.UserID]model.User) []res.PrivateGroupMember {
	s := make([]res.PrivateGroupMember, len(members))
	for i, member := range members {
		u := userMap[member.UserID]
		s[i] = res.PrivateGroupMember{
			ID:       member.UserID,
			UserName: u.UserName,
			NickName: u.NickName,
			Avatar:   res.UserAvatar(u.Avatar),
		}
	}

	return s
}
