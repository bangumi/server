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
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/gstr"
	"github.com/bangumi/server/internal/user"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Handler) ListPersonRevision(c echo.Context) error {
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}
	personID, err := req.ParseID(c.QueryParam("person_id"))
	if err != nil {
		return err
	}

	return h.listPersonRevision(c, personID, page)
}

func (h Handler) listPersonRevision(c echo.Context, personID model.PersonID, page req.PageQuery) error {
	var response = res.Paged{
		Limit:  page.Limit,
		Offset: page.Offset,
	}
	count, err := h.r.CountPersonRelated(c.Request().Context(), personID)
	if err != nil {
		return errgo.Wrap(err, "revision.CountPersonRelated")
	}

	if count == 0 {
		response.Data = []int{}
		return c.JSON(http.StatusOK, response)
	}

	if err = page.Check(count); err != nil {
		return err
	}

	response.Total = count

	revisions, err := h.r.ListPersonRelated(c.Request().Context(), personID, page.Limit, page.Offset)
	if err != nil {
		return errgo.Wrap(err, "revision.ListPersonRelated")
	}

	data := make([]res.PersonRevision, len(revisions))

	creatorIDs := make([]model.UserID, 0, len(revisions))
	for _, revision := range revisions {
		creatorIDs = append(creatorIDs, revision.CreatorID)
	}

	creatorMap, err := h.u.GetByIDs(c.Request().Context(), lo.Uniq(creatorIDs))
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	for i := range revisions {
		data[i] = convertModelPersonRevision(&revisions[i], creatorMap)
	}
	response.Data = data

	return c.JSON(http.StatusOK, response)
}

func (h Handler) GetPersonRevision(c echo.Context) error {
	id, err := gstr.ParseUint32(c.Param("id"))
	if err != nil || id <= 0 {
		return res.BadRequest(fmt.Sprintf("bad param id: %s", c.Param("id")))
	}
	r, err := h.r.GetPersonRelated(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get person related revision")
	}

	creatorMap, err := h.u.GetByIDs(c.Request().Context(), []model.UserID{r.CreatorID})
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	return c.JSON(http.StatusOK, convertModelPersonRevision(&r, creatorMap))
}

func (h Handler) ListCharacterRevision(c echo.Context) error {
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}

	characterID, err := req.ParseID(c.QueryParam("character_id"))
	if err != nil {
		return err
	}

	return h.listCharacterRevision(c, characterID, page)
}

func (h Handler) listCharacterRevision(c echo.Context, characterID model.CharacterID, page req.PageQuery) error {
	var response = res.Paged{
		Limit:  page.Limit,
		Offset: page.Offset,
	}
	count, err := h.r.CountCharacterRelated(c.Request().Context(), characterID)
	if err != nil {
		return errgo.Wrap(err, "revision.CountCharacterRelated")
	}

	if count == 0 {
		response.Data = []int{}
		return c.JSON(http.StatusOK, response)
	}

	if err = page.Check(count); err != nil {
		return err
	}

	response.Total = count

	revisions, err := h.r.ListCharacterRelated(c.Request().Context(), characterID, page.Limit, page.Offset)

	if err != nil {
		return errgo.Wrap(err, "revision.ListCharacterRelated")
	}

	creatorIDs := make([]model.UserID, 0, len(revisions))
	for _, revision := range revisions {
		creatorIDs = append(creatorIDs, revision.CreatorID)
	}
	creatorMap, err := h.u.GetByIDs(c.Request().Context(), lo.Uniq(creatorIDs))

	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	data := make([]res.CharacterRevision, len(revisions))
	for i := range revisions {
		data[i] = convertModelCharacterRevision(&revisions[i], creatorMap)
	}
	response.Data = data
	return c.JSON(http.StatusOK, response)
}

func (h Handler) GetCharacterRevision(c echo.Context) error {
	id, err := gstr.ParseUint32(c.Param("id"))
	if err != nil || id <= 0 {
		return res.NewError(
			http.StatusBadRequest,
			fmt.Sprintf("bad param id: %s", strconv.Quote(c.Param("id"))),
		)
	}
	r, err := h.r.GetCharacterRelated(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get character related revision")
	}

	creatorMap, err := h.u.GetByIDs(c.Request().Context(), []model.UserID{r.CreatorID})
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	return c.JSON(http.StatusOK, convertModelCharacterRevision(&r, creatorMap))
}

