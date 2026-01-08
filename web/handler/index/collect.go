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

package index

import (
	"errors"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h *Handler) CollectIndex(c echo.Context) error {
	iid, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}
	user := accessor.GetFromCtx(c)
	return h.collectIndex(c, iid, user)
}

func (h *Handler) UncollectIndex(c echo.Context) error {
	iid, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}
	user := accessor.GetFromCtx(c)
	return h.uncollectIndex(c, iid, user)
}

func (h *Handler) collectIndex(c echo.Context, indexID uint32, user *accessor.Accessor) error {
	ctx := c.Request().Context()

	if _, ok, err := h.getIndexWithCache(ctx, user, indexID); err != nil {
		return res.InternalError(c, err, "get index error")
	} else if !ok {
		return res.NotFound("index not found")
	}
	// check if the user has collected the index
	if _, err := h.i.GetIndexCollect(ctx, indexID, user.ID); err == nil {
		return nil // already collected
	} else if !errors.Is(err, gerr.ErrNotFound) {
		return res.InternalError(c, err, "get index collect error")
	}
	// add the collect
	if err := h.i.AddIndexCollect(ctx, indexID, user.ID); err != nil {
		return res.InternalError(c, err, "add index collect failed")
	}
	return nil
}

func (h *Handler) uncollectIndex(c echo.Context, indexID uint32, user *accessor.Accessor) error {
	ctx := c.Request().Context()
	if _, ok, err := h.getIndexWithCache(ctx, user, indexID); err != nil {
		return res.InternalError(c, err, "get index error")
	} else if !ok {
		return res.NotFound("index not found")
	}
	// check if the user has collected the index
	if _, err := h.i.GetIndexCollect(ctx, indexID, user.ID); err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("index not collected")
		}
		return res.InternalError(c, err, "get index collect error")
	}
	// delete the collect
	if err := h.i.DeleteIndexCollect(ctx, indexID, user.ID); err != nil {
		return res.InternalError(c, err, "delete index collect failed")
	}
	return nil
}
