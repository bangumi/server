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

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/gmap"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) listComments(
	c *fiber.Ctx, commentType domain.CommentType, id model.TopicID,
) (res.PagedComment, error) {
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return res.PagedComment{}, res.ErrNotFound
	}

	count, err := h.topic.CountReplies(c.Context(), commentType, id)
	if err != nil {
		return res.PagedComment{}, errgo.Wrap(err, "repo.comments.Count")
	}

	if count == 0 {
		return res.PagedComment{}, res.JSON(
			c, res.PagedComment{Data: []res.Comment{}, Total: count, Limit: page.Limit, Offset: page.Offset},
		)
	}

	if err = page.check(count); err != nil {
		return res.PagedComment{}, err
	}

	comments, err := h.topic.ListReplies(c.Context(), commentType, id, page.Limit, page.Offset)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.PagedComment{}, res.ErrNotFound
		}
		return res.PagedG[res.Comment]{}, errgo.Wrap(err, "Comment.GetCommentsByMentionedID")
	}

	uidMap := make(map[model.UserID]bool, len(comments))
	for _, comment := range comments {
		uidMap[comment.CreatorID] = true
		for _, sub := range comment.SubComments {
			uidMap[sub.CreatorID] = true
		}
	}

	userMap, err := h.u.GetByIDs(c.Context(), gmap.Keys(uidMap)...)
	if err != nil {
		return res.PagedComment{}, errgo.Wrap(err, "user.GetByIDs")
	}

	return res.PagedComment{
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
		Data:   convertModelComments(comments, userMap),
	}, nil
}

func convertModelComments(
	comments []model.Comment, userMap map[model.UserID]model.User,
) []res.Comment {
	result := make([]res.Comment, len(comments))
	for k, comment := range comments {
		var replies = make([]res.SubComment, len(comment.SubComments))

		for i, subComment := range comment.SubComments {
			replies[i] = res.SubComment{
				CreatedAt: subComment.CreatedAt,
				Text:      subComment.Content,
				Creator:   convertModelUser(userMap[subComment.CreatorID]),
				ID:        subComment.ID,
			}
		}

		result[k] = res.Comment{
			ID:        comment.ID,
			Text:      comment.Content,
			CreatedAt: comment.CreatedAt,
			Creator:   convertModelUser(userMap[comment.CreatorID]),
			Replies:   replies,
		}
	}

	return result
}
