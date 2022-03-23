// Copyright (c) 2022 Trim21 <trim21.me@gmail.com>
//
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

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
)

func NewService(s domain.SubjectRepo, p domain.PersonRepo) domain.SubjectService {
	return service{repo: s}
}

type service struct {
	repo domain.SubjectRepo
}

func (s service) Get(ctx context.Context, id uint32) (model.Subject, error) {
	return s.repo.Get(ctx, id) //nolint:wrapcheck
}

func (s service) GetPersonRelated(
	ctx context.Context, personID model.PersonIDType,
) ([]model.SubjectPersonRelation, error) {
	relations, err := s.repo.GetPersonRelated(ctx, personID)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetPersonRelated")
	}

	var subjectIDs = make([]model.SubjectIDType, len(relations))
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
	ctx context.Context, characterID model.PersonIDType,
) ([]model.SubjectCharacterRelation, error) {
	relations, err := s.repo.GetCharacterRelated(ctx, characterID)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetCharacterRelated")
	}

	var subjectIDs = make([]model.SubjectIDType, len(relations))
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
	ctx context.Context, subjectID model.SubjectIDType,
) ([]model.SubjectInternalRelation, error) {
	relations, err := s.repo.GetSubjectRelated(ctx, subjectID)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.GetSubjectRelated")
	}

	var subjectIDs = make([]model.SubjectIDType, len(relations))
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
	ctx context.Context, subjectID model.SubjectIDType, characterIDs ...model.CharacterIDType,
) (map[model.CharacterIDType][]model.Person, error) {
	return s.repo.GetActors(ctx, subjectID, characterIDs...) //nolint:wrapcheck
}
