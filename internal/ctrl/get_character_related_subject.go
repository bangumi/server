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

package ctrl

import (
	"context"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/subject"
)

func (ctl Ctrl) GetCharacterRelatedSubjects(
	ctx context.Context,
	user domain.Auth,
	characterID model.CharacterID,
) (model.Character, []model.SubjectCharacterRelation, error) {
	character, err := ctl.GetCharacter(ctx, user, characterID)
	if err != nil {
		return model.Character{}, nil, err
	}

	if character.Redirect != 0 {
		return model.Character{}, nil, domain.ErrCharacterNotFound
	}

	relations, err := ctl.subject.GetCharacterRelated(ctx, characterID)
	if err != nil {
		return model.Character{}, nil, errgo.Wrap(err, "SubjectRepo.GetCharacterRelated")
	}

	var subjectIDs = slice.Map(relations, func(item domain.SubjectCharacterRelation) model.SubjectID {
		return item.SubjectID
	})

	subjects, err := ctl.subject.GetByIDs(ctx, subjectIDs, subject.Filter{})
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
