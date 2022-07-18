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
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/logger/log"
	"github.com/bangumi/server/internal/web/req"
	"github.com/bangumi/server/internal/web/res"
)

func (h Handler) ListCollection(c *fiber.Ctx) error {
	v := h.getHTTPAccessor(c)
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
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

	u, err := h.u.GetByName(c.Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("user doesn't exist or has been removed")
		}

		return errgo.Wrap(err, "user.GetByName")
	}

	var showPrivate = u.ID == v.ID

	return h.listCollection(c, u, subjectType, collectionType, page, showPrivate)
}

func (h Handler) listCollection(
	c *fiber.Ctx,
	u model.User,
	subjectType model.SubjectType,
	collectionType model.CollectionType,
	page pageQuery,
	showPrivate bool,
) error {
	count, err := h.collect.CountSubjectCollections(c.Context(), u.ID, subjectType, collectionType, showPrivate)
	if err != nil {
		return h.InternalError(c, err, "failed to count user's subject collections", log.UserID(u.ID))
	}

	if count == 0 {
		return c.JSON(res.Paged{Data: []int{}, Total: count, Limit: page.Limit, Offset: page.Offset})
	}

	if err = page.check(count); err != nil {
		return err
	}

	collections, err := h.collect.ListSubjectCollection(c.Context(),
		u.ID, subjectType, collectionType, showPrivate, page.Limit, page.Offset)
	if err != nil {
		return h.InternalError(c, err, "failed to list user's subject collections", log.UserID(u.ID))
	}

	subjectIDs := slice.Map(collections, func(item model.SubjectCollection) model.SubjectID {
		return item.SubjectID
	})

	subjectMap, err := h.app.Query.GetSubjectByIDs(c.Context(), subjectIDs...)
	if err != nil {
		return h.InternalError(c, err, "failed to get subjects")
	}

	var data = make([]res.SubjectCollection, len(collections))
	for i, collection := range collections {
		s := subjectMap[collection.SubjectID]
		data[i] = convertModelSubjectCollection(collection, res.ToSlimSubjectV0(s))
	}

	return c.JSON(res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}

func (h Handler) GetCollection(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}

	subjectID, err := req.ParseSubjectID(c.Params("subject_id"))
	if err != nil {
		return err
	}

	return h.getCollection(c, username, subjectID)
}

func (h Handler) getCollection(c *fiber.Ctx, username string, subjectID model.SubjectID) error {
	const notFoundMessage = "subject is not collected by user"
	v := h.getHTTPAccessor(c)

	u, err := h.u.GetByName(c.Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound("user doesn't exist or has been removed")
		}

		return h.InternalError(c, err, "failed to get user by name", zap.String("name", username))
	}

	var showPrivate = u.ID == v.ID

	collection, err := h.collect.GetSubjectCollection(c.Context(), u.ID, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound(notFoundMessage)
		}

		return h.InternalError(c, err, "failed to get user's subject collection",
			log.UserID(u.ID), log.SubjectID(subjectID))
	}

	if !showPrivate && collection.Private {
		return res.NotFound(notFoundMessage)
	}

	s, err := h.app.Query.GetSubject(c.Context(), v.Auth, subjectID)
	if err != nil {
		return h.InternalError(c, err, "failed to subject info", log.SubjectID(subjectID))
	}

	return res.JSON(c, convertModelSubjectCollection(collection, res.ToSlimSubjectV0(s)))
}

func convertModelSubjectCollection(c model.SubjectCollection, subject res.SlimSubjectV0) res.SubjectCollection {
	return res.SubjectCollection{
		SubjectID:   c.SubjectID,
		SubjectType: c.SubjectType,
		Rate:        c.Rate,
		Type:        c.Type,
		Tags:        c.Tags,
		EpStatus:    c.EpStatus,
		VolStatus:   c.VolStatus,
		UpdatedAt:   c.UpdatedAt,
		Private:     c.Private,
		Comment:     nilString(c.Comment),
		Subject:     subject,
	}
}
