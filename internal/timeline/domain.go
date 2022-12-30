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

package timeline

import (
	"context"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/internal/auth"
	"github.com/bangumi/server/internal/collection"
	"github.com/bangumi/server/internal/episode"
	"github.com/bangumi/server/internal/model"
)

type Repo interface {
	// WithQuery is used to replace repo's query to txn
	WithQuery(query *query.Query) Repo
	GetByID(ctx context.Context, id model.TimeLineID) (*model.TimeLine, error)
	ListByUID(ctx context.Context, uid model.UserID, limit int, since model.TimeLineID) ([]*model.TimeLine, error)
	Create(ctx context.Context, tl *model.TimeLine) error

	ChangeSubjectCollection(
		ctx context.Context,
		u auth.Auth,
		sbj model.Subject,
		collect model.SubjectCollection,
		comment string,
		rate uint8,
	) error

	ChangeEpisodeStatus(
		ctx context.Context,
		u auth.Auth,
		sbj model.Subject,
		episode episode.Episode,
		update collection.Update,
	) error
}
