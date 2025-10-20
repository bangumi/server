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
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/pkg/vars"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Person) GetRelatedSubjects(c echo.Context) error {
	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	r, err := h.person.Get(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "failed to get person")
	}

	if r.Redirect != 0 {
		return res.ErrNotFound
	}

	relations, err := h.getPersonRelated(c.Request().Context(), id)
	if err != nil {
		return errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	var response = make([]res.PersonRelatedSubject, len(relations))
	for i, relation := range relations {
		response[i] = res.PersonRelatedSubject{
			SubjectID: relation.Subject.ID,
			Eps:       relation.Eps,
			Type:      relation.Subject.TypeID,
			Staff:     vars.StaffMap[relation.Subject.TypeID][relation.TypeID].String(),
			Name:      relation.Subject.Name,
			NameCn:    relation.Subject.NameCN,
			Image:     res.SubjectImage(relation.Subject.Image).Large,
		}
	}

	res.SetCacheControl(c, res.CacheControlParams{Public: true, MaxAge: time.Hour})

	return c.JSON(http.StatusOK, response)
}

func (h Person) getPersonRelated(
	ctx context.Context, personID model.PersonID,
) ([]model.SubjectPersonRelation, error) {
	person, err := h.person.Get(ctx, personID)
	if err != nil {
		return nil, errgo.Wrap(err, "failed to get user")
	}

	relations, err := h.subject.GetPersonRelated(ctx, personID)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	subjects, err := h.subject.GetByIDs(ctx,
		slice.Map(relations, func(r domain.SubjectPersonRelation) model.SubjectID { return r.SubjectID }),
		subject.Filter{})
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetByIDs")
	}

	var results = make([]model.SubjectPersonRelation, 0, len(relations))
	for _, rel := range relations {
		s, ok := subjects[rel.SubjectID]
		if !ok {
			continue
		}
		results = append(results, model.SubjectPersonRelation{
			Person:  person,
			Subject: s,
			TypeID:  rel.TypeID,
			Eps:     rel.Eps,
		})
	}

	return results, nil
}
