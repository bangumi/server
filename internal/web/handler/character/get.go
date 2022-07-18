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
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h Character) Get(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)
	id, err := req.ParseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	r, err := h.app.Query.GetCharacter(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return h.InternalError(c, err, "failed to get character", log.CharacterID(id))
	}

	if r.Redirect != 0 {
		return c.Redirect("/v0/characters/" + strconv.FormatUint(uint64(r.Redirect), 10))
	}

	return res.JSON(c, convertModelCharacter(r))
}

func (h Character) GetImage(c *fiber.Ctx) error {
	u := h.GetHTTPAccessor(c)
	id, err := req.ParseCharacterID(c.Params("id"))
	if err != nil {
		return err
	}

	p, err := h.app.Query.GetCharacterNoRedirect(c.Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}
		return h.InternalError(c, err, "failed to get character", log.CharacterID(id))
	}

	l, ok := res.PersonImage(p.Image).Select(c.Query("type"))
	if !ok {
		return res.BadRequest("bad image type: " + c.Query("type"))
	}

	if l == "" {
		return c.Redirect(res.DefaultImageURL)
	}

	return c.Redirect(l)
}
