// Copyright (c) 2022 Sociosarbis <136657577@qq.com>
// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
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

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/model"
	"github.com/bangumi/server/web/res"
)

const CacheDuration = 300

func (h Handler) ListPersonRevision(c *fiber.Ctx) error {
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("bad query args: %s", err.Error()))
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
		data[i] = convertModelRevision(&revisions[i], creatorMap)
	}
	response.Data = data
	return c.JSON(response)
}

func (h Handler) GetPersionRevision(c *fiber.Ctx) error {
	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("bad param id: %s", c.Params("id")))
	}
	r, err := h.r.GetPersonRelated(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound)
	}

	creatorMap, err := h.u.GetByIDs(c.Context(), r.CreatorID)

	if err != nil {
		return errgo.Wrap(err, "user.GetByIDS")
	}

	return c.JSON(convertModelRevision(&r, creatorMap))

}

func listUniqueCreatorID(revisions []model.Revision) []model.IDType {
	m := map[model.IDType]bool{}
	ret := []model.IDType{}
	for _, r := range revisions {
		if _, ok := m[r.CreatorID]; !ok {
			m[r.CreatorID] = true
			ret = append(ret, r.CreatorID)
		}
	}
	return ret
}

func CastPersonData(raw map[string]interface{}) map[string]res.PersonRevisionDataItem {
	data, err := json.Marshal(raw)
	if err == nil {
		item := map[string]res.PersonRevisionDataItem{}
		err = json.Unmarshal(data, &item)
		if err == nil {
			return item
		}
	}
	return nil
}

func convertModelRevision(r *model.Revision, creatorMap map[model.IDType]model.User) res.PersonRevision {
	creator := creatorMap[r.CreatorID]
	return res.PersonRevision{
		ID:      r.ID,
		Type:    r.Type,
		Summary: r.Summary,
		Creator: res.Creator{
			Username: creator.UserName,
			Nickname: creator.UserName,
		},
		CreatedAt: r.CreatedAt,
		Data:      CastPersonData(r.Data),
	}

}
