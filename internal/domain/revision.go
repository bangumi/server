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

type RevisionRepo interface {
	CountPersonRelated(ctx context.Context, personID model.PersonID) (int64, error)

	ListPersonRelated(
		ctx context.Context, personID model.PersonID, limit int, offset int,
	) ([]model.PersonRevision, error)

	GetPersonRelated(ctx context.Context, id model.PersonID) (model.PersonRevision, error)

	CountSubjectRelated(ctx context.Context, id model.SubjectID) (int64, error)

	ListSubjectRelated(
		ctx context.Context, id model.SubjectID, limit int, offset int,
	) ([]model.SubjectRevision, error)

	GetSubjectRelated(ctx context.Context, id model.SubjectID) (model.SubjectRevision, error)

	CountCharacterRelated(ctx context.Context, characterID model.CharacterID) (int64, error)

	ListCharacterRelated(
		ctx context.Context, characterID model.CharacterID, limit int, offset int,
	) ([]model.CharacterRevision, error)

	GetCharacterRelated(ctx context.Context, id model.CharacterID) (model.CharacterRevision, error)
}
