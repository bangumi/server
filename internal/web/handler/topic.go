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
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) getUserMapOfTopics(c *fiber.Ctx, topics ...model.Topic) (map[uint32]model.User, error) {
	userIDs := make([]model.UIDType, 0)
	for _, topic := range topics {
		userIDs = append(userIDs, topic.UID)
		for _, v := range topic.Comments.Data {
			userIDs = append(userIDs, v.UID)
		}
	}
	userMap, err := h.u.GetByIDs(c.Context(), dedupeUIDs(userIDs...)...)
	if err != nil {
		return nil, errgo.Wrap(err, "user.GetByIDs")
	}
	return userMap, nil
}

func (h Handler) listTopics(c *fiber.Ctx, topicType model.TopicType, id uint32) error {
	topics, err := h.t.GetTopicsByObjectID(c.Context(), topicType, id)
	if err != nil {
		return errgo.Wrap(err, "repo.topic.GetTopicsByObjectID")
	}

	userMap, err := h.getUserMapOfTopics(c, topics...)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	response := make([]res.Topic, 0)
	for _, v := range topics {
		creator := userMap[v.UID]
		response = append(response, res.Topic{
			ID:        v.ID,
			Title:     v.Title,
			CreatedAt: v.CreatedAt,
			Creator: res.Creator{
				Username: creator.UserName,
				Nickname: creator.NickName,
			},
			Replies:  v.Replies,
			Comments: &res.Comments{},
		})
	}
	return c.JSON(response)
}
