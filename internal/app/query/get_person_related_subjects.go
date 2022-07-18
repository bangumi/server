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

package query

import (
	"context"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func (q Query) GetPersonRelatedCharacters(
	ctx context.Context, personID model.PersonID,
) ([]model.PersonCharacterRelation, error) {
	relations, err := q.character.GetPersonRelated(ctx, personID)
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

	characters, err := q.character.GetByIDs(ctx, characterIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "CharacterRepo.GetByIDs")
	}

	subjects, err := q.subject.GetByIDs(ctx, subjectIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetByIDs")
	}

	person, err := q.person.Get(ctx, personID)
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
