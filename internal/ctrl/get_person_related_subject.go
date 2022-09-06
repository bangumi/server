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

	subjects, err := ctl.GetSubjectByIDs(ctx,
		slice.Map(relations, func(r domain.SubjectPersonRelation) model.SubjectID { return r.SubjectID }),
		SubjectFilter{})
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetByIDs")
	}

	var results = make([]model.SubjectPersonRelation, len(relations))
	for i, rel := range relations {
		results[i].Subject = subjects[rel.SubjectID]
		results[i].TypeID = rel.TypeID
		results[i].Person = person
	}

	return results, nil
}
