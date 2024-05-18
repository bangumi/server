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

func (h User) ListCharacterCollection(c echo.Context) error {
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

	return h.listCharacterCollection(c, u, page)
}

func (h User) listCharacterCollection(c echo.Context, u user.User, page req.PageQuery) error {
	count, err := h.collect.CountPersonCollections(c.Request().Context(), u.ID, collection.PersonCollectCategoryCharacter)
	if err != nil {
		return errgo.Wrap(err, "failed to count user's character collections")
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
		c.Request().Context(),
		u.ID, collection.PersonCollectCategoryCharacter,
		page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "failed to list user's person collections")
	}

	characterIDs := slice.Map(cols, func(item collection.UserPersonCollection) model.PersonID {
		return item.TargetID
	})

	characterMap, err := h.character.GetByIDs(c.Request().Context(), characterIDs)
	if err != nil {
		return errgo.Wrap(err, "failed to get persons")
	}

	var data = make([]res.PersonCollection, 0, len(cols))

	for _, col := range cols {
		character, ok := characterMap[col.TargetID]
		if !ok {
			continue
		}
		data = append(data, res.ConvertModelCharacterCollection(col, character))
	}

	return c.JSON(http.StatusOK, res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}
