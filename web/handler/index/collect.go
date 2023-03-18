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

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
	"github.com/labstack/echo/v4"
)

func (h *Handler) CollectIndex(c echo.Context) error {
	iid, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}
	uid := accessor.GetFromCtx(c).ID
	return h.collectIndex(c, iid, uid)
}

func (h *Handler) UncollectIndex(c echo.Context) error {
	iid, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}
	uid := accessor.GetFromCtx(c).ID
	return h.collectIndex(c, iid, uid)
}

func (h *Handler) collectIndex(c echo.Context, indexID uint32, uid uint32) error {
	ctx := c.Request().Context()
	if _, err := h.i.Get(ctx, indexID); err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("index not found")
		}
		return res.InternalError(c, err, "get index error")
	}
	if err := h.i.AddIndexCollect(ctx, indexID, uid); err != nil {
		return res.InternalError(c, err, "add index collect failed")
	}
	return nil
}

func (h *Handler) uncollectIndex(c echo.Context, indexID uint32, uid uint32) error {
	ctx := c.Request().Context()
	if _, err := h.i.Get(ctx, indexID); err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("index not found")
		}
		return res.InternalError(c, err, "get index error")
	}
	if err := h.i.DeleteIndexCollect(ctx, indexID, uid); err != nil {
		return res.InternalError(c, err, "delete index collect failed")
	}
	return nil
}
