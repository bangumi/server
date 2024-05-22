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
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Character) Get(c echo.Context) error {
	u := accessor.GetFromCtx(c)
	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	r, err := h.character.Get(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get character")
	}

	if r.Redirect != 0 {
		return c.Redirect(http.StatusFound, "/v0/characters/"+strconv.FormatUint(uint64(r.Redirect), 10))
	}

	if !auth.AllowReadCharacter(u.Auth, r) {
		return res.ErrNotFound
	}

	return c.JSON(http.StatusOK, convertModelCharacter(r))
}

func (h Character) GetImage(c echo.Context) error {
	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	p, err := h.character.Get(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}
		return errgo.Wrap(err, "failed to get character")
	}

	l, ok := res.PersonImage(p.Image).Select(c.QueryParam("type"))
	if !ok {
		return res.BadRequest("bad image type: " + c.QueryParam("type"))
	}

	if l == "" {
		return c.Redirect(http.StatusFound, res.DefaultImageURL)
	}

	return c.Redirect(http.StatusFound, l)
}
