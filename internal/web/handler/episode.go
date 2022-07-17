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
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/app/query"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/strparse"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) GetEpisode(c *fiber.Ctx) error {
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

	return c.JSON(e)
}

// first try to read from cache, then fallback to reading from database.
// return data, database record existence and error.
func (h Handler) getEpisodeWithCache(ctx context.Context, id model.EpisodeID) (res.Episode, bool, error) {
	e, err := h.app.Query.GetEpisode(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.Episode{}, false, nil
		}

		return res.Episode{}, false, errgo.Wrap(err, "app.Query.GetEpisode")
	}

	return convertModelEpisode(e), true, nil
}

func convertModelEpisode(s model.Episode) res.Episode {
	return res.Episode{
		ID:          s.ID,
		Name:        s.Name,
		NameCN:      s.NameCN,
		Ep:          s.Ep,
		Sort:        s.Sort,
		Duration:    s.Duration,
		Airdate:     s.Airdate,
		SubjectID:   s.SubjectID,
		Description: s.Description,
		Comment:     s.Comment,
		Type:        s.Type,
		Disc:        s.Disc,
	}
}

const episodeDefaultLimit = 100
const episodeMaxLimit = 200

func (h Handler) ListEpisode(c *fiber.Ctx) error {
	u := h.getHTTPAccessor(c)

	page, err := getPageQuery(c, episodeDefaultLimit, episodeMaxLimit)
	if err != nil {
		return err
	}

	epType, err := parseEpTypeOptional(c.Query("type"))
	if err != nil {
		return err
	}

	subjectID, err := parseSubjectID(c.Query("subject_id"))
	if err != nil {
		return err
	}
	if subjectID == 0 {
		return res.BadRequest("missing required query `subject_id`")
	}

	_, err = h.app.Query.GetSubject(c.Context(), u.Auth, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to get subject")
	}

	episodes, count, err := h.app.Query.ListEpisode(c.Context(), subjectID, epType, page.Limit, page.Offset)
	if err != nil {
		if errors.Is(err, query.ErrOffsetTooBig) {
			return res.BadRequest("offset should be less than or equal to " + strconv.FormatInt(count, 10))
		}
		return h.InternalError(c, err, "failed to list episode")
	}

	var data = make([]res.Episode, len(episodes))
	for i, episode := range episodes {
		data[i] = convertModelEpisode(episode)
	}

	return c.JSON(res.PagedG[model.Episode]{
		Limit:  page.Limit,
		Offset: page.Offset,
		Data:   episodes,
		Total:  count,
	})
}

func parseEpTypeOptional(s string) (*model.EpType, error) {
	if s == "" {
		return nil, nil //nolint:nilnil
	}

	v, err := strparse.Uint8(s)
	if err != nil {
		return nil, res.BadRequest("wrong value for query `type`")
	}

	switch v {
	case model.EpTypeNormal, model.EpTypeSpecial,
		model.EpTypeOpening, model.EpTypeEnding,
		model.EpTypeMad, model.EpTypeOther:
		return &v, nil
	}

	return nil, res.BadRequest(strconv.Quote(s) + " is not valid episode type")
}
