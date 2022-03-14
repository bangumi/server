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

package domain

import (
	"context"

	"github.com/bangumi/server/model"
)

type CharacterRepo interface {
	Get(ctx context.Context, id model.CharacterIDType) (model.Character, error)
	GetByIDs(ctx context.Context, ids ...model.CharacterIDType) (map[model.CharacterIDType]model.Character, error)

	GetPersonRelated(ctx context.Context, characterID model.PersonIDType) ([]PersonCharacterRelation, error)
	GetSubjectRelated(ctx context.Context, subjectID model.SubjectIDType) ([]SubjectCharacterRelation, error)
}

type CharacterService interface {
	Get(ctx context.Context, id uint32) (model.Character, error)

	GetPersonRelated(ctx context.Context, personID model.PersonIDType) ([]model.PersonCharacterRelation, error)
	GetSubjectRelated(ctx context.Context, subjectID model.SubjectIDType) ([]model.SubjectCharacterRelation, error)
}

type PersonCharacterRelation struct {
	CharacterID model.CharacterIDType
	PersonID    model.PersonIDType
	SubjectID   model.SubjectIDType
	TypeID      uint8
}
