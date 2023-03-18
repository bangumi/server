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

package index

import (
	"context"
	"time"

	"github.com/bangumi/server/internal/model"
)

type Repo interface {
	Get(ctx context.Context, id model.IndexID) (model.Index, error)
	New(ctx context.Context, i *model.Index) error
	Update(ctx context.Context, id model.IndexID, title string, desc string) error
	Delete(ctx context.Context, id model.IndexID) error

	CountSubjects(ctx context.Context, id model.IndexID, subjectType model.SubjectType) (int64, error)
	ListSubjects(
		ctx context.Context, id model.IndexID, subjectType model.SubjectType, limit, offset int,
	) ([]Subject, error)
	AddOrUpdateIndexSubject(
		ctx context.Context, id model.IndexID, subjectID model.SubjectID, sort uint32, comment string,
	) (*Subject, error)
	DeleteIndexSubject(
		ctx context.Context, id model.IndexID, subjectID model.SubjectID,
	) error

	// AddIndexCollect add an index collect to given user
	AddIndexCollect(ctx context.Context, id model.IndexID, uid model.UserID) error

	// DeleteIndexCollect remove index collect from given user
	DeleteIndexCollect(ctx context.Context, id model.IndexID, uid model.UserID) error
}

type Subject struct {
	Comment string
	AddedAt time.Time
	Subject model.Subject
}
