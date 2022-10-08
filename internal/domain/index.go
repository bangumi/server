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

// 和目录直接相关数据库操作.
type indexRepo interface {
	Get(ctx context.Context, id model.IndexID) (model.Index, error)
	New(ctx context.Context, i *model.Index) error
	Update(ctx context.Context, id model.IndexID, title string, desc string) error
	Delete(ctx context.Context, id model.IndexID) error
}

type IndexSubjectInfo struct {
	IndexID   model.IndexID
	SubjectID model.SubjectID
}

type IndexEditSubjectInfo struct {
	IndexID   model.IndexID
	SubjectID model.SubjectID
	Sort      uint32
	Comment   string
}

// 和目录内条目相关的数据库操作.
type indexSubjectRepo interface {
	CountSubjects(ctx context.Context, id model.IndexID, subjectType model.SubjectType) (int64, error)
	ListSubjects(
		ctx context.Context, id model.IndexID, subjectType model.SubjectType, limit, offset int,
	) ([]IndexSubject, error)
	AddIndexSubject(ctx context.Context, info IndexEditSubjectInfo) (*IndexSubject, error)
	UpdateIndexSubject(ctx context.Context, info IndexEditSubjectInfo) error
	DeleteIndexSubject(ctx context.Context, info IndexSubjectInfo) error
}

// 与用户和目录相关的操作.
type indexUserRepo interface {
	GetIndicesByUser(
		ctx context.Context, creatorID model.UserID, limit int, offset int,
	) ([]model.Index, error)
	GetCollectedIndicesByUser(
		ctx context.Context, creatorID model.UserID, limit int, offset int,
	) ([]model.IndexCollect, error)
}

type IndexRepo interface {
	indexRepo
	indexSubjectRepo
	indexUserRepo
}

type IndexSubject struct {
	Comment string
	AddedAt time.Time
	Subject model.Subject
}
