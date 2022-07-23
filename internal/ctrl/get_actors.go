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

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/gmap"
	"github.com/bangumi/server/internal/pkg/generic/slice"
)

func (ctl Ctrl) GetActors(
	ctx context.Context,
	subjectID model.SubjectID,
	characterIDs ...model.CharacterID,
) (map[model.CharacterID][]model.Person, error) {
	actors, err := ctl.subject.GetActors(ctx, subjectID, characterIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "subjectRepo.GetActors")
	}

	vs := slice.Unique(slice.Reduce(gmap.Values(actors)))

	persons, err := ctl.person.GetByIDs(ctx, vs...)
	if err != nil {
		return nil, errgo.Wrap(err, "failed to get persons")
	}

	var result = make(map[model.CharacterID][]model.Person, len(actors))

	for characterID, ids := range actors {
		result[characterID] = slice.Map(ids, func(item model.PersonID) model.Person {
			return persons[item]
		})
	}

	return result, nil
}
