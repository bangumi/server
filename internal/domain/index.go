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
	New(ctx context.Context, i *model.Index) error
	Update(ctx context.Context, id model.IndexID, title string, desc string) error
	Delete(ctx context.Context, id model.IndexID) error

	CountSubjects(ctx context.Context, id model.IndexID, subjectType model.SubjectType) (int64, error)
	ListSubjects(
		ctx context.Context, id model.IndexID, subjectType model.SubjectType, limit, offset int,
	) ([]IndexSubject, error)
	AddIndexSubject(
		ctx context.Context, id model.IndexID, subjectID model.SubjectID, sort uint32, comment string,
	) (*IndexSubject, error)
	UpdateIndexSubject(
		ctx context.Context, id model.IndexID, subjectID model.SubjectID, sort uint32, comment string,
	) error
	DeleteIndexSubject(
		ctx context.Context, id model.IndexID, subjectID model.SubjectID,
	) error

	GetIndicesByUser(
		ctx context.Context, creatorID model.UserID, limit int, offset int,
	) ([]model.Index, error)
	GetCollectedIndicesByUser(
		ctx context.Context, creatorID model.UserID, limit int, offset int,
	) ([]model.IndexCollect, error)
}

type IndexSubject struct {
	Comment string
	AddedAt time.Time
	Subject model.Subject
}
