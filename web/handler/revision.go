// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
//
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
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v2"
	"github.com/mitchellh/mapstructure"

	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/web/res"
)

const CacheDuration = 300

func (h Handler) ListPersonRevision(c *fiber.Ctx) error {
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return err
	}
	personID, err := strparse.Uint32(c.Query("person_id"))
	if err != nil || personID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("bad query person_id: %s", c.Query("person_id")))
	}

	return h.listPersonRevision(c, personID, page)
}

func (h Handler) listPersonRevision(c *fiber.Ctx, personID model.PersonIDType, page pageQuery) error {
	var response = res.Paged{
		Limit:  page.Limit,
		Offset: page.Offset,
	}
	count, err := h.r.CountPersonRelated(c.Context(), personID)
	if err != nil {
		return errgo.Wrap(err, "revision.CountPersonRelated")
	}

	if count == 0 {
		response.Data = []int{}
		return c.JSON(response)
	}

	if err = page.check(count); err != nil {
		return err
	}

	response.Total = count

	revisions, err := h.r.ListPersonRelated(c.Context(), personID, page.Limit, page.Offset)

	if err != nil {
		return errgo.Wrap(err, "revision.ListPersonRelated")
	}

	data := make([]res.PersonRevision, len(revisions))

	creatorMap, err := h.u.GetByIDs(c.Context(), listUniqueCreatorID(revisions)...)

	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}
	for i := range revisions {
		data[i] = convertModelPersonRevision(&revisions[i], creatorMap)
	}
	response.Data = data
	return c.JSON(response)
}

func (h Handler) GetPersonRevision(c *fiber.Ctx) error {
	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("bad param id: %s", c.Params("id")))
	}
	r, err := h.r.GetPersonRelated(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	creatorMap, err := h.u.GetByIDs(c.Context(), r.CreatorID)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	return c.JSON(convertModelPersonRevision(&r, creatorMap))

}

func (h Handler) ListSubjectRevision(c *fiber.Ctx) error {
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return err
	}
	subjectID, err := strparse.Uint32(c.Query("subject_id"))
	if err != nil || subjectID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("bad query subject_id: %s", c.Query("subject_id")))
	}

	return h.listSubjectRevision(c, subjectID, page)
}

func (h Handler) listSubjectRevision(c *fiber.Ctx, subjectID model.SubjectIDType, page pageQuery) error {
	var response = res.Paged{
		Limit:  page.Limit,
		Offset: page.Offset,
	}
	count, err := h.r.CountSubjectRelated(c.Context(), subjectID)
	if err != nil {
		return errgo.Wrap(err, "revision.CountSubjectRelated")
	}

	if count == 0 {
		response.Data = []int{}
		return c.JSON(response)
	}

	if err = page.check(count); err != nil {
		return err
	}

	response.Total = count

	revisions, err := h.r.ListSubjectRelated(c.Context(), subjectID, page.Limit, page.Offset)

	if err != nil {
		return errgo.Wrap(err, "revision.ListSubjectRelated")
	}

	data := make([]res.SubjectRevision, len(revisions))

	creatorMap, err := h.u.GetByIDs(c.Context(), listUniqueCreatorID(revisions)...)

	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}
	for i := range revisions {
		data[i] = convertModelSubjectRevision(&revisions[i], creatorMap)
	}
	response.Data = data
	return c.JSON(response)
}

func (h Handler) GetSubjectRevision(c *fiber.Ctx) error {
	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("bad param id: %s", c.Params("id")))
	}
	r, err := h.r.GetSubjectRelated(c.Context(), id)
	if err != nil {
		return fiber.ErrNotFound
	}

	creatorMap, err := h.u.GetByIDs(c.Context(), r.CreatorID)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	return c.JSON(convertModelSubjectRevision(&r, creatorMap))

}

func listUniqueCreatorID(revisions []model.Revision) []model.IDType {
	m := make(map[model.IDType]bool, len(revisions))
	ret := make([]model.IDType, len(revisions))
	i := 0
	for _, r := range revisions {
		if _, ok := m[r.CreatorID]; !ok {
			m[r.CreatorID] = true
			ret[i] = r.CreatorID
			i++
		}
	}
	return ret[:i]
}

func SafeDecodeExtra(k1 reflect.Type, k2 reflect.Type, input interface{}) (interface{}, error) {
	if k2.Name() == "Extra" && k1.Kind() != reflect.Map {
		return map[string]string{}, nil
	}
	return input, nil
}

func CastPersonData(raw map[string]interface{}) map[string]res.PersonRevisionDataItem {
	if raw == nil {
		return nil
	}
	items := make(map[string]res.PersonRevisionDataItem, len(raw))
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:    "json",
		DecodeHook: SafeDecodeExtra,
		Result:     &items,
	})
	if err != nil {
		return nil
	}
	if err := decoder.Decode(raw); err != nil {
		return nil
	}
	return items
}

func convertModelPersonRevision(r *model.Revision, creatorMap map[model.IDType]model.User) res.PersonRevision {
	creator := creatorMap[r.CreatorID]
	ret := res.PersonRevision{
		ID:      r.ID,
		Type:    r.Type,
		Summary: r.Summary,
		Creator: res.Creator{
			Username: creator.UserName,
			Nickname: creator.UserName,
		},
		CreatedAt: r.CreatedAt,
		Data:      nil,
	}
	if data, ok := r.Data.(map[string]interface{}); ok {
		ret.Data = CastPersonData(data)
	}
	return ret
}

func convertModelSubjectRevision(r *model.Revision, creatorMap map[model.IDType]model.User) res.SubjectRevision {
	creator := creatorMap[r.CreatorID]
	var data *res.SubjectRevisionData
	if r.Data != nil {
		if subjectData, ok := r.Data.(*model.SubjectRevisionData); ok {
			data = &res.SubjectRevisionData{
				Name:         subjectData.Name,
				NameCN:       subjectData.NameCN,
				VoteField:    subjectData.VoteField,
				FieldInfobox: subjectData.FieldInfobox,
				FieldSummary: subjectData.FieldSummary,
				Platform:     subjectData.Platform,
				TypeID:       subjectData.TypeID,
				SubjectID:    subjectData.SubjectID,
				FieldEps:     subjectData.FieldEps,
				Type:         subjectData.Type,
			}
		}
	}
	return res.SubjectRevision{
		ID:      r.ID,
		Type:    r.Type,
		Summary: r.Summary,
		Creator: res.Creator{
			Username: creator.UserName,
			Nickname: creator.UserName,
		},
		CreatedAt: r.CreatedAt,
		Data:      data,
	}

}
