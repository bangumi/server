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
	"github.com/bangumi/server/dal/query"
	"time"

	"github.com/bangumi/server/internal/model"
)

type Repo interface {
	IndexRepo
	CommentRepo
	SubjectRepo
	CollectRepo
}

//nolint:revive
type IndexRepo interface {
	Get(ctx context.Context, id model.IndexID) (model.Index, error)
	New(ctx context.Context, i *model.Index) error
	Update(ctx context.Context, id model.IndexID, title string, desc string) error
	Delete(ctx context.Context, id model.IndexID) error
}

type SubjectRepo interface {
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
}

type CollectRepo interface {
	// GetIndexCollect get and index colelct item if exists
	GetIndexCollect(ctx context.Context, id model.IndexID, uid model.UserID) (*IndexCollect, error)

	// AddIndexCollect add an index collect to given user
	AddIndexCollect(ctx context.Context, id model.IndexID, uid model.UserID) error

	// DeleteIndexCollect remove index collect from given user
	DeleteIndexCollect(ctx context.Context, id model.IndexID, uid model.UserID) error
}

type CommentRepo interface {
	// WithQuery is used to replace repo's query to txn
	WithQuery(query *query.Query) IndexRepo
	// GetIndexComments 查询所有 当前 Index下的 Comment
	GetIndexComments(ctx context.Context, id model.IndexID, offset int, limit int) ([]model.IndexComment, error)
	// GetIndexComment 相对应的获取指定 Comment
	GetIndexComment(ctx context.Context, id model.CommentID) (*model.IndexComment, error)
	AddIndexComment(ctx context.Context, newComment model.IndexComment) error
	// UpdateIndexComment 目录的评论需要更新吗？我不确定，但是先写再说
	UpdateIndexComment(ctx context.Context, indexID model.IndexID, comment string) error
	DeleteIndexComment(ctx context.Context, id model.IndexID) error
}

type Subject struct {
	Comment string
	AddedAt time.Time
	Subject model.Subject
}

//nolint:revive
type IndexCollect struct {
	ID          uint32
	IndexID     model.IndexID
	UserID      model.UserID
	CreatedTime time.Time
}
