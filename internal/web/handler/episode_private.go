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
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) GetEpisodeComments(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)

	id, err := req.ParseEpisodeID(c.Params("id"))
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

	pagedComments, err := h.listComments(c, domain.CommentEpisode, model.TopicID(id))
	if err != nil {
		return h.InternalError(c, err, "failed to get comments", log.SubjectID(e.SubjectID))
	}

	return c.JSON(pagedComments)
}
