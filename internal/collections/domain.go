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

package collections

import (
	"context"
	"time"

	"github.com/bangumi/server/dal/query"
	"github.com/bangumi/server/internal/collections/domain/collection"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
)

type Repo interface { //nolint:interfacebloat
	// WithQuery is used to replace repo's query to txn
	WithQuery(query *query.Query) Repo
	CountSubjectCollections(
		ctx context.Context,
		userID model.UserID,
		subjectType model.SubjectType,
		collectionType collection.SubjectCollection,
		showPrivate bool,
	) (int64, error)

	ListSubjectCollection(
		ctx context.Context,
		userID model.UserID,
		subjectType model.SubjectType,
		collectionType collection.SubjectCollection,
		showPrivate bool,
		limit, offset int,
	) ([]collection.UserSubjectCollection, error)

	GetSubjectCollection(
		ctx context.Context, userID model.UserID, subjectID model.SubjectID,
	) (collection.UserSubjectCollection, error)

	GetSubjectEpisodesCollection(
		ctx context.Context, userID model.UserID, subjectID model.SubjectID,
	) (collection.UserSubjectEpisodesCollection, error)

	UpdateSubjectCollection(
		ctx context.Context, userID model.UserID, subject model.Subject,
		at time.Time, ip string,
		update func(ctx context.Context, s *collection.Subject) (*collection.Subject, error),
	) error

	UpdateOrCreateSubjectCollection(
		ctx context.Context, userID model.UserID, subject model.Subject,
		at time.Time, ip string,
		update func(ctx context.Context, s *collection.Subject) (*collection.Subject, error),
	) error

	UpdateEpisodeCollection(
		ctx context.Context,
		userID model.UserID, subjectID model.SubjectID,
		episodeIDs []model.EpisodeID, collection collection.EpisodeCollection,
		at time.Time,
	) (collection.UserSubjectEpisodesCollection, error)

	GetPersonCollection(
		ctx context.Context, userID model.UserID,
		cat collection.PersonCollectCategory, targetID model.PersonID,
	) (collection.UserPersonCollection, error)

	AddPersonCollection(
		ctx context.Context, userID model.UserID,
		cat collection.PersonCollectCategory, targetID model.PersonID,
	) error

	RemovePersonCollection(
		ctx context.Context, userID model.UserID,
		cat collection.PersonCollectCategory, targetID model.PersonID,
	) error

	CountPersonCollections(
		ctx context.Context,
		userID model.UserID,
		cat collection.PersonCollectCategory,
	) (int64, error)

	ListPersonCollection(
		ctx context.Context,
		userID model.UserID,
		cat collection.PersonCollectCategory,
		limit, offset int,
	) ([]collection.UserPersonCollection, error)
}

type Update struct {
	IP string

	Comment   null.String
	Tags      []string // nil 表示无数据，[]string{} 表示清空tag
	VolStatus null.Uint32
	EpStatus  null.Uint32
	Type      null.Null[collection.SubjectCollection]
	Rate      null.Uint8
	Privacy   null.Null[collection.CollectPrivacy]
}
