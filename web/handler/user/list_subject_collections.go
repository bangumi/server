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
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/internal/user"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h User) ListSubjectCollection(c echo.Context) error {
	v := accessor.GetFromCtx(c)
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	username := c.Param("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}

	subjectType, err := req.ParseSubjectType(c.QueryParam("subject_type"))
	if err != nil {
		return res.BadRequest(err.Error())
	}

	collectionType, err := req.ParseCollectionType(c.QueryParam("type"))
	if err != nil {
		return err
	}

	u, err := h.user.GetByName(c.Request().Context(), username)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.NotFound("user doesn't exist or has been removed")
		}

		return errgo.Wrap(err, "user.GetByName")
	}

	var showPrivate = u.ID == v.ID

	return h.listCollection(c, u, subjectType, collectionType, page, showPrivate)
}

func (h User) listCollection(
	c echo.Context,
	u user.User,
	subjectType model.SubjectType,
	collectionType collection.SubjectCollection,
	page req.PageQuery,
	showPrivate bool,
) error {
	count, err := h.collect.CountSubjectCollections(c.Request().Context(), u.ID, subjectType, collectionType, showPrivate)
	if err != nil {
		return errgo.Wrap(err, "failed to count user's subject collections")
	}

	if count == 0 {
		return c.JSON(http.StatusOK, res.Paged{Data: []int{}, Total: count, Limit: page.Limit, Offset: page.Offset})
	}

	if err = page.Check(count); err != nil {
		return err
	}

	collections, err := h.collect.ListSubjectCollection(c.Request().Context(),
		u.ID, subjectType, collectionType, showPrivate, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "failed to list user's subject collections")
	}

	subjectIDs := slice.Map(collections, func(item collection.UserSubjectCollection) model.SubjectID {
		return item.SubjectID
	})

	subjectMap, err := h.subject.GetByIDs(c.Request().Context(), subjectIDs, subject.Filter{})
	if err != nil {
		return errgo.Wrap(err, "failed to get subjects")
	}

	var data = make([]res.SubjectCollection, 0, len(collections))
	for _, collect := range collections {
		s, ok := subjectMap[collect.SubjectID]
		if !ok {
			continue
		}

		data = append(data, res.ConvertModelSubjectCollection(collect, res.ToSlimSubjectV0(s)))
	}

	return c.JSON(http.StatusOK, res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}
