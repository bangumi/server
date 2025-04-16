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

	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/character"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/subject"
	"github.com/bangumi/server/web/req"
	"github.com/bangumi/server/web/res"
)

func (h Person) GetRelatedCharacters(c echo.Context) error {
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

	relations, err := h.getPersonRelatedCharacters(c.Request().Context(), id)
	if err != nil {
		return errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	var compositeIDs = make([]character.SubjectCompositeID, len(relations))
	for i, relation := range relations {
		compositeIDs[i] = character.SubjectCompositeID{
			CharacterID: relation.Character.ID,
			SubjectID:   relation.Subject.ID,
		}
	}
	subjectRelations, err := h.character.GetSubjectRelationByIDs(c.Request().Context(), compositeIDs)
	if err != nil {
		return errgo.Wrap(err, "CharacterRepo.GetRelations")
	}
	var mSubjectRelations = make(map[model.CharacterID]map[model.SubjectID]uint8)
	for _, rel := range subjectRelations {
		if mSubjectRelations[rel.CharacterID] == nil {
			mSubjectRelations[rel.CharacterID] = make(map[model.SubjectID]uint8)
		}
		mSubjectRelations[rel.CharacterID][rel.SubjectID] = rel.TypeID
	}

	var response = make([]res.PersonRelatedCharacter, len(relations))
	for i, rel := range relations {
		var subjectTypeID uint8
		if m2 := mSubjectRelations[rel.Character.ID]; m2 != nil {
			subjectTypeID = m2[rel.Subject.ID]
		}
		response[i] = res.PersonRelatedCharacter{
			ID:            rel.Character.ID,
			Name:          rel.Character.Name,
			Type:          rel.Character.Type,
			Images:        res.PersonImage(rel.Character.Image),
			SubjectID:     rel.Subject.ID,
			SubjectType:   rel.Subject.TypeID,
			SubjectName:   rel.Subject.Name,
			SubjectNameCn: rel.Subject.NameCN,
			Staff:         res.CharacterStaffString(subjectTypeID),
		}
	}

	res.SetCacheControl(c, res.CacheControlParams{Public: true, MaxAge: time.Hour})

	return c.JSON(http.StatusOK, response)
}

func (h Person) getPersonRelatedCharacters(
	ctx context.Context, personID model.PersonID,
) ([]model.PersonCharacterRelation, error) {
	relations, err := h.character.GetPersonRelated(ctx, personID)
	if err != nil {
		return nil, errgo.Wrap(err, "CharacterRepo.GetPersonRelated")
	}

	if len(relations) == 0 {
		return []model.PersonCharacterRelation{}, nil
	}

	var characterIDs = make([]model.CharacterID, len(relations))
	var subjectIDs = make([]model.SubjectID, len(relations))
	for i, relation := range relations {
		characterIDs[i] = relation.CharacterID
		subjectIDs[i] = relation.SubjectID
	}

	characters, err := h.character.GetByIDs(ctx, characterIDs)
	if err != nil {
		return nil, errgo.Wrap(err, "CharacterRepo.GetByIDs")
	}

	subjects, err := h.subject.GetByIDs(ctx, subjectIDs, subject.Filter{})
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetByIDs")
	}

	person, err := h.person.Get(ctx, personID)
	if err != nil {
		return nil, errgo.Wrap(err, "PersonRepo.GetByIDs")
	}

	var results = make([]model.PersonCharacterRelation, len(relations))
	for i, rel := range relations {
		results[i] = model.PersonCharacterRelation{
			Character: characters[rel.CharacterID],
			Person:    person,
			Subject:   subjects[rel.SubjectID],
		}
	}

	return results, nil
}
