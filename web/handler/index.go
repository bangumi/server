// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/compat"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/pkg/wiki"
	"github.com/bangumi/server/web/handler/cachekey"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/util"
)

func (h Handler) getIndexWithCache(c context.Context, id uint32) (res.Index, bool, error) {
	var key = cachekey.Index(id)

	var r res.Index
	ok, err := h.cache.Get(c, key, &r)
	if err != nil {
		return r, ok, errgo.Wrap(err, "cache.Get")
	}

	i, err := h.i.Get(c, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.Index{}, false, nil
		}

		return res.Index{}, false, errgo.Wrap(err, "Index.Get")
	}

	u, err := h.u.GetByID(c, i.CreatorID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.log.Error("index missing creator",
				zap.Uint32("index_id", id), zap.Uint32("user_id", i.CreatorID))
			return res.Index{}, false, fiber.ErrInternalServerError
		}

		return res.Index{}, false, errgo.Wrap(err, "Index.Get")
	}

	r = res.Index{
		CreatedAt: i.CreatedAt,
		Creator: res.Creator{
			Username: u.UserName,
			Nickname: u.NickName,
		},
		Title:       i.Title,
		Description: i.Description,
		Total:       i.Total,
		ID:          id,
		Stat: res.Stat{
			Comments: i.Comments,
			Collects: i.Collects,
		},
		Ban:  i.Ban,
		NSFW: i.NSFW,
	}

	if e := h.cache.Set(c, key, r, time.Hour); e != nil {
		h.log.Error("can't set response to cache", zap.Error(e))
	}

	return r, true, nil
}

func (h Handler) GetIndex(c *fiber.Ctx) error {
	user := h.getHTTPAccessor(c)

	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id == 0 {
		return fiber.NewError(http.StatusBadRequest, "bad index_id")
	}

	r, ok, err := h.getIndexWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.NSFW && !user.AllowNSFW() {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	return c.JSON(r)
}
func (h Handler) GetIndexSubjects(c *fiber.Ctx) error {
	user := h.getHTTPAccessor(c)

	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id == 0 {
		return fiber.NewError(http.StatusBadRequest, "bad index_id")
	}

	subjectType, err := parseSubjectType(c.Query("type"))
	if err != nil {
		return errgo.Wrap(err, "invalid query `type` for subject type")
	}

	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return err
	}

	r, ok, err := h.getIndexWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || (r.NSFW && !user.AllowNSFW()) {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	return h.getIndexSubjects(c, id, subjectType, page)
}

func (h Handler) getIndexSubjects(
	c *fiber.Ctx, id uint32, subjectType uint8, page pageQuery,
) error {
	count, err := h.i.CountSubjects(c.Context(), id, subjectType)
	if err != nil {
		return errgo.Wrap(err, "Index.CountSubjects")
	}

	if count == 0 {
		return c.JSON(res.Paged{
			Data:   []int{},
			Total:  0,
			Limit:  page.Limit,
			Offset: page.Offset,
		})
	}

	if err = page.check(count); err != nil {
		return err
	}

	subjects, err := h.i.ListSubjects(c.Context(), id, subjectType, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "Index.ListSubjects")
	}

	var data = make([]res.SlimSubjectV0, len(subjects))
	for i, s := range subjects {
		data[i] = res.SlimSubjectV0{
			AddedAt: s.AddedAt,
			Date:    nilString(s.Subject.Date),
			Image:   model.SubjectImage(s.Subject.Image),
			Name:    s.Subject.Name,
			NameCN:  s.Subject.NameCN,
			Comment: s.Comment,
			Infobox: compat.V0Wiki(wiki.ParseOmitError(s.Subject.Infobox).NonZero()),
			ID:      s.Subject.ID,
			TypeID:  s.Subject.TypeID,
		}
	}

	return c.JSON(res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}
