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
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/strparse"
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
		return res.NewError(http.StatusBadRequest, "missing require parameters `username`")
	}

	subjectType, err := parseSubjectType(c.Query("subject_type"))
	if err != nil {
		return res.NewError(http.StatusBadRequest, err.Error())
	}

	collectionType, err := parseCollectionType(c.Query("type"))
	if err != nil {
		return res.NewError(http.StatusBadRequest, "bad query 'type': "+err.Error())
	}

	u, err := h.u.GetByName(c.Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NewError(http.StatusNotFound, "user doesn't exist or has been removed")
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
	collectionType uint8,
	page pageQuery,
	showPrivate bool,
) error {
	count, err := h.u.CountCollections(c.Context(), u.ID, subjectType, collectionType, showPrivate)
	if err != nil {
		return errgo.Wrap(err, "user.CountCollections")
	}

	if count == 0 {
		return c.JSON(res.Paged{Data: []int{}, Total: count, Limit: page.Limit, Offset: page.Offset})
	}

	if err = page.check(count); err != nil {
		return err
	}

	collections, err := h.u.ListCollections(c.Context(),
		u.ID, subjectType, collectionType, showPrivate, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "user.ListCollections")
	}

	var data = make([]res.Collection, len(collections))
	for i, collection := range collections {
		c := res.Collection{
			SubjectID:   collection.SubjectID,
			SubjectType: collection.SubjectType,
			Rate:        collection.Rate,
			Type:        collection.Type,
			Tags:        collection.Tags,
			EpStatus:    collection.EpStatus,
			VolStatus:   collection.VolStatus,
			UpdatedAt:   collection.UpdatedAt,
			Private:     collection.Private,
			Comment:     nilString(collection.Comment),
		}

		data[i] = c
	}

	return c.JSON(res.Paged{
		Data:   data,
		Total:  count,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}

func parseCollectionType(s string) (uint8, error) {
	if s == "" {
		return 0, nil
	}

	t, err := strparse.Uint8(s)
	if err != nil {
		return 0, res.NewError(http.StatusBadRequest, "bad collection type: "+strconv.Quote(s))
	}

	switch t {
	case 1, 2, 3, 4, 5: //nolint:gomnd
		return t, nil
	}

	return 0, res.NewError(http.StatusBadRequest, strconv.Quote(s)+"is not a valid collection type")
}

func (h Handler) GetCollection(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return res.BadRequest("missing require parameters `username`")
	}

	subjectID, err := parseSubjectID(c.Params("subject_id"))
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
			return res.NewError(http.StatusNotFound, "user doesn't exist or has been removed")
		}

		return errgo.Wrap(err, "user.GetByName")
	}

	var showPrivate = u.ID == v.ID

	collection, err := h.u.GetCollection(c.Context(), u.ID, subjectID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound(notFoundMessage)
		}

		return errgo.Wrap(err, "user.GetCollection")
	}

	if !showPrivate && collection.Private {
		return res.NotFound(notFoundMessage)
	}

	return res.JSON(c, res.Collection{
		SubjectID:   collection.SubjectID,
		SubjectType: collection.SubjectType,
		Rate:        collection.Rate,
		Type:        collection.Type,
		Tags:        collection.Tags,
		EpStatus:    collection.EpStatus,
		VolStatus:   collection.VolStatus,
		UpdatedAt:   collection.UpdatedAt,
		Private:     collection.Private,
		Comment:     nilString(collection.Comment),
	})
}
