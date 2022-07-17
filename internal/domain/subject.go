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
	Get(ctx context.Context, id model.SubjectID) (model.Subject, error)
	GetByIDs(ctx context.Context, ids ...model.SubjectID) (map[model.SubjectID]model.Subject, error)

	GetPersonRelated(ctx context.Context, personID model.PersonID) ([]SubjectPersonRelation, error)
	GetCharacterRelated(ctx context.Context, characterID model.CharacterID) ([]SubjectCharacterRelation, error)
	GetSubjectRelated(ctx context.Context, subjectID model.SubjectID) ([]SubjectInternalRelation, error)

	GetActors(
		ctx context.Context, subjectID model.SubjectID, characterIDs ...model.CharacterID,
	) (map[model.CharacterID][]model.Person, error)
}

type SubjectService interface {
	// Get return a repository model.
	Get(ctx context.Context, id model.SubjectID) (model.Subject, error)
	GetByIDs(ctx context.Context, ids ...model.SubjectID) (map[model.SubjectID]model.Subject, error)

	GetCharacterRelated(ctx context.Context, characterID model.CharacterID) ([]model.SubjectCharacterRelation, error)
	GetSubjectRelated(ctx context.Context, subjectID model.SubjectID) ([]model.SubjectInternalRelation, error)

	GetActors(
		ctx context.Context, subjectID model.SubjectID, characterIDs ...model.CharacterID,
	) (map[model.CharacterID][]model.Person, error)
}

type SubjectPersonRelation struct {
	TypeID uint16

	PersonID  model.PersonID
	SubjectID model.SubjectID
}

func (r SubjectPersonRelation) GetSubjectID() model.SubjectID {
	return r.SubjectID
}

type SubjectCharacterRelation struct {
	TypeID uint8

	SubjectID   model.SubjectID
	CharacterID model.CharacterID
}

type SubjectInternalRelation struct {
	TypeID uint16

	SourceID      model.SubjectID
	DestinationID model.SubjectID
}
