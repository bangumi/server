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
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) GetPersonComments(c *fiber.Ctx) error {
	id, err := parsePersonID(c.Params("id"))
	if err != nil {
		return err
	}

	r, ok, err := h.getPersonWithCache(c.Context(), id)
	if err != nil {
		return err
	}

	if !ok || r.Redirect != 0 {
		return res.ErrNotFound
	}

	pagedComments, err := h.listComments(c, domain.CommentPerson, model.TopicID(id))
	if err != nil {
		return err
	}
	return c.JSON(pagedComments)
}
