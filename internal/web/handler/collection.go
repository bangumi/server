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
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/logger/log"
	"github.com/bangumi/server/internal/model"
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

	subjectType, err := parseSubjectType(c.Query("subject_type"))
	if err != nil {
		return res.BadRequest(err.Error())
	}

	collectionType, err := parseCollectionType(c.Query("type"))
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

	var data = make([]res.SubjectCollection, len(collections))
	for i, collection := range collections {
		data[i] = convertModelSubjectCollection(collection)
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

	return res.JSON(c, convertModelSubjectCollection(collection))
}

const subjectNotFoundMessage = "subject not found"

func (h Handler) PatchSubjectCollection(c *fiber.Ctx) error {
	v := h.getHTTPAccessor(c)
	if !v.login {
		return res.Unauthorized(res.DefaultUnauthorizedMessage)
	}

	subjectID, err := parseSubjectID(c.Params("subject_id"))
	if err != nil {
		return err
	}

	s, ok, err := h.getSubjectWithCache(c.Context(), subjectID)
	if err != nil {
		return h.InternalError(c, err, "failed to get subject info")
	}
	if !ok || s.Redirect != 0 {
		return res.NotFound(subjectNotFoundMessage)
	}

	var input req.PatchSubjectCollection
	if err = json.UnmarshalNoEscape(c.Body(), &input); err != nil {
		return res.FromError(c, err, http.StatusUnprocessableEntity, "can't decode request body as json")
	}

	if err = validateSubjectCollectionRequest(s, input); err != nil {
		return err
	}

	return h.patchSubjectCollection(c, v.ID, s, input)
}

func (h Handler) patchSubjectCollection(
	c *fiber.Ctx, userID model.UserID, s res.SubjectV0, input req.PatchSubjectCollection,
) error {
	collection, err := h.collect.GetSubjectCollection(c.Context(), userID, s.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.NotFound(subjectNotFoundMessage)
		}

		return errgo.Wrap(err, "user.GetSubjectCollection")
	}

	update := model.SubjectCollectionUpdate{
		UpdatedAt:   time.Now(),
		SubjectType: s.TypeID,
		Rate:        collection.Rate,
		Comment:     input.Comment.Default(collection.Comment),
		VolStatus:   input.VolStatus.Default(collection.VolStatus),
		EpStatus:    input.EpStatus.Default(collection.EpStatus),
		Type:        model.CollectionType(input.Type.Default(uint8(collection.Type))),
		Private:     input.Private.Default(collection.Private),
	}

	if input.Tags != nil {
		update.Tags = input.Tags
	}

	if err = h.collect.UpdateSubjectCollection(c.Context(), userID, s.ID, update); err != nil {
		return h.InternalError(c, err, "failed to update subject collection")
	}

	collection, err = h.collect.GetSubjectCollection(c.Context(), userID, s.ID)
	if err != nil {
		return errgo.Wrap(err, "collection.GetSubjectCollection")
	}

	return res.JSON(c, convertModelSubjectCollection(collection))
}

func validateSubjectCollectionRequest(s res.SubjectV0, input req.PatchSubjectCollection) error {
	if s.TypeID == model.SubjectTypeMusic {
		if input.VolStatus.HasValue() || input.EpStatus.HasValue() {
			return res.BadRequest("can't update `vol_status` or `ep_status` for music subject")
		}
	}

	if s.TypeID != model.SubjectTypeBook && input.VolStatus.HasValue() {
		return res.BadRequest("subject is not book, you can't add `vol_status` key to request body")
	}

	return nil
}

func convertModelSubjectCollection(collection model.SubjectCollection) res.SubjectCollection {
	return res.SubjectCollection{
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
}
