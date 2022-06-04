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

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/util"
)

func (h Handler) listComments(c *fiber.Ctx, commentType model.CommentType, id uint32) error {
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	comments, err := h.m.GetCommentsByMentionedID(c.Context(), commentType, page.Limit, page.Offset, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.Status(http.StatusNotFound).JSON(res.Error{
				Title:   "Not Found",
				Details: util.DetailFromRequest(c),
			})
		}
		return errgo.Wrap(err, "Comment.GetCommentsByMentionedID")
	}

	userIDs := make([]model.UIDType, 0)
	for _, v := range comments.Data {
		userIDs = append(userIDs, v.UID)
	}

	userMap, err := h.u.GetByIDs(c.Context(), dedupeUIDs(userIDs...)...)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}
	comments.Data = model.ConvertModelCommentsToTree(comments.Data, 0)

	return c.JSON(res.Comments{
		Total:  comments.Total,
		Limit:  comments.Limit,
		Offset: comments.Offset,
		Data:   convertModelTopicComments(comments.Data, userMap),
	})
}
