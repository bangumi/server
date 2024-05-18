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

package user

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/user"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h User) ListPersonCollection(c echo.Context) error {
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	username := c.Param("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}

	u, err := h.user.GetByName(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("user doesn't exist or has been removed")
		}

		return errgo.Wrap(err, "user.GetByName")
	}

	return h.listPersonCollection(c, u, page)
}

func (h User) listPersonCollection(c echo.Context, u user.User, page req.PageQuery) error {
	count, err := h.collect.CountPersonCollections(c.Request().Context(), u.ID, collection.PersonCollectCategoryPerson)
	if err != nil {
		return errgo.Wrap(err, "failed to count user's person collections")
	}

	if count == 0 {
		if count == 0 {
			return c.JSON(http.StatusOK, res.Paged{Data: []int{}, Total: count, Limit: page.Limit, Offset: page.Offset})
		}
	}

	if err = page.Check(count); err != nil {
		return err
	}

	cols, err := h.collect.ListPersonCollection(
		c.Request().Context(), u.ID, collection.PersonCollectCategoryPerson,
		page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "failed to list user's person collections")
	}

	personIDs := slice.Map(cols, func(item collection.UserPersonCollection) model.PersonID {
		return item.TargetID
	})

	personMap, err := h.person.GetByIDs(c.Request().Context(), personIDs)
	if err != nil {
		return errgo.Wrap(err, "failed to get persons")
	}

	var data = make([]res.PersonCollection, 0, len(cols))

	for _, col := range cols {
		person, ok := personMap[col.TargetID]
		if !ok {
			continue
		}
		data = append(data, res.ConvertModelPersonCollection(col, person))
	}

	return c.JSON(http.StatusOK, res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}
