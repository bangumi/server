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

func (h Handler) listComments(c *fiber.Ctx, commentType domain.CommentType, id uint32) error {
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	comments, err := h.m.GetComments(c.Context(), commentType, id, page.Limit, page.Offset)
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
	for _, v := range comments {
		userIDs = append(userIDs, v.UID)
	}

	userMap, err := h.u.GetByIDs(c.Context(), dedupeUIDs(userIDs...)...)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}
	comments = model.ConvertModelCommentsToTree(comments, 0)
	count, err := h.m.Count(c.Context(), commentType, id)
	if err != nil {
		return errgo.Wrap(err, "repo.comments.Count")
	}

	return c.JSON(res.Paged{
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
		Data:   convertModelTopicComments(comments, userMap),
	})
}
