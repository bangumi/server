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

package web

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/compat"
	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/pkg/vars"
	"github.com/bangumi/server/pkg/wiki"
	"github.com/bangumi/server/web/res"
)

func subjectCacheKey(id uint32) string {
	return "chii:res:0:repository:" + strconv.FormatUint(uint64(id), 10)
}

func (h Handler) getSubject(c *fiber.Ctx) error {
	u, ok := c.Context().UserValue(ctxKeyUser).(accessor) // get visitor
	if !ok {
		panic("can't convert type")
	}

	id, err := strparse.Uint32(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("bad id " + err.Error())
	}

	r, ok, err := h.getSubjectWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: detailFromRequest(c),
		})
	}

	if r.Redirect != 0 {
		return c.Redirect("/v1/subjects/" + strconv.FormatUint(uint64(r.Redirect), 10))
	}

	if r.NSFW && !u.AllowNSFW() {
		// default Handler will return a 404 response
		return c.Next()
	}

	return c.JSON(r)
}

func (h Handler) getSubjectWithCache(ctx context.Context, id uint32) (res.SubjectV0, bool, error) {
	var key = subjectCacheKey(id)

	// try to read from cache
	var r res.SubjectV0
	ok, err := h.cache.Get(ctx, key, &r)
	if err != nil {
		return r, ok, errgo.Wrap(err, "cache.Get")
	}

	if ok {
		return r, ok, nil
	}

	s, err := h.s.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.SubjectV0{}, false, nil
		}

		return r, ok, errgo.Wrap(err, "repo.Set")
	}

	r = convertModelSubject(s)
	if e := h.cache.Set(ctx, key, r, time.Minute); e != nil {
		logger.Error("can't set response to cache", zap.Error(e))
	}

	return r, true, nil
}

func platformString(s model.Subject) *string {
	platform, ok := vars.PlatformMap[s.TypeID][s.PlatformID]
	if !ok {
		logger.Warn("unknown platform",
			zap.Uint8("type", s.TypeID), zap.Uint16("platform", s.PlatformID))

		return nil
	}

	v := platform.String()

	return &v
}

func convertModelSubject(s model.Subject) res.SubjectV0 {
	tags, err := compat.ParseTags(s.CompatRawTags)
	if err != nil {
		logger.Warn("failed to parse tags", zap.Uint32("subject_id", s.ID))
	}

	var date *string
	if s.Date != "" {
		date = &s.Date
	}

	return res.SubjectV0{
		ID:       s.ID,
		Image:    model.SubjectImage(s.Image),
		Summary:  s.Summary,
		Name:     s.Name,
		Platform: platformString(s),
		NameCN:   s.NameCN,
		Date:     date,
		Infobox:  compat.V0Wiki(wiki.ParseOmitError(s.Infobox).NonZero()),
		Volumes:  s.Volumes,
		Redirect: s.Redirect,
		Eps:      s.Eps,
		Tags:     tags,
		Collection: res.Collection{
			OnHold:  s.OnHold,
			Wish:    s.Wish,
			Dropped: s.Dropped,
			Collect: s.Collect,
			Doing:   s.Doing,
		},
		TypeID: s.TypeID,
		Locked: s.Locked(),
		NSFW:   s.NSFW,
		Rating: res.Rating{
			Rank:  s.Rating.Rank,
			Total: s.Rating.Total,
			Count: res.Count{
				Field1:  s.Rating.Count.Field1,
				Field2:  s.Rating.Count.Field2,
				Field3:  s.Rating.Count.Field3,
				Field4:  s.Rating.Count.Field4,
				Field5:  s.Rating.Count.Field5,
				Field6:  s.Rating.Count.Field6,
				Field7:  s.Rating.Count.Field7,
				Field8:  s.Rating.Count.Field8,
				Field9:  s.Rating.Count.Field9,
				Field10: s.Rating.Count.Field10,
			},
			Score: s.Rating.Score,
		},
	}
}
