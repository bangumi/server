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
)

func NewService(s domain.SubjectRepo, p domain.PersonRepo) domain.SubjectService {
	return service{repo: s}
}

type service struct {
	repo domain.SubjectRepo
}

func (s service) Get(ctx context.Context, id model.SubjectID) (model.Subject, error) {
	return s.repo.Get(ctx, id) //nolint:wrapcheck
}

func (s service) GetByIDs(ctx context.Context, ids ...model.SubjectID) (map[model.SubjectID]model.Subject, error) {
	return s.repo.GetByIDs(ctx, ids...) //nolint:wrapcheck
}

func (s service) GetActors(
	ctx context.Context, subjectID model.SubjectID, characterIDs ...model.CharacterID,
) (map[model.CharacterID][]model.Person, error) {
	return s.repo.GetActors(ctx, subjectID, characterIDs...) //nolint:wrapcheck
}
