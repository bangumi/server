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
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/ctrl"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) GetEpisode(c *fiber.Ctx) error {
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

		return h.InternalError(c, err, "failed to get episode", log.EpisodeID(id))
	}

	_, err = h.ctrl.GetSubject(c.Context(), u.Auth, e.SubjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to find subject of episode", log.SubjectID(e.SubjectID))
	}

	return res.JSON(c, res.ConvertModelEpisode(e))
}

func (h Handler) ListEpisode(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)

	page, err := req.GetPageQuery(c, req.EpisodeDefaultLimit, req.EpisodeMaxLimit)
	if err != nil {
		return err
	}

	epType, err := req.ParseEpTypeOptional(c.Query("type"))
	if err != nil {
		return err
	}

	subjectID, err := req.ParseSubjectID(c.Query("subject_id"))
	if err != nil {
		return err
	}
	if subjectID == 0 {
		return res.BadRequest("missing required query `subject_id`")
	}

	_, err = h.ctrl.GetSubject(c.Context(), u.Auth, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to get subject")
	}

	episodes, count, err := h.ctrl.ListEpisode(c.Context(), subjectID, epType, page.Limit, page.Offset)
	if err != nil {
		if errors.Is(err, ctrl.ErrOffsetTooBig) {
			return res.BadRequest("offset should be less than or equal to " + strconv.FormatInt(count, 10))
		}
		return h.InternalError(c, err, "failed to list episode")
	}

	var data = make([]res.Episode, len(episodes))
	for i, episode := range episodes {
		data[i] = res.ConvertModelEpisode(episode)
	}

	return c.JSON(res.PagedG[res.Episode]{
		Limit:  page.Limit,
		Offset: page.Offset,
		Data:   slice.Map(episodes, res.ConvertModelEpisode),
		Total:  count,
	})
}
