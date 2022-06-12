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
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/util"
)

const canViewStateClosedTopic = -time.Hour * 24 * 180
const canViewStateDeleteTopic = -time.Hour * 24 * 180

func (h Handler) getTopic(c *fiber.Ctx, topicType domain.TopicType, id model.TopicIDType) (model.Topic, error) {
	u := h.getHTTPAccessor(c)
	topic, err := h.t.Get(c.Context(), topicType, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return model.Topic{}, c.Status(http.StatusNotFound).JSON(res.Error{
				Title:   "Not Found",
				Details: util.DetailFromRequest(c),
			})
		}
		return model.Topic{}, errgo.Wrap(err, "Topic.Get")
	}

	switch {
	case !u.Permission.ManageTopicState && topic.Status == model.TopicStatusReview,
		topic.State == model.TopicStateClosed && !u.RegisteredTime(canViewStateClosedTopic),
		topic.State == model.TopicStateDelete && !u.RegisteredTime(canViewStateDeleteTopic):
		return model.Topic{}, c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	return topic, nil
}

func (h Handler) getUserMapOfTopics(c *fiber.Ctx, topics ...model.Topic) (map[uint32]model.User, error) {
	userIDs := make([]model.UIDType, 0)
	for _, topic := range topics {
		userIDs = append(userIDs, topic.UID)
		for _, v := range topic.Comments {
			userIDs = append(userIDs, v.UID)
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
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}
	var response = res.Paged{
		Limit:  page.Limit,
		Offset: page.Offset,
	}

	statuses := []model.TopicStatus{model.TopicStatusNormal}
	if u.Permission.ManageTopicState {
		statuses = append(statuses, model.TopicStatusReview)
	}

	topics, err := h.t.List(c.Context(), topicType, id, statuses, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "repo.topic.GetTopics")
	}

	userMap, err := h.getUserMapOfTopics(c, topics...)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	count, err := h.t.Count(c.Context(), topicType, id, statuses)
	if err != nil {
		return errgo.Wrap(err, "repo.topic.Count")
	}

	response.Total = count
	var data = make([]res.Topic, len(topics))
	for i, topic := range topics {
		data[i] = res.Topic{
			ID:        topic.ID,
			Title:     topic.Title,
			CreatedAt: topic.CreatedAt,
			UpdatedAt: topic.UpdatedAt,
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
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return err
	}

	userMap, err := h.getUserMapOfTopics(c, topic)
	if err != nil {
		return err
	}

	commentType := map[domain.TopicType]domain.CommentType{
		domain.TopicTypeGroup:   domain.CommentTypeGroupTopic,
		domain.TopicTypeSubject: domain.CommentTypeSubjectTopic,
	}[topicType]

	comments, err := h.m.List(c.Context(), commentType, topic.ID, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "repo.comments.GetComments")
	}

	count, err := h.m.Count(c.Context(), commentType, topic.ID)
	if err != nil {
		return errgo.Wrap(err, "repo.comments.Count")
	}

	comments = model.ConvertModelCommentsToTree(comments, 0)
	response := res.Topic{
		ID:        topic.ID,
		Title:     topic.Title,
		CreatedAt: topic.CreatedAt,
		UpdatedAt: topic.UpdatedAt,
		Creator:   convertModelUser(userMap[topic.UID]),
		Replies:   topic.Replies,
		Comments: &res.Paged{
			Total:  count,
			Limit:  page.Limit,
			Offset: page.Offset,
			Data:   convertModelTopicComments(comments, userMap),
		},
	}
	return c.JSON(response)
}

func convertModelTopicComments(comments []model.Comment, userMap map[uint32]model.User) []res.Comment {
	replies := make([]res.Comment, 0)
	for _, v := range comments {
		replies = append(replies, res.Comment{
			ID:        v.ID,
			Text:      v.Content,
			CreatedAt: v.CreatedAt,
			Creator:   convertModelUser(userMap[v.UID]),
			Replies:   convertModelTopicComments(v.Replies, userMap),
		})
	}
	return replies
}
