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

	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/gmap"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

type ResPrivateTopicDetailWithGroup struct {
	*res.PrivateTopicDetail
	Group res.PrivateGroup `json:"group"`
}

func (h Handler) GetGroupTopic(c *fiber.Ctx) error {
	topicID, err := req.ParseTopicID(c.Params("topic_id"))
	if err != nil {
		return err
	}

	data, err := h.getResTopicWithComments(c, domain.TopicTypeGroup, topicID)
	if err != nil {
		return err
	}

	group, err := h.g.GetByID(context.TODO(), model.GroupID(data.ParentID))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return errgo.Wrap(err, "failed to get group")
	}

	return res.JSON(c, ResPrivateTopicDetailWithGroup{
		PrivateTopicDetail: data,
		Group: res.PrivateGroup{
			ID:        group.ID,
			Name:      group.Name,
			CreatedAt: group.CreatedAt,
			Title:     group.Title,
			Icon:      groupIconPrefix + group.Icon,
		},
	})
}

func (h Handler) GetSubjectTopic(c *fiber.Ctx) error {
	topicID, err := req.ParseTopicID(c.Params("topic_id"))
	if err != nil {
		return err
	}

	data, err := h.getResTopicWithComments(c, domain.TopicTypeSubject, topicID)
	if err != nil {
		return err
	}

	return res.JSON(c, data)
}

func (h Handler) ListSubjectTopics(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)

	id, err := req.ParseSubjectID(c.Params("id"))
	if err != nil || id == 0 {
		return res.BadRequest(err.Error())
	}

	_, err = h.ctrl.GetSubjectNoRedirect(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return errgo.Wrap(err, "failed to subject")
	}

	return h.listTopics(c, domain.TopicTypeSubject, uint32(id))
}

func (h Handler) GetEpisodeComments(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)

	id, err := req.ParseEpisodeID(c.Params("id"))
	if err != nil {
		return err
	}

	e, err := h.ctrl.GetEpisode(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get episode")
	}

	_, err = h.ctrl.GetSubjectNoRedirect(c.Context(), u.Auth, e.SubjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get subject of episode")
	}

	return h.listComments(c, u.Auth, domain.CommentEpisode, model.TopicID(id))
}

func (h Handler) GetPersonComments(c *fiber.Ctx) error {
	id, err := req.ParsePersonID(c.Params("id"))
	if err != nil {
		return err
	}

	r, err := h.ctrl.GetPerson(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get person")
	}

	if r.Redirect != 0 {
		return res.ErrNotFound
	}

	u := h.GetHTTPAccessor(c)
	return h.listComments(c, u.Auth, domain.CommentPerson, model.TopicID(id))
}

func (h Handler) GetCharacterComments(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)
	id, err := req.ParseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	_, err = h.ctrl.GetCharacterNoRedirect(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get character")
	}

	return h.listComments(c, u.Auth, domain.CommentCharacter, model.TopicID(id))
}

func (h Handler) GetIndexComments(c *fiber.Ctx) error {
	user := h.GetHTTPAccessor(c)

	id, err := req.ParseIndexID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getIndexWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.NSFW && !user.AllowNSFW() {
		return res.ErrNotFound
	}

	u := h.GetHTTPAccessor(c)
	return h.listComments(c, u.Auth, domain.CommentIndex, model.TopicID(id))
}

func (h Handler) listComments(
	c *fiber.Ctx,
	u domain.Auth,
	commentType domain.CommentType,
	id model.TopicID,
) error {
	// a noop limit to fetch all comments
	comments, _, err := h.ctrl.ListReplies(c.Context(), commentType, id, 100000, 0) //nolint:gomnd
	if err != nil {
		return errgo.Wrap(err, "topic.ListReplies")
	}

	userMap, err := h.ctrl.GetUsersByIDs(c.Context(), commentsToUserIDs(comments))
	if err != nil {
		return errgo.Wrap(err, "query.GetUsersByIDs")
	}

	var friends map[model.UserID]domain.FriendItem
	if u.ID != 0 {
		friends, err = h.ctrl.GetFriends(c.Context(), u.ID)
		if err != nil {
			return errgo.Wrap(err, "userRepo.GetFriends")
		}
	}

	return res.JSON(c, res.PrivateComments{Comments: convertModelComments(comments, userMap, friends)})
}

func commentsToUserIDs(comments []model.Comment) []model.UserID {
	uidMap := make(map[model.UserID]struct{}, len(comments))
	for _, comment := range comments {
		uidMap[comment.CreatorID] = struct{}{}
		for _, sub := range comment.SubComments {
			uidMap[sub.CreatorID] = struct{}{}
		}
	}

	return gmap.Keys(uidMap)
}

func convertModelComments(
	comments []model.Comment,
	userMap map[model.UserID]model.User,
	friends map[model.UserID]domain.FriendItem,
) []res.PrivateComment {
	result := make([]res.PrivateComment, len(comments))
	for k, comment := range comments {
		var replies = make([]res.PrivateSubComment, len(comment.SubComments))

		for i, subComment := range comment.SubComments {
			subComment = auth.RewriteSubCommit(subComment)
			_, ok := friends[subComment.CreatorID]

			replies[i] = res.PrivateSubComment{
				IsFriend:  ok,
				CreatedAt: subComment.CreatedAt,
				Text:      subComment.Content,
				Creator:   res.ConvertModelUser(userMap[subComment.CreatorID]),
				ID:        subComment.ID,
				State:     res.ToCommentState(subComment.State),
			}
		}

		_, ok := friends[comment.CreatorID]

		comment = auth.RewriteCommit(comment)
		result[k] = res.PrivateComment{
			ID:        comment.ID,
			Text:      comment.Content,
			IsFriend:  ok,
			CreatedAt: comment.CreatedAt,
			Creator:   res.ConvertModelUser(userMap[comment.CreatorID]),
			Replies:   replies,
			State:     res.ToCommentState(comment.State),
		}
	}

	return result
}
