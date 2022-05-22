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

	"github.com/bangumi/server/internal/model"
)

type SubjectRepo interface {
	// Get return a repository model.
	Get(ctx context.Context, id uint32) (model.Subject, error)
	GetByIDs(ctx context.Context, ids ...model.SubjectIDType) (map[model.SubjectIDType]model.Subject, error)

	GetPersonRelated(ctx context.Context, personID model.PersonIDType) ([]SubjectPersonRelation, error)
	GetCharacterRelated(ctx context.Context, characterID model.PersonIDType) ([]SubjectCharacterRelation, error)
	GetSubjectRelated(ctx context.Context, subjectID model.SubjectIDType) ([]SubjectInternalRelation, error)

	GetActors(
		ctx context.Context, subjectID model.SubjectIDType, characterIDs ...model.CharacterIDType,
	) (map[model.CharacterIDType][]model.Person, error)
}

type SubjectService interface {
	// Get return a repository model.
	Get(ctx context.Context, id uint32) (model.Subject, error)

	GetPersonRelated(ctx context.Context, personID model.PersonIDType) ([]model.SubjectPersonRelation, error)
	GetCharacterRelated(ctx context.Context, characterID model.PersonIDType) ([]model.SubjectCharacterRelation, error)
	GetSubjectRelated(ctx context.Context, subjectID model.SubjectIDType) ([]model.SubjectInternalRelation, error)

	GetActors(
		ctx context.Context, subjectID model.SubjectIDType, characterIDs ...model.CharacterIDType,
	) (map[model.CharacterIDType][]model.Person, error)
}

type SubjectPersonRelation struct {
	TypeID uint16

	PersonID  model.PersonIDType
	SubjectID model.SubjectIDType
}

type SubjectCharacterRelation struct {
	TypeID uint8

	SubjectID   model.SubjectIDType
	CharacterID model.CharacterIDType
}

type SubjectInternalRelation struct {
	TypeID uint16

	SourceID      model.SubjectIDType
	DestinationID model.SubjectIDType
}
