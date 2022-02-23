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

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/web/handler/cachekey"
	"github.com/bangumi/server/web/res"
	"github.com/bangumi/server/web/util"
)

func (h Handler) GetEpisode(c *fiber.Ctx) error {
	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id == 0 {
		return fiber.NewError(http.StatusBadRequest, "bad id: "+c.Params("id"))
	}

	r, ok, err := h.getEpisodeWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok {
		return c.Status(http.StatusNotFound).JSON(res.Error{
			Title:   "Not Found",
			Details: util.DetailFromRequest(c),
		})
	}

	return c.JSON(r)
}

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
