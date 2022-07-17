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

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) getTopic(c *fiber.Ctx, topicType domain.TopicType, id model.TopicID) (model.Topic, error) {
	u := h.getHTTPAccessor(c)
	topic, err := h.topic.Get(c.Context(), topicType, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return model.Topic{}, res.ErrNotFound
		}
		return model.Topic{}, errgo.Wrap(err, "Topic.Get")
	}

	if !auth.CanViewTopicContent(u.Auth, topic) {
		return model.Topic{}, res.ErrNotFound
	}

	return topic, nil
}

func (h Handler) listTopics(c *fiber.Ctx, topicType domain.TopicType, id uint32) error {
	u := h.getHTTPAccessor(c)
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return err
	}
	var response = res.Paged{
		Data:   res.EmptySlice(),
		Limit:  page.Limit,
		Offset: page.Offset,
	}

	statuses := auth.TopicStatuses(u.Auth)

	count, err := h.topic.Count(c.Context(), topicType, id, statuses)
	if err != nil {
		return errgo.Wrap(err, "repo.topic.Count")
	}

	if count == 0 {
		return res.JSON(c, response)
	}

	if err = page.check(count); err != nil {
		return err
	}

	topics, err := h.topic.List(c.Context(), topicType, id, statuses, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "repo.topic.GetTopics")
	}

	userIDs := slice.Map(topics, func(item model.Topic) model.UserID {
		return item.CreatorID
	})
	userMap, err := h.u.GetByIDs(c.Context(), dedupeUIDs(userIDs...)...)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	response.Total = count
	var data = make([]res.Topic, len(topics))
	for i, topic := range topics {
		data[i] = res.Topic{
			ID:        topic.ID,
			Title:     topic.Title,
			CreatedAt: topic.CreatedAt,
			UpdatedAt: topic.UpdatedAt,
			Creator:   convertModelUser(userMap[topic.CreatorID]),
			Replies:   topic.Replies,
		}
	}
	response.Data = data
	return res.JSON(c, response)
}

var errUnknownTopicType = errors.New("unknown topic type")

func (h Handler) getResTopicWithComments(c *fiber.Ctx, topicType domain.TopicType, topic model.Topic) error {
	var commentType domain.CommentType
	switch topicType {
	case domain.TopicTypeGroup:
		commentType = domain.CommentTypeGroupTopic
	case domain.TopicTypeSubject:
		commentType = domain.CommentTypeSubjectTopic
	default:
		return errUnknownTopicType
	}

	content, err := h.topic.GetTopicContent(c.Context(), topicType, topic.ID)
	if err != nil {
		return h.InternalError(c, err, "failed to get topic content")
	}

	pagedComments, err := h.listComments(c, commentType, topic.ID)
	if err != nil {
		return err
	}

	u, err := h.u.GetByIDs(c.Context(), topic.CreatorID)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	response := res.TopicDetail{
		ID:        topic.ID,
		Title:     topic.Title,
		CreatedAt: topic.CreatedAt,
		UpdatedAt: topic.UpdatedAt,
		Creator:   convertModelUser(u[topic.CreatorID]),
		Replies:   topic.Replies,
		Comments:  pagedComments,
		Text:      content.Content,
	}
	return c.JSON(response)
}
