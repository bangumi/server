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
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/gmap"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) listTopics(c *fiber.Ctx, topicType domain.TopicType, id uint32) error {
	u := h.GetHTTPAccessor(c)
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}
	topics, count, err := h.ctrl.ListTopics(c.Context(), u.Auth, topicType, id, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "ctrl.ListTopics")
	}

	if count == 0 {
		return res.JSON(c, res.Paged{
			Data:   res.EmptySlice(),
			Total:  count,
			Limit:  page.Limit,
			Offset: page.Offset,
		})
	}

	if err = page.Check(count); err != nil {
		return err
	}

	userIDs := slice.Map(topics, func(item model.Topic) model.UserID {
		return item.CreatorID
	})
	userMap, err := h.ctrl.GetUsersByIDs(c.Context(), slice.UniqueUnsorted(userIDs)...)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	var data = make([]res.PrivateTopic, len(topics))
	for i, topic := range topics {
		data[i] = res.PrivateTopic{
			ID:         topic.ID,
			Title:      topic.Title,
			CreatedAt:  topic.CreatedAt,
			UpdatedAt:  topic.UpdatedAt,
			Creator:    res.ConvertModelUser(userMap[topic.CreatorID]),
			ReplyCount: topic.Replies,
		}
	}
	return res.JSON(c, res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}

func (h Handler) getResTopicWithComments(
	c *fiber.Ctx, topicType domain.TopicType, topicID model.TopicID,
) (*res.PrivateTopicDetail, error) {
	a := h.GetHTTPAccessor(c)

	t, err := h.ctrl.GetTopic(context.TODO(), a.Auth, topicType, topicID, 0, 0)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, res.ErrNotFound
		}

		return nil, errgo.Wrap(err, "failed to get topic")
	}

	var userIDs = make(map[model.UserID]struct{}, len(t.Replies))
	for _, reply := range t.Replies {
		userIDs[reply.CreatorID] = struct{}{}
		for _, comment := range reply.SubComments {
			userIDs[comment.CreatorID] = struct{}{}
		}
	}

	userIDs[t.CreatorID] = struct{}{}

	users, err := h.ctrl.GetUsersByIDs(context.TODO(), gmap.Keys(userIDs)...)
	if err != nil {
		return nil, errgo.Wrap(err, "ctrl.GetUsersByIDs")
	}

	var friends map[model.UserID]domain.FriendItem
	if a.Login {
		friends, err = h.ctrl.GetFriends(context.TODO(), a.ID)
		if err != nil {
			return nil, errgo.Wrap(err, "userRepo.GetFriends")
		}
	}

	return &res.PrivateTopicDetail{
		ParentID:  t.ParentID,
		ID:        t.ID,
		Title:     t.Title,
		IsFriend:  gmap.Has(friends, t.CreatorID),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		Creator:   res.ConvertModelUser(users[t.CreatorID]),
		State:     res.ToCommentState(t.State),
		Comments:  fromModelComments(t.Replies, users, friends),
		Text:      t.Content,
	}, nil
}

func fromModelComments(
	replies []model.Comment,
	users map[model.UserID]model.User,
	friends map[model.UserID]domain.FriendItem,
) []res.PrivateComment {
	var comments = make([]res.PrivateComment, 0, len(replies))
	for _, reply := range replies {
		var subComments = make([]res.PrivateSubComment, 0, len(reply.SubComments))
		for _, comment := range reply.SubComments {
			_, f := friends[comment.CreatorID]
			subComments = append(subComments, res.PrivateSubComment{
				CreatedAt: comment.CreatedAt,
				Text:      comment.Content,
				Creator:   res.ConvertModelUser(users[comment.CreatorID]),
				IsFriend:  f,
				State:     res.ToCommentState(comment.State),
				ID:        comment.ID,
			})
		}

		_, f := friends[reply.CreatorID]
		comments = append(comments, res.PrivateComment{
			CreatedAt: reply.CreatedAt,
			Text:      reply.Content,
			Creator:   res.ConvertModelUser(users[reply.CreatorID]),
			Replies:   subComments,
			ID:        reply.ID,
			IsFriend:  f,
			State:     res.ToCommentState(reply.State),
		})
	}

	return comments
}
