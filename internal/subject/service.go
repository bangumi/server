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

package subject

import (
	"context"

	"github.com/bangumi/server/internal/domain"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/errgo"
)

func NewService(s domain.SubjectRepo, p domain.PersonRepo) domain.SubjectService {
	return service{repo: s}
}

type service struct {
	repo domain.SubjectRepo
}

func (s service) Get(ctx context.Context, id model.SubjectID) (model.Subject, error) {
	return s.repo.Get(ctx, id) //nolint:wrapcheck
}

func (s service) GetByIDs(ctx context.Context, ids ...model.SubjectID) (map[model.SubjectID]model.Subject, error) {
	return s.repo.GetByIDs(ctx, ids...) //nolint:wrapcheck
}

func (s service) GetPersonRelated(
	ctx context.Context, personID model.PersonID,
) ([]model.SubjectPersonRelation, error) {
	relations, err := s.repo.GetPersonRelated(ctx, personID)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	var subjectIDs = make([]model.SubjectID, len(relations))
	var results = make([]model.SubjectPersonRelation, len(relations))
	for i, relation := range relations {
		subjectIDs[i] = relation.SubjectID
	}

	subjects, err := s.repo.GetByIDs(ctx, subjectIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetByIDs")
	}

	for i, rel := range relations {
		results[i].Subject = subjects[rel.SubjectID]
		results[i].TypeID = rel.TypeID
	}

	return results, nil
}

func (s service) GetCharacterRelated(
	ctx context.Context, characterID model.CharacterID,
) ([]model.SubjectCharacterRelation, error) {
	relations, err := s.repo.GetCharacterRelated(ctx, characterID)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetCharacterRelated")
	}

	var subjectIDs = make([]model.SubjectID, len(relations))
	for i, relation := range relations {
		subjectIDs[i] = relation.SubjectID
	}

	subjects, err := s.repo.GetByIDs(ctx, subjectIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetByIDs")
	}

	var results = make([]model.SubjectCharacterRelation, len(relations))
	for i, rel := range relations {
		results[i] = model.SubjectCharacterRelation{
			Subject: subjects[rel.SubjectID],
			TypeID:  rel.TypeID,
		}
	}

	return results, nil
}

func (s service) GetSubjectRelated(
	ctx context.Context, subjectID model.SubjectID,
) ([]model.SubjectInternalRelation, error) {
	relations, err := s.repo.GetSubjectRelated(ctx, subjectID)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetSubjectRelated")
	}

	var subjectIDs = make([]model.SubjectID, len(relations))
	var results = make([]model.SubjectInternalRelation, len(relations))
	for i, relation := range relations {
		subjectIDs[i] = relation.DestinationID
	}

	subjects, err := s.repo.GetByIDs(ctx, subjectIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetByIDs")
	}

	for i, rel := range relations {
		results[i].Destination = subjects[rel.DestinationID]
		results[i].TypeID = rel.TypeID
	}

	return results, nil
}

func (s service) GetActors(
	ctx context.Context, subjectID model.SubjectID, characterIDs ...model.CharacterID,
) (map[model.CharacterID][]model.Person, error) {
	return s.repo.GetActors(ctx, subjectID, characterIDs...) //nolint:wrapcheck
}
