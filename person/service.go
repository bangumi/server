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

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
)

func NewService(p domain.PersonRepo) domain.PersonService {
	return service{repo: p}
}

type service struct {
	repo domain.PersonRepo
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

	var subjectIDs = make([]model.SubjectIDType, len(relations))
	var results = make([]model.SubjectPersonRelation, len(relations))
	for i, relation := range relations {
		subjectIDs[i] = relation.SubjectID
	}

	characters, err := s.repo.GetByIDs(ctx, subjectIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "PersonRepo.GetByIDs")
	}

	for i, rel := range relations {
		results[i].Person = characters[rel.PersonID]
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

	var subjectIDs = make([]model.SubjectIDType, len(relations))
	var results = make([]model.PersonCharacterRelation, len(relations))
	for i, relation := range relations {
		subjectIDs[i] = relation.SubjectID
	}

	characters, err := s.repo.GetByIDs(ctx, subjectIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "PersonRepo.GetByIDs")
	}

	for i, rel := range relations {
		results[i].Person = characters[rel.PersonID]
	}

	return results, nil
}
