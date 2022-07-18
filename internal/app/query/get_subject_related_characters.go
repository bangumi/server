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

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
)

func (q Query) GetSubjectRelatedCharacters(
	ctx context.Context,
	user domain.Auth,
	subjectID model.SubjectID,
) (model.Subject, []model.SubjectCharacterRelation, error) {
	s, err := q.GetSubjectNoRedirect(ctx, user, subjectID)
	if err != nil {
		return s, nil, err
	}

	relations, err := q.character.GetSubjectRelated(ctx, subjectID)
	if err != nil {
		return s, nil, errgo.Wrap(err, "CharacterRepo.GetSubjectRelated")
	}

	var characterIDs = slice.Map(relations, func(item domain.SubjectCharacterRelation) model.CharacterID {
		return item.CharacterID
	})

	characters, err := q.character.GetByIDs(ctx, characterIDs...)
	if err != nil {
		return s, nil, errgo.Wrap(err, "CharacterRepo.GetByIDs")
	}

	var results = make([]model.SubjectCharacterRelation, len(relations))
	for i, rel := range relations {
		results[i] = model.SubjectCharacterRelation{
			Character: characters[rel.CharacterID],
			TypeID:    rel.TypeID,
		}
	}

	return s, results, nil
}
