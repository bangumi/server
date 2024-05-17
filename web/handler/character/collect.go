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

package character

import (
	"errors"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Character) CollectCharacter(c echo.Context) error {
	cid, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}
	uid := accessor.GetFromCtx(c).ID
	return h.collectCharacter(c, cid, uid)
}

func (h Character) UncollectCharacter(c echo.Context) error {
	cid, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}
	uid := accessor.GetFromCtx(c).ID
	return h.uncollectCharacter(c, cid, uid)
}

func (h Character) collectCharacter(c echo.Context, cid uint32, uid uint32) error {
	ctx := c.Request().Context()
	// check if the character exists
	if _, err := h.character.Get(ctx, cid); err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}
		return res.InternalError(c, err, "get character error")
	}
	// check if the user has collected the character
	if _, err := h.collect.GetPersonCollect(ctx, uid, collection.PersonCollectCategoryCharacter, cid); err == nil {
		return nil // already collected
	} else if !errors.Is(err, gerr.ErrNotFound) {
		return res.InternalError(c, err, "get character collect error")
	}
	// add the collect
	if err := h.collect.AddPersonCollect(ctx, uid, collection.PersonCollectCategoryCharacter, cid); err != nil {
		return res.InternalError(c, err, "add character collect failed")
	}
	return nil
}

func (h Character) uncollectCharacter(c echo.Context, cid uint32, uid uint32) error {
	ctx := c.Request().Context()
	// check if the character exists
	if _, err := h.character.Get(ctx, cid); err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}
		return res.InternalError(c, err, "get character error")
	}
	// check if the user has collected the character
	if _, err := h.collect.GetPersonCollect(ctx, uid, collection.PersonCollectCategoryCharacter, cid); err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("character not collected")
		}
		return res.InternalError(c, err, "get character collect error")
	}
	// remove the collect
	if err := h.collect.RemovePersonCollect(ctx, uid, collection.PersonCollectCategoryCharacter, cid); err != nil {
		return res.InternalError(c, err, "remove character collect failed")
	}
	return nil
}
