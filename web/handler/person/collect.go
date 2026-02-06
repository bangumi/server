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

package person

import (
	"errors"

	"github.com/labstack/echo/v5"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Person) CollectPerson(c echo.Context) error {
	pid, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	uid := accessor.GetFromCtx(c).ID
	return h.collectPerson(c, pid, uid)
}

func (h Person) UncollectPerson(c echo.Context) error {
	pid, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	uid := accessor.GetFromCtx(c).ID
	return h.uncollectPerson(c, pid, uid)
}

func (h Person) collectPerson(c echo.Context, pid uint32, uid uint32) error {
	ctx := c.Request().Context()
	// check if the person exists
	if _, err := h.person.Get(ctx, pid); err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}
		return res.InternalError(c, err, "get person error")
	}
	// check if the user has collected the person
	if _, err := h.collect.GetPersonCollection(ctx, uid, collection.PersonCollectCategoryPerson, pid); err == nil {
		return nil // already collected
	} else if !errors.Is(err, gerr.ErrNotFound) {
		return res.InternalError(c, err, "get person collect error")
	}
	// add the collect
	if err := h.collect.AddPersonCollection(ctx, uid, collection.PersonCollectCategoryPerson, pid); err != nil {
		return res.InternalError(c, err, "add person collect failed")
	}
	return nil
}

func (h Person) uncollectPerson(c echo.Context, pid uint32, uid uint32) error {
	ctx := c.Request().Context()
	// check if the person exists
	if _, err := h.person.Get(ctx, pid); err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}
		return res.InternalError(c, err, "get person error")
	}
	// check if the user has collected the person
	if _, err := h.collect.GetPersonCollection(ctx, uid, collection.PersonCollectCategoryPerson, pid); err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("person not collected")
		}
		return res.InternalError(c, err, "get person collect error")
	}
	// remove the collect
	if err := h.collect.RemovePersonCollection(ctx, uid, collection.PersonCollectCategoryPerson, pid); err != nil {
		return res.InternalError(c, err, "remove person collect failed")
	}
	return nil
}
