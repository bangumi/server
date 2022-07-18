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

func (q Query) GetSubjectRelatedSubjects(
	ctx context.Context,
	user domain.Auth,
	subjectID model.SubjectID,
) (model.Subject, []model.SubjectInternalRelation, error) {
	s, err := q.GetSubjectNoRedirect(ctx, user, subjectID)
	if err != nil {
		return model.Subject{}, nil, err
	}

	relations, err := q.subject.GetSubjectRelated(ctx, subjectID)
	if err != nil {
		return s, nil, errgo.Wrap(err, "SubjectRepo.GetSubjectRelated")
	}

	subjects, err := q.GetSubjectByIDs(ctx, slice.Map(relations, domain.SubjectInternalRelation.GetDestinationID)...)
	if err != nil {
		return s, nil, errgo.Wrap(err, "SubjectRepo.GetByIDs")
	}

	var results = make([]model.SubjectInternalRelation, len(relations))
	for i, rel := range relations {
		results[i] = model.SubjectInternalRelation{
			Destination: subjects[rel.DestinationID],
			TypeID:      rel.TypeID,
			Source:      s,
		}
	}

	return s, results, nil
}
