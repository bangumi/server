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

func (h User) GetCharacterCollection(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}

	characterID, err := req.ParseID(c.Param("character_id"))
	if err != nil {
		return err
	}

	return h.getCharacterCollection(c, username, characterID)
}

func (h User) getCharacterCollection(c echo.Context, username string, characterID model.CharacterID) error {
	const notFoundMessage = "character is not collected by user"

	character, err := h.character.Get(c.Request().Context(), characterID)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get character")
	}

	u, err := h.user.GetByName(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("user doesn't exist or has been removed")
		}

		return errgo.Wrap(err, "failed to get user by name")
	}

	collect, err := h.collect.GetPersonCollection(
		c.Request().Context(),
		u.ID,
		collection.PersonCollectCategoryCharacter,
		characterID,
	)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound(notFoundMessage)
		}

		return errgo.Wrap(err, "failed to get user's character collection")
	}

	return c.JSON(http.StatusOK, res.ConvertModelCharacterCollection(collect, character))
}
