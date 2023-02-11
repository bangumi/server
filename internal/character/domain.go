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
	"github.com/bangumi/server/internal/model"
)

type Repo interface {
	Get(ctx context.Context, id model.CharacterID) (model.Character, error)
	GetByIDs(ctx context.Context, ids []model.CharacterID) (map[model.CharacterID]model.Character, error)

	GetPersonRelated(ctx context.Context, personID model.PersonID) ([]domain.PersonCharacterRelation, error)
	GetSubjectRelated(ctx context.Context, subjectID model.SubjectID) ([]domain.SubjectCharacterRelation, error)

	GetSubjectRelationByIDs(ctx context.Context, ids []SubjectCompositeId) ([]domain.SubjectCharacterRelation, error)
}

type SubjectCompositeId struct {
	CharacterID model.CharacterID
	SubjectID   model.SubjectID
}
