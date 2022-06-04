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
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/cachekey"
	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/internal/web/res"
	"github.com/bangumi/server/internal/web/util"
	"github.com/bangumi/server/pkg/vars/enum"
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
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	s, ok, err := h.getSubjectWithCache(c.Context(), e.SubjectID)
	if err != nil {
		return err
	}
	if !ok || s.Redirect != 0 || (s.NSFW && !u.AllowNSFW()) {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	return c.JSON(e)
}

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
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	s, ok, err := h.getSubjectWithCache(c.Context(), e.SubjectID)
	if err != nil {
		return err
	}

	if !ok || s.Redirect != 0 || (s.NSFW && !u.AllowNSFW()) {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	return h.listComments(c, model.CommentEpisode, id)
}

// first try to read from cache, then fallback to reading from database.
// return data, database record existence and error.
func (h Handler) getEpisodeWithCache(ctx context.Context, id uint32) (res.Episode, bool, error) {
	var key = cachekey.Episode(id)
	// try to read from cache
	var r res.Episode
	ok, err := h.cache.Get(ctx, key, &r)
	if err != nil {
		return r, ok, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		return r, ok, nil
	}

	s, err := h.e.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.Episode{}, false, nil
		}

		return r, ok, errgo.Wrap(err, "repo.episode.Set")
	}

	r = convertModelEpisode(s)

	if e := h.cache.Set(ctx, key, r, time.Minute); e != nil {
		h.log.Error("can't set response to cache", zap.Error(e))
	}

	return r, true, nil
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

	epType, err := parseEpType(c.Query("type"))
	if err != nil {
		return err
	}

	subjectID, err := parseSubjectID(c.Query("subject_id"))
	if err != nil {
		return err
	}
	if subjectID == 0 {
		return fiber.NewError(http.StatusBadRequest, "missing required query `subject_id`")
	}

	subject, ok, err := h.getSubjectWithCache(c.Context(), subjectID)
	if err != nil {
		return err
	}

	if !ok || subject.Redirect != 0 || (subject.NSFW && !u.AllowNSFW()) {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	return h.listEpisode(c, subjectID, page, epType)
}

func (h Handler) listEpisode(
	c *fiber.Ctx,
	subjectID model.SubjectIDType,
	page pageQuery,
	epType enum.EpType,
) error {
	var response = res.Paged{
		Limit:  page.Limit,
		Offset: page.Offset,
	}

	var count int64
	var err error
	if epType < 0 {
		count, err = h.e.Count(c.Context(), subjectID)
		if err != nil {
			return errgo.Wrap(err, "episode.Count")
		}
	} else {
		count, err = h.e.CountByType(c.Context(), subjectID, epType)
		if err != nil {
			return errgo.Wrap(err, "episode.CountByType")
		}
	}

	if count == 0 {
		response.Data = []int{}
		return c.JSON(response)
	}

	if int64(page.Offset) >= count {
		return fiber.NewError(http.StatusBadRequest, "offset if greater than count")
	}

	response.Total = count

	var episodes []model.Episode
	if epType < 0 {
		episodes, err = h.e.List(c.Context(), subjectID, page.Limit, page.Offset)
		if err != nil {
			return errgo.Wrap(err, "episode.List")
		}
	} else {
		episodes, err = h.e.ListByType(c.Context(), subjectID, epType, page.Limit, page.Offset)
		if err != nil {
			return errgo.Wrap(err, "episode.ListByType")
		}
	}

	var data = make([]res.Episode, len(episodes))
	for i, episode := range episodes {
		data[i] = convertModelEpisode(episode)
	}
	response.Data = data

	return c.JSON(response)
}

func parseEpType(s string) (model.EpTypeType, error) {
	if s == "" {
		return -1, nil
	}

	v, err := strparse.Uint8(s)
	if err != nil {
		return -1, fiber.NewError(http.StatusBadRequest, "wrong value for query `type`")
	}

	e := model.EpTypeType(v)
	switch e {
	case enum.EpTypeNormal, enum.EpTypeSpecial,
		enum.EpTypeOpening, enum.EpTypeEnding,
		enum.EpTypeMad, enum.EpTypeOther:
		return e, nil
	}

	return 0, fiber.NewError(http.StatusBadRequest, strconv.Quote(s)+" is not valid episode type")
}
