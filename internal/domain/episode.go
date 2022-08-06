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

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
)

type EpisodeRepo interface {
	// WithQuery is used to replace repo's query to txn
	WithQuery(query *query.Query) EpisodeRepo

	Get(ctx context.Context, episodeID model.EpisodeID) (model.Episode, error)

	// Count all episode for a subject.
	Count(ctx context.Context, subjectID model.SubjectID, filter EpisodeFilter) (int64, error)

	// List return all episode.
	List(
		ctx context.Context,
		subjectID model.SubjectID,
		filter EpisodeFilter,
		limit int, offset int,
	) ([]model.Episode, error)
}

type EpisodeFilter struct {
	Type null.Null[model.EpType]
}
