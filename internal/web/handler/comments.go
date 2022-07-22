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
	"github.com/bangumi/server/internal/pkg/generic/gmap"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) GetSubjectTopic(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)

	topicID, err := req.ParseTopicID(c.Params("topic_id"))
	if err != nil {
		return err
	}

	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	topic, err := h.getTopic(c, domain.TopicTypeSubject, topicID)
	if err != nil {
		return err
	}

	subjectID, err := getExpectSubjectID(c, topic)
	if err != nil {
		return err
	}

	_, err = h.app.Query.GetSubjectNoRedirect(c.Context(), u.Auth, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to subject", log.SubjectID(subjectID))
	}

	return h.getResTopicWithComments(c, domain.TopicTypeSubject, topic, page)
}

func (h Handler) ListSubjectTopics(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)

	id, err := req.ParseSubjectID(c.Params("id"))
	if err != nil || id == 0 {
		return res.BadRequest(err.Error())
	}

	_, err = h.app.Query.GetSubjectNoRedirect(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to subject", log.SubjectID(id))
	}

	return h.listTopics(c, domain.TopicTypeSubject, uint32(id))
}

func (h Handler) GetEpisodeComments(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)

	id, err := req.ParseEpisodeID(c.Params("id"))
	if err != nil {
		return err
	}

	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	e, err := h.app.Query.GetEpisode(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get episode", log.EpisodeID(id))
	}

	_, err = h.app.Query.GetSubjectNoRedirect(c.Context(), u.Auth, e.SubjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get subject of episode", log.SubjectID(e.SubjectID))
	}

	pagedComments, err := h.listComments(c, u.Auth, domain.CommentEpisode, model.TopicID(id), page)
	if err != nil {
		return h.InternalError(c, err, "failed to get comments", log.SubjectID(e.SubjectID))
	}

	return c.JSON(pagedComments)
}

func (h Handler) GetPersonComments(c *fiber.Ctx) error {
	id, err := req.ParsePersonID(c.Params("id"))
	if err != nil {
		return err
	}
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	r, err := h.app.Query.GetPerson(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get person", log.PersonID(id))
	}

	if r.Redirect != 0 {
		return res.ErrNotFound
	}

	u := h.GetHTTPAccessor(c)
	pagedComments, err := h.listComments(c, u.Auth, domain.CommentPerson, model.TopicID(id), page)
	if err != nil {
		return err
	}
	return c.JSON(pagedComments)
}

func (h Handler) GetCharacterComments(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)
	id, err := req.ParseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	_, err = h.app.Query.GetCharacterNoRedirect(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get character", log.CharacterID(id))
	}

	pagedComments, err := h.listComments(c, u.Auth, domain.CommentCharacter, model.TopicID(id), page)
	if err != nil {
		return err
	}
	return res.JSON(c, pagedComments)
}

func (h Handler) GetIndexComments(c *fiber.Ctx) error {
	user := h.GetHTTPAccessor(c)

	id, err := req.ParseIndexID(c.Params("id"))
	if err != nil {
		return err
	}

	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
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
	pagedComments, err := h.listComments(c, u.Auth, domain.CommentIndex, model.TopicID(id), page)
	if err != nil {
		return err
	}
	return c.JSON(pagedComments)
}

func (h Handler) listComments(
	c *fiber.Ctx,
	u domain.Auth,
	commentType domain.CommentType,
	id model.TopicID,
	page req.PageQuery,
) (res.PagedComment, error) {
	count, err := h.topic.CountReplies(c.Context(), commentType, id)
	if err != nil {
		return res.PagedComment{}, errgo.Wrap(err, "repo.comments.Count")
	}

	if count == 0 {
		return res.PagedComment{Data: []res.PrivateComment{}, Total: count, Limit: page.Limit, Offset: page.Offset}, nil
	}

	if err = page.Check(count); err != nil {
		return res.PagedComment{}, err
	}

	comments, err := h.topic.ListReplies(c.Context(), commentType, id, page.Limit, page.Offset)
	if err != nil {
		return res.PagedComment{}, errgo.Wrap(err, "topic.ListReplies")
	}

	userMap, err := h.app.Query.GetUsersByIDs(c.Context(), commentsToUserIDs(comments)...)
	if err != nil {
		return res.PagedComment{}, errgo.Wrap(err, "user.GetByIDs")
	}

	var friends map[model.UserID]domain.FriendItem
	if u.ID != 0 {
		friends, err = h.u.GetFriends(c.Context(), u.ID)
		if err != nil {
			return res.PagedComment{}, errgo.Wrap(err, "failed to get friends")
		}
	}

	return res.PagedComment{
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
		Data:   convertModelComments(comments, userMap, friends),
	}, nil
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
				Friend:    ok,
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
			Friend:    ok,
			CreatedAt: comment.CreatedAt,
			Creator:   res.ConvertModelUser(userMap[comment.CreatorID]),
			Replies:   replies,
			State:     res.ToCommentState(comment.State),
		}
	}

	return result
}
