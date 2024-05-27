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

//go:build !gen

package episode

import (
	"context"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
)

type Repo interface {
	EpRepo
	CommentRepo
}

type EpRepo interface {
	// WithQuery is used to replace repo's query to txn
	WithQuery(query *query.Query) Repo

	Get(ctx context.Context, episodeID model.EpisodeID) (Episode, error)

	// Count all episode for a subject.
	Count(ctx context.Context, subjectID model.SubjectID, filter Filter) (int64, error)

	// List return all episode.
	List(
		ctx context.Context,
		subjectID model.SubjectID,
		filter Filter,
		limit int, offset int,
	) ([]Episode, error)
}

type CommentRepo interface {
	// GetAllComment 获取当前EP下所有评论
	GetAllComment(ctx context.Context, episodeID model.EpisodeID, offset int, limit int) ([]model.EpisodeComment, error)
	AddNewComment(ctx context.Context, comment model.EpisodeComment) error
	DeleteComment(ctx context.Context, episodeID model.EpisodeID, userId model.UserID, commentID model.CommentID) error
}

type Filter struct {
	Type null.Null[Type]
}
