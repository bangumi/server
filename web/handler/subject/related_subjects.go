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

package subject

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/pkg/null"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/accessor"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Subject) GetRelatedSubjects(c echo.Context) error {
	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	u := accessor.GetFromCtx(c)

	relations, err := h.getSubjectRelatedSubjects(c.Request().Context(), u.Auth, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return res.ErrNotFound
		}

		return errgo.Wrap(err, "repo")
	}

	var response = make([]res.SubjectRelatedSubject, len(relations))
	for i, relation := range relations {
		response[i] = res.SubjectRelatedSubject{
			Images:    res.SubjectImage(relation.Destination.Image),
			Name:      relation.Destination.Name,
			NameCn:    relation.Destination.NameCN,
			Relation:  readableRelation(relation.Destination.TypeID, relation.TypeID),
			Type:      relation.Destination.TypeID,
			SubjectID: relation.Destination.ID,
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (h Subject) getSubjectRelatedSubjects(
	ctx context.Context,
	user auth.Auth,
	subjectID model.SubjectID,
) ([]model.SubjectInternalRelation, error) {
	currentSubject, err := h.subject.Get(ctx, subjectID, subject.Filter{NSFW: null.Bool{
		Value: false, Set: !user.AllowNSFW(),
	}})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, res.ErrNotFound
		}
		return nil, errgo.Wrap(err, "failed to get subject")
	}

	relations, err := h.subject.GetSubjectRelated(ctx, subjectID)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetSubjectRelated")
	}

	subjects, err := h.subject.GetByIDs(ctx,
		slice.Map(relations, func(r domain.SubjectInternalRelation) model.SubjectID { return r.DestinationID }),
		subject.Filter{NSFW: null.Bool{Value: false, Set: !user.AllowNSFW()}},
	)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetByIDs")
	}

	var results = make([]model.SubjectInternalRelation, 0, len(relations))
	for _, rel := range relations {
		s, ok := subjects[rel.DestinationID]
		if !ok {
			continue
		}

		results = append(results, model.SubjectInternalRelation{
			Destination: s,
			TypeID:      rel.TypeID,
			Source:      currentSubject,
		})
	}

	return results, nil
}
