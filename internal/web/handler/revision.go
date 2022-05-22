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
	"fmt"
	"strconv"

	"github.com/goccy/go-reflect"
	"github.com/gofiber/fiber/v2"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/strparse"
	"github.com/bangumi/server/internal/web/res"
)

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
		if errors.Is(err, domain.ErrNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.ErrInternalServerError
	}

	creatorMap, err := h.u.GetByIDs(c.Context(), r.CreatorID)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	return c.JSON(convertModelPersonRevision(&r, creatorMap))
}

func (h Handler) ListCharacterRevision(c *fiber.Ctx) error {
	page, err := getPageQuery(c, defaultPageLimit, defaultMaxPageLimit)
	if err != nil {
		return err
	}
	characterID, err := strparse.Uint32(c.Query("character_id"))
	if err != nil || characterID <= 0 {
		return fiber.NewError(
			fiber.StatusBadRequest,
			fmt.Sprintf("bad query character_id: %s", strconv.Quote(c.Query("character_id"))),
		)
	}
	return h.listCharacterRevision(c, characterID, page)
}

func (h Handler) listCharacterRevision(c *fiber.Ctx, characterID model.CharacterIDType, page pageQuery) error {
	var response = res.Paged{
		Limit:  page.Limit,
		Offset: page.Offset,
	}
	count, err := h.r.CountCharacterRelated(c.Context(), characterID)
	if err != nil {
		return errgo.Wrap(err, "revision.CountCharacterRelated")
	}

	if count == 0 {
		response.Data = []int{}
		return c.JSON(response)
	}

	if err = page.check(count); err != nil {
		return err
	}

	response.Total = count

	revisions, err := h.r.ListCharacterRelated(c.Context(), characterID, page.Limit, page.Offset)

	if err != nil {
		return errgo.Wrap(err, "revision.ListCharacterRelated")
	}

	creatorIDs := make([]uint32, 0, len(revisions))
	for _, revision := range revisions {
		creatorIDs = append(creatorIDs, revision.RevisionCommon.CreatorID)
	}
	creatorMap, err := h.u.GetByIDs(c.Context(), dedupeCreatorID(creatorIDs)...)

	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	data := make([]res.CharacterRevision, len(revisions))
	for i := range revisions {
		data[i] = convertModelCharacterRevision(&revisions[i], creatorMap)
	}
	response.Data = data
	return c.JSON(response)
}

func (h Handler) GetCharacterRevision(c *fiber.Ctx) error {
	id, err := strparse.Uint32(c.Params("id"))
	if err != nil || id <= 0 {
		return fiber.NewError(
			fiber.StatusBadRequest,
			fmt.Sprintf("bad param id: %s", strconv.Quote(c.Params("id"))),
		)
	}
	r, err := h.r.GetCharacterRelated(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.ErrInternalServerError
	}

	creatorMap, err := h.u.GetByIDs(c.Context(), r.CreatorID)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	return c.JSON(convertModelCharacterRevision(&r, creatorMap))
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
	if err != nil || id == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "bad param id: "+strconv.Quote(c.Params("id")))
	}
	r, err := h.r.GetSubjectRelated(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.ErrInternalServerError
	}

	creatorMap, err := h.u.GetByIDs(c.Context(), r.CreatorID)
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	return c.JSON(convertModelSubjectRevision(&r, creatorMap))
}

func dedupeCreatorID(creatorIDs []uint32) []model.UIDType {
	m := make(map[model.UIDType]bool, len(creatorIDs))
	ret := make([]model.UIDType, 0, len(creatorIDs))
	for _, r := range creatorIDs {
		if _, ok := m[r]; !ok {
			m[r] = true
			ret = append(ret, r)
		}
	}
	return ret
}

func listUniqueCreatorID(revisions []model.Revision) []model.UIDType {
	m := make(map[model.UIDType]bool, len(revisions))
	ret := make([]model.UIDType, len(revisions))
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

func convertModelPersonRevision(r *model.Revision, creatorMap map[model.UIDType]model.User) res.PersonRevision {
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
	if data, ok := r.Data.(map[string]model.PersonRevisionDataItem); ok {
		ret.Data = make(map[string]res.PersonRevisionDataItem, len(data))
		for id, item := range data {
			ret.Data[id] = res.PersonRevisionDataItem{
				InfoBox: item.InfoBox,
				Summary: item.Summary,
				Profession: res.Profession{
					Writer:      item.Profession.Writer,
					Producer:    item.Profession.Producer,
					Mangaka:     item.Profession.Mangaka,
					Artist:      item.Profession.Artist,
					Seiyu:       item.Profession.Seiyu,
					Illustrator: item.Profession.Illustrator,
					Actor:       item.Profession.Actor,
				},
				Extra: res.Extra{
					Img: item.Extra.Img,
				},
				Name: item.Name,
			}
		}
	}
	return ret
}

func convertModelSubjectRevision(r *model.Revision, creatorMap map[model.UIDType]model.User) res.SubjectRevision {
	creator := creatorMap[r.CreatorID]
	var data *res.SubjectRevisionData
	v := reflect.ValueNoEscapeOf(r.Data)
	if v.IsValid() && (!v.IsZero()) && (!v.IsNil()) {
		// can't compare r.Data != nil
		// see https://yourbasic.org/golang/gotcha-why-nil-error-not-equal-nil/ for this detail.
		// replace it with generic type after we upgrade to go 1.18
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

func convertModelCharacterRevision(
	r *model.CharacterRevision, creatorMap map[model.UIDType]model.User,
) res.CharacterRevision {
	creator := creatorMap[r.CreatorID]
	ret := res.CharacterRevision{
		ID:      r.ID,
		Type:    r.Type,
		Summary: r.Summary,
		Creator: res.Creator{
			Username: creator.UserName,
			Nickname: creator.UserName,
		},
		CreatedAt: r.CreatedAt,
	}
	ret.Data = make(map[string]res.CharacterRevisionDataItem, len(r.Data.CharacterRevisionEdit))
	for id, item := range r.Data.CharacterRevisionEdit {
		ret.Data[id] = res.CharacterRevisionDataItem{
			InfoBox: item.InfoBox,
			Summary: item.Summary,
			Extra: res.Extra{
				Img: item.Extra.Img,
			},
			Name: item.Name,
		}
	}
	return ret
}
