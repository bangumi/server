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

	"github.com/labstack/echo/v5"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h User) GetPersonCollection(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}

	personID, err := req.ParseID(c.Param("person_id"))
	if err != nil {
		return err
	}

	return h.getPersonCollection(c, username, personID)
}

func (h User) getPersonCollection(c echo.Context, username string, personID model.PersonID) error {
	const notFoundMessage = "person is not collected by user"

	person, err := h.person.Get(c.Request().Context(), personID)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get person")
	}

	u, err := h.user.GetByName(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("user doesn't exist or has been removed")
		}

		return errgo.Wrap(err, "failed to get user by name")
	}

	collect, err := h.collect.GetPersonCollection(
		c.Request().Context(), u.ID, collection.PersonCollectCategoryPerson, personID)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound(notFoundMessage)
		}

		return errgo.Wrap(err, "failed to get person collect")
	}

	return c.JSON(http.StatusOK, res.ConvertModelPersonCollection(collect, person))
}
