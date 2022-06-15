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

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) GetGroupByName(c *fiber.Ctx) error {
	groupName := c.Params("name")
	g, err := h.g.GetByName(c.Context(), groupName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound(c)
		}

		return res.InternalError(c, err, "un expected error")
	}

	return res.JSON(c, res.Group{
		Name:        g.Name,
		Title:       g.Title,
		Description: g.Description,
		Icon:        "https://lain.bgm.tv/pic/icon/l/" + g.Icon,
		// ID:          g.ID,
	})
}
