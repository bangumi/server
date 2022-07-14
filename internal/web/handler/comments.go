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
	"github.com/bangumi/server/internal/web/res"
)

//nolint:gocyclo
func (h Handler) listComments(c *fiber.Ctx, commentType domain.CommentType, id model.TopicID) (*res.Paged, error) {
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return nil, res.ErrNotFound
	}

	count, err := h.comment.Count(c.Context(), commentType, id)
	if err != nil {
		return nil, errgo.Wrap(err, "repo.comments.Count")
	}

	if count == 0 {
		return nil, c.JSON(res.Paged{Data: []res.Comment{}, Total: count, Limit: page.Limit, Offset: page.Offset})
	}

	if err = page.check(count); err != nil {
		return nil, err
	}

	comments, err := h.comment.List(c.Context(), commentType, id, page.Limit, page.Offset)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, res.ErrNotFound
		}
		return nil, errgo.Wrap(err, "Comment.GetCommentsByMentionedID")
	}

	uids := make([]model.UserID, 0)

	extIDs := make([]model.CommentID, 0)
	for _, v := range comments {
		uids = append(uids, v.CreatorID)
		extIDs = append(extIDs, v.ID)
	}

	var relatedComments map[model.CommentID][]model.Comment
	if len(extIDs) != 0 {
		relatedComments, err = h.comment.GetByRelateIDs(c.Context(), commentType, extIDs...)
		if err != nil {
			return nil, errgo.Wrap(err, "repo.comments.GetByIDs")
		}
		for _, v := range relatedComments {
			for _, vv := range v {
				uids = append(uids, vv.CreatorID)
			}
		}
	}

	userMap, err := h.u.GetByIDs(c.Context(), dedupeUIDs(uids...)...)
	if err != nil {
		return nil, errgo.Wrap(err, "user.GetByIDs")
	}

	return &res.Paged{
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
		Data:   convertModelComments(comments, relatedComments, userMap),
	}, nil
}

func convertModelComments(
	comments []model.Comment, cm map[model.CommentID][]model.Comment, userMap map[model.UserID]model.User,
) []res.Comment {
	result := make([]res.Comment, len(comments))
	for k, v := range comments {
		result[k] = res.Comment{
			ID:        v.ID,
			Text:      v.Content,
			CreatedAt: v.CreatedAt,
			Creator:   convertModelUser(userMap[v.CreatorID]),
		}
		if relates, ok := cm[v.ID]; ok {
			result[k].Replies = make([]res.Comment, len(relates))
			for i, related := range relates {
				result[k].Replies[i] = res.Comment{
					CreatedAt: related.CreatedAt,
					Creator:   convertModelUser(userMap[related.CreatorID]),
					Text:      related.Content,
					ID:        related.ID,
				}
			}
		}
	}
	return result
}
