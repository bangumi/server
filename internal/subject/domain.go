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
	"github.com/bangumi/server/internal/pkg/null"
)

type Filter struct {
	NSFW null.Bool
}

type Repo interface {
	// Get return a repository model.
	Get(ctx context.Context, id model.SubjectID, filter Filter) (model.Subject, error)
	GetByIDs(ctx context.Context, ids []model.SubjectID, filter Filter) (map[model.SubjectID]model.Subject, error)

	GetPersonRelated(ctx context.Context, personID model.PersonID) ([]domain.SubjectPersonRelation, error)
	GetCharacterRelated(ctx context.Context, characterID model.CharacterID) ([]domain.SubjectCharacterRelation, error)
	GetSubjectRelated(ctx context.Context, subjectID model.SubjectID) ([]domain.SubjectInternalRelation, error)

	GetActors(
		ctx context.Context, subjectID model.SubjectID, characterIDs ...model.CharacterID,
	) (map[model.CharacterID][]model.PersonID, error)
}
