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

package character

import (
	"context"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/internal/errgo"
	"github.com/bangumi/server/model"
)

func NewService(c domain.CharacterRepo) domain.CharacterService {
	return service{repo: c}
}

type service struct {
	repo domain.CharacterRepo
}

func (s service) Get(ctx context.Context, id uint32) (model.Character, error) {
	return s.repo.Get(ctx, id) //nolint:wrapcheck
}

func (s service) GetPersonRelated(
	ctx context.Context, personID model.PersonIDType,
) ([]model.PersonCharacterRelation, error) {
	relations, err := s.repo.GetPersonRelated(ctx, personID)
	if err != nil {
		return nil, errgo.Wrap(err, "CharacterRepo.GetPersonRelated")
	}

	var characterIDs = make([]model.CharacterIDType, len(relations))
	for i, relation := range relations {
		characterIDs[i] = relation.CharacterID
	}

	characters, err := s.repo.GetByIDs(ctx, characterIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "CharacterRepo.GetByIDs")
	}

	var results = make([]model.PersonCharacterRelation, len(relations))
	for i, rel := range relations {
		results[i].Character = characters[rel.CharacterID]
	}

	return results, nil
}

func (s service) GetSubjectRelated(
	ctx context.Context, subjectID model.SubjectIDType,
) ([]model.SubjectCharacterRelation, error) {
	relations, err := s.repo.GetSubjectRelated(ctx, subjectID)
	if err != nil {
		return nil, errgo.Wrap(err, "CharacterRepo.GetSubjectRelated")
	}

	var characterIDs = make([]model.CharacterIDType, len(relations))
	for i, relation := range relations {
		characterIDs[i] = relation.CharacterID
	}

	characters, err := s.repo.GetByIDs(ctx, characterIDs...)
	if err != nil {
		return nil, errgo.Wrap(err, "CharacterRepo.GetByIDs")
	}

	var results = make([]model.SubjectCharacterRelation, len(relations))
	for i, rel := range relations {
		results[i] = model.SubjectCharacterRelation{
			Character: characters[rel.CharacterID],
			TypeID:    rel.TypeID,
		}
	}

	return results, nil
}
