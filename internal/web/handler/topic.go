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
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/web/res"
)

const canViewStateClosedTopic = -time.Hour * 24 * 180
const canViewStateDeleteTopic = -time.Hour * 24 * 180

func (h Handler) getTopic(c *fiber.Ctx, topicType domain.TopicType, id model.TopicID) (model.Topic, error) {
	u := h.getHTTPAccessor(c)
	topic, err := h.topic.Get(c.Context(), topicType, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return model.Topic{}, res.ErrNotFound
		}
		return model.Topic{}, errgo.Wrap(err, "Topic.Get")
	}

	switch {
	case !u.Permission.ManageTopicState && topic.Status == model.TopicStatusReview,
		topic.State == model.TopicStateClosed && !u.RegisteredTime(canViewStateClosedTopic),
		topic.State == model.TopicStateDelete && !u.RegisteredTime(canViewStateDeleteTopic):
		return model.Topic{}, res.ErrNotFound
	}

	return topic, nil
}

func (h Handler) getUserMapOfTopics(c *fiber.Ctx, topics ...model.Topic) (map[model.UserID]model.User, error) {
	userIDs := make([]model.UserID, 0)
	for _, topic := range topics {
		userIDs = append(userIDs, topic.UID)
		for _, v := range topic.Comments {
			userIDs = append(userIDs, v.CreatorID)
		}
	}
	userMap, err := h.u.GetByIDs(c.Context(), dedupeUIDs(userIDs...)...)
	if err != nil {
		return nil, errgo.Wrap(err, "user.GetByIDs")
	}
	return userMap, nil
}

func (h Handler) listTopics(c *fiber.Ctx, topicType domain.TopicType, id uint32) error {
	u := h.getHTTPAccessor(c)
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return res.ErrNotFound
	}
	var response = res.Paged{
		Limit:  page.Limit,
		Offset: page.Offset,
	}

	statuses := []model.TopicStatus{model.TopicStatusNormal}
	if u.Permission.ManageTopicState {
		statuses = append(statuses, model.TopicStatusReview)
	}

	topics, err := h.topic.List(c.Context(), topicType, id, statuses, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "repo.topic.GetTopics")
	}

	userMap, err := h.getUserMapOfTopics(c, topics...)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	count, err := h.topic.Count(c.Context(), topicType, id, statuses)
	if err != nil {
		return errgo.Wrap(err, "repo.topic.Count")
	}

	response.Total = count
	var data = make([]res.Topic, len(topics))
	for i, topic := range topics {
		data[i] = res.Topic{
			ID:        topic.ID,
			Title:     topic.Title,
			CreatedAt: topic.CreatedTime,
			UpdatedAt: topic.UpdatedTime,
			Creator:   convertModelUser(userMap[topic.UID]),
			Replies:   topic.Replies,
		}
	}
	response.Data = data
	return c.JSON(response)
}

func (h Handler) getResTopicWithComments(
	c *fiber.Ctx, topicType domain.TopicType, topic model.Topic,
) error {
	commentType := map[domain.TopicType]domain.CommentType{
		domain.TopicTypeGroup:   domain.CommentTypeGroupTopic,
		domain.TopicTypeSubject: domain.CommentTypeSubjectTopic,
	}[topicType]

	pagedComments, err := h.listComments(c, commentType, topic.ID)
	if err != nil {
		return err
	}

	u, err := h.u.GetByIDs(c.Context(), topic.UID)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	response := res.Topic{
		ID:        topic.ID,
		Title:     topic.Title,
		CreatedAt: topic.CreatedTime,
		UpdatedAt: topic.UpdatedTime,
		Creator:   convertModelUser(u[topic.UID]),
		Replies:   topic.Replies,
		Comments:  pagedComments,
	}
	return c.JSON(response)
}
