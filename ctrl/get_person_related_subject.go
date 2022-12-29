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

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/subject"
)

func (ctl Ctrl) GetPersonRelated(
	ctx context.Context, personID model.PersonID,
) ([]model.SubjectPersonRelation, error) {
	person, err := ctl.GetPerson(ctx, personID)
	if err != nil {
		return nil, err
	}

	relations, err := ctl.subject.GetPersonRelated(ctx, personID)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	subjects, err := ctl.subject.GetByIDs(ctx,
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
		})
	}

	return results, nil
}
