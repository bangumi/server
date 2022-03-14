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

package person

import (
	"context"
	"fmt"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
)

func NewService(p domain.PersonRepo, s domain.SubjectRepo) domain.PersonService {
	return service{repo: p, s: s}
}

type service struct {
	repo domain.PersonRepo
	s    domain.SubjectRepo
}

func (s service) Get(ctx context.Context, id uint32) (model.Person, error) {
	return s.repo.Get(ctx, id) //nolint:wrapcheck
}

func (s service) GetSubjectRelated(
	ctx context.Context, subjectID model.SubjectIDType,
) ([]model.SubjectPersonRelation, error) {
	relations, err := s.repo.GetSubjectRelated(ctx, subjectID)
	if err != nil {
		return nil, errgo.Wrap(err, "PersonRepo.GetSubjectRelated")
	}

	var personIDs = make([]model.PersonIDType, len(relations))
	for i, relation := range relations {
		personIDs[i] = relation.PersonID
	}

	persons, err := s.repo.GetByIDs(ctx, personIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "PersonRepo.GetByIDs")
	}

	var results = make([]model.SubjectPersonRelation, len(relations))
	for i, rel := range relations {
		results[i].Person = persons[rel.PersonID]
	}

	return results, nil
}

func (s service) GetCharacterRelated(
	ctx context.Context, subjectID model.CharacterIDType,
) ([]model.PersonCharacterRelation, error) {
	relations, err := s.repo.GetCharacterRelated(ctx, subjectID)
	if err != nil {
		return nil, errgo.Wrap(err, "PersonRepo.GetCharacterRelated")
	}

	var personIDs = make([]model.PersonIDType, len(relations))
	for i, relation := range relations {
		personIDs[i] = relation.PersonID
	}

	fmt.Println(personIDs)

	persons, err := s.repo.GetByIDs(ctx, personIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "PersonRepo.GetByIDs")
	}

	fmt.Println(persons)

	subject, err := s.s.Get(ctx, subjectID)
	if err != nil {
		return nil, errgo.Wrap(err, "SubjectRepo.Get")
	}

	var results = make([]model.PersonCharacterRelation, len(relations))
	for i, rel := range relations {
		results[i] = model.PersonCharacterRelation{
			Subject: subject,
			Person:  persons[rel.PersonID],
			TypeID:  rel.TypeID,
		}
	}

	return results, nil
}
