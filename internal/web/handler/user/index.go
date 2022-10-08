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
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h User) Indices(c *fiber.Ctx) error {
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}
	username := c.Params("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}
	user, err := h.user.GetByName(c.UserContext(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("can't find user with username " + strconv.Quote(username))
		}
		return errgo.Wrap(err, "failed to get user by username")
	}

	arr, err := h.index.GetIndicesByUser(c.UserContext(), user.ID, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "failed to get indices by user id")
	}

	ret := make([]res.Index, len(arr))
	for i := range arr {
		ret[i] = res.ConvertIndexModel(arr[i], user)
	}

	return c.JSON(ret)
}

func (h User) CollectedIndices(c *fiber.Ctx) error {
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}
	username := c.Params("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}
	user, err := h.user.GetByName(c.UserContext(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("can't find user with username " + strconv.Quote(username))
		}
		return errgo.Wrap(err, "failed to get user by username")
	}

	arr, err := h.index.GetCollectedIndicesByUser(c.UserContext(), user.ID, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "failed to get indices by user id")
	}

	return c.JSON(
		res.ConvertIndexCollectModel(arr, user),
	)
}
