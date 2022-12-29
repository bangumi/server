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
	"github.com/bangumi/server/internal/model"
)

type Repo interface {
	Get(ctx context.Context, id model.PersonID) (model.Person, error)
	GetByIDs(ctx context.Context, ids []model.PersonID) (map[model.PersonID]model.Person, error)

	GetSubjectRelated(ctx context.Context, subjectID model.SubjectID) ([]domain.SubjectPersonRelation, error)
	GetCharacterRelated(ctx context.Context, subjectID model.CharacterID) ([]domain.PersonCharacterRelation, error)
}

type Service interface {
	Get(ctx context.Context, id model.PersonID) (model.Person, error)

	GetSubjectRelated(ctx context.Context, subjectID model.SubjectID) ([]model.SubjectPersonRelation, error)
	GetCharacterRelated(ctx context.Context, characterID model.CharacterID) ([]model.PersonCharacterRelation, error)
}
