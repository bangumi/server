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
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) GetEpisodeComments(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)

	id, err := parseEpisodeID(c.Params("id"))
	if err != nil {
		return err
	}

	e, ok, err := h.getEpisodeWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok {
		return res.ErrNotFound
	}

	s, ok, err := h.getSubjectWithCache(c.Context(), e.SubjectID)
	if err != nil {
		return err
	}

	if !ok || s.Redirect != 0 || (s.NSFW && !u.AllowNSFW()) {
		return res.ErrNotFound
	}

	pagedComments, err := h.listComments(c, domain.CommentEpisode, model.TopicID(id))
	if err != nil {
		return err
	}
	return c.JSON(pagedComments)
}
