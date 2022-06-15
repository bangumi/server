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

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/internal/model"
)

func NewService(p domain.PersonRepo, s domain.SubjectRepo) domain.PersonService {
	return service{repo: p, s: s}
}

type service struct {
	repo domain.PersonRepo
	s    domain.SubjectRepo
}

func (s service) Get(ctx context.Context, id model.PersonID) (model.Person, error) {
	return s.repo.Get(ctx, id) //nolint:wrapcheck
}

func (s service) GetSubjectRelated(
	ctx context.Context, subjectID model.SubjectID,
) ([]model.SubjectPersonRelation, error) {
	relations, err := s.repo.GetSubjectRelated(ctx, subjectID)
	if err != nil {
		return nil, errgo.Wrap(err, "PersonRepo.GetSubjectRelated")
	}

	var personIDs = make([]model.PersonID, len(relations))
	for i, relation := range relations {
		personIDs[i] = relation.PersonID
	}

	persons, err := s.repo.GetByIDs(ctx, personIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "PersonRepo.GetByIDs")
	}

	subject, err := s.s.Get(ctx, subjectID)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetByID")
	}

	var results = make([]model.SubjectPersonRelation, len(relations))
	for i, rel := range relations {
		results[i] = model.SubjectPersonRelation{
			Person:  persons[rel.PersonID],
			Subject: subject,
			TypeID:  rel.TypeID,
		}
	}

	return results, nil
}

func (s service) GetCharacterRelated(
	ctx context.Context, characterID model.CharacterID,
) ([]model.PersonCharacterRelation, error) {
	relations, err := s.repo.GetCharacterRelated(ctx, characterID)
	if err != nil {
		return nil, errgo.Wrap(err, "PersonRepo.GetCharacterRelated")
	}

	var personIDs = make([]model.PersonID, len(relations))
	var subjectIDs = make([]model.SubjectID, len(relations))
	for i, relation := range relations {
		personIDs[i] = relation.PersonID
		subjectIDs[i] = relation.SubjectID
	}

	persons, err := s.repo.GetByIDs(ctx, personIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "PersonRepo.GetByIDs")
	}

	subjects, err := s.s.GetByIDs(ctx, subjectIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetByID")
	}

	var results = make([]model.PersonCharacterRelation, len(relations))
	for i, rel := range relations {
		results[i] = model.PersonCharacterRelation{
			Subject: subjects[rel.SubjectID],
			Person:  persons[rel.PersonID],
		}
	}

	return results, nil
}
