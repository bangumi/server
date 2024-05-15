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

package character

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/trim21/errgo"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Character) GetRelatedSubjects(c echo.Context) error {
	id, err := req.ParseID(c.Param("id"))
	if err != nil {
		return err
	}

	_, relations, err := h.getCharacterRelatedSubjects(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, gerr.ErrNotFound) {
			return res.ErrNotFound
		}
		return errgo.Wrap(err, "repo.GetCharacterRelated")
	}

	var response = make([]res.CharacterRelatedSubject, len(relations))
	for i, relation := range relations {
		s := relation.Subject
		response[i] = res.CharacterRelatedSubject{
			ID:     s.ID,
			Type:   s.TypeID,
			Name:   s.Name,
			NameCn: s.NameCN,
			Staff:  res.CharacterStaffString(relation.TypeID),
			Image:  res.SubjectImage(s.Image).Large,
		}
	}

	return c.JSON(http.StatusOK, response)
}

func (h Character) getCharacterRelatedSubjects(
	ctx context.Context,
	characterID model.CharacterID,
) (model.Character, []model.SubjectCharacterRelation, error) {
	character, err := h.c.Get(ctx, characterID)
	if err != nil {
		return model.Character{}, nil, errgo.Wrap(err, "get character")
	}

	if character.Redirect != 0 {
		return model.Character{}, nil, gerr.ErrCharacterNotFound
	}

	relations, err := h.subject.GetCharacterRelated(ctx, characterID)
	if err != nil {
		return model.Character{}, nil, errgo.Wrap(err, "SubjectRepo.GetCharacterRelated")
	}

	var subjectIDs = slice.Map(relations, func(item domain.SubjectCharacterRelation) model.SubjectID {
		return item.SubjectID
	})

	subjects, err := h.subject.GetByIDs(ctx, subjectIDs, subject.Filter{})
	if err != nil {
		return model.Character{}, nil, errgo.Wrap(err, "SubjectRepo.GetByIDs")
	}

	var results = make([]model.SubjectCharacterRelation, 0, len(relations))
	for _, rel := range relations {
		s, ok := subjects[rel.SubjectID]
		if !ok {
			continue
		}
		results = append(results, model.SubjectCharacterRelation{
			Subject:   s,
			TypeID:    rel.TypeID,
			Character: character,
		})
	}

	return character, results, nil
}
