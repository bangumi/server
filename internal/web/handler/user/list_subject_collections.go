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

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h User) ListSubjectCollection(c *fiber.Ctx) error {
	v := h.GetHTTPAccessor(c)
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	username := c.Params("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}

	subjectType, err := req.ParseSubjectType(c.Query("subject_type"))
	if err != nil {
		return res.BadRequest(err.Error())
	}

	collectionType, err := req.ParseCollectionType(c.Query("type"))
	if err != nil {
		return err
	}

	u, err := h.user.GetByName(c.Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("user doesn't exist or has been removed")
		}

		return errgo.Wrap(err, "user.GetByName")
	}

	var showPrivate = u.ID == v.ID

	return h.listCollection(c, u, subjectType, collectionType, page, showPrivate)
}

func (h User) listCollection(
	c *fiber.Ctx,
	u model.User,
	subjectType model.SubjectType,
	collectionType model.SubjectCollection,
	page req.PageQuery,
	showPrivate bool,
) error {
	count, err := h.collect.CountSubjectCollections(c.Context(), u.ID, subjectType, collectionType, showPrivate)
	if err != nil {
		return h.InternalError(c, err, "failed to count user's subject collections", log.UserID(u.ID))
	}

	if count == 0 {
		return c.JSON(res.Paged{Data: []int{}, Total: count, Limit: page.Limit, Offset: page.Offset})
	}

	if err = page.Check(count); err != nil {
		return err
	}

	collections, err := h.collect.ListSubjectCollection(c.Context(),
		u.ID, subjectType, collectionType, showPrivate, page.Limit, page.Offset)
	if err != nil {
		return h.InternalError(c, err, "failed to list user's subject collections", log.UserID(u.ID))
	}

	subjectIDs := slice.Map(collections, func(item model.UserSubjectCollection) model.SubjectID {
		return item.SubjectID
	})

	subjectMap, err := h.app.Query.GetSubjectByIDs(c.Context(), subjectIDs...)
	if err != nil {
		return h.InternalError(c, err, "failed to get subjects")
	}

	var data = make([]res.SubjectCollection, len(collections))
	for i, collection := range collections {
		s := subjectMap[collection.SubjectID]
		data[i] = res.ConvertModelSubjectCollection(collection, res.ToSlimSubjectV0(s))
	}

	return c.JSON(res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}