func (h Handler) ListSubjectRevision(c echo.Context) error {
	page, err := req.GetPageQuery(c, req.DefaultPageLimit, req.DefaultMaxPageLimit)
	if err != nil {
		return err
	}
	subjectID, err := req.ParseID(c.QueryParam("subject_id"))
	if err != nil {
		return err
	}

	return h.listSubjectRevision(c, subjectID, page)
}

func (h Handler) listSubjectRevision(c echo.Context, subjectID model.SubjectID, page req.PageQuery) error {
	var response = res.Paged{
		Limit:  page.Limit,
		Offset: page.Offset,
	}
	count, err := h.r.CountSubjectRelated(c.Request().Context(), subjectID)
	if err != nil {
		return errgo.Wrap(err, "revision.CountSubjectRelated")
	}

	if count == 0 {
		response.Data = []int{}
		return c.JSON(http.StatusOK, response)
	}

	if err = page.Check(count); err != nil {
		return err
	}

	response.Total = count

	revisions, err := h.r.ListSubjectRelated(c.Request().Context(), subjectID, page.Limit, page.Offset)

	if err != nil {
		return errgo.Wrap(err, "revision.ListSubjectRelated")
	}

	data := make([]res.SubjectRevision, len(revisions))

	creatorIDs := make([]model.UserID, 0, len(revisions))
	for _, revision := range revisions {
		creatorIDs = append(creatorIDs, revision.CreatorID)
	}
	creatorMap, err := h.u.GetByIDs(c.Request().Context(), lo.Uniq(creatorIDs))

	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}
	for i := range revisions {
		data[i] = convertModelSubjectRevision(&revisions[i], creatorMap)
	}
	response.Data = data
	return c.JSON(http.StatusOK, response)
}

func (h Handler) GetSubjectRevision(c echo.Context) error {
	id, err := gstr.ParseUint32(c.Param("id"))
	if err != nil || id == 0 {
		return res.BadRequest("bad param id: " + strconv.Quote(c.Param("id")))
	}
	r, err := h.r.GetSubjectRelated(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get subject related revision")
	}

	creatorMap, err := h.u.GetByIDs(c.Request().Context(), []model.UserID{r.CreatorID})
	if err != nil {
		return errgo.Wrap(err, "user.GetByIDs")
	}

	return c.JSON(http.StatusOK, convertModelSubjectRevision(&r, creatorMap))
}

func convertModelPersonRevision(r *model.PersonRevision, creatorMap map[model.UserID]user.User) res.PersonRevision {
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
	ret.Data = make(map[string]res.PersonRevisionDataItem, len(r.Data))
	for id, item := range r.Data {
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

	return ret
}

func convertModelSubjectRevision(
	r *model.SubjectRevision, creatorMap map[model.UserID]user.User,
) res.SubjectRevision {
	creator := creatorMap[r.CreatorID]
	var data *res.SubjectRevisionData
	if r.Data != nil {
		subjectData := r.Data
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
	r *model.CharacterRevision, creatorMap map[model.UserID]user.User,
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
	ret.Data = make(map[string]res.CharacterRevisionDataItem, len(r.Data))
	for id, item := range r.Data {
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

func convertModelEpisodeRevision(r *model.EpisodeRevision, creatorMap map[model.UserID]user.User) res.EpisodeRevision {
	creator := creatorMap[r.CreatorID]
	ret := res.EpisodeRevision{
		ID:      r.ID,
		Type:    r.Type,
		Summary: r.Summary,
		Creator: res.Creator{
			Username: creator.UserName,
			Nickname: creator.UserName,
		},
		CreatedAt: r.CreatedAt,
	}
	ret.Data = make(map[string]res.EpisodeRevisionDataItem, len(r.Data))
	for id, item := range r.Data {
		ret.Data[id] = res.EpisodeRevisionDataItem(item)
	}
	return ret
}
