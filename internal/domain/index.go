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
	"time"

	"github.com/bangumi/server/internal/model"
)

type IndexRepo interface {
	Get(ctx context.Context, id model.IndexID) (model.Index, error)

	CountSubjects(ctx context.Context, id model.IndexID, subjectType model.SubjectType) (int64, error)
	ListSubjects(
		ctx context.Context, id model.IndexID, subjectType model.SubjectType, limit, offset int,
	) ([]IndexSubject, error)
}

type IndexSubject struct {
	Comment string
	AddedAt time.Time
	Subject model.Subject
}
