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
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/web/res"
)

func (h Handler) getCollection(c *fiber.Ctx, u model.User, page pageQuery, showPrivate bool) error {
	count, err := h.u.CountCollections(c.Context(), u.ID, 0, 0, showPrivate)
	if err != nil {
		return errgo.Wrap(err, "user.CountCollections")
	}

	if count == 0 {
		return c.JSON(res.Paged{Data: []int{}, Total: count, Limit: page.Limit, Offset: page.Offset})
	}

	if err = page.check(count); err != nil {
		return err
	}

	collections, err := h.u.ListCollections(c.Context(),
		u.ID, 0, 0, showPrivate, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "user.ListCollections")
	}

	var data = make([]res.Collection, len(collections))
	for i, collection := range collections {
		c := res.Collection{
			SubjectID:   collection.SubjectID,
			SubjectType: collection.SubjectType,
			Rate:        collection.Rate,
			Type:        collection.Type,
			Tags:        collection.Tags,
			EpStatus:    collection.EpStatus,
			VolStatus:   collection.VolStatus,
			UpdatedAt:   collection.UpdatedAt,
			Private:     collection.Private,
		}

		if collection.HasComment {
			c.Comment = &collection.Comment
		}

		data[i] = c
	}

	return c.JSON(res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}

func (h Handler) ListCollection(c *fiber.Ctx) error {
	v := h.getUser(c)
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "bad query args: "+err.Error())
	}

	username := c.Params("username")
	if username == "" {
		return fiber.NewError(http.StatusBadRequest, "missing require parameters `username`")
	}

	u, err := h.u.GetByName(c.Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fiber.NewError(http.StatusNotFound, "user doesn't exist or has been removed")
		}

		return errgo.Wrap(err, "user.GetByName")
	}

	var showPrivate = u.ID == v.ID

	return h.getCollection(c, u, page, showPrivate)
}
