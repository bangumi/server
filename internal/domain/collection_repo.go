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

	"github.com/bangumi/server/internal/dal/query"
	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
)

type CollectionRepo interface {
	// WithQuery is used to replace repo's query to txn
	WithQuery(query *query.Query) CollectionRepo
	CountSubjectCollections(
		ctx context.Context,
		userID model.UserID,
		subjectType model.SubjectType,
		collectionType model.SubjectCollection,
		showPrivate bool,
	) (int64, error)

	ListSubjectCollection(
		ctx context.Context,
		userID model.UserID,
		subjectType model.SubjectType,
		collectionType model.SubjectCollection,
		showPrivate bool,
		limit, offset int,
	) ([]model.UserSubjectCollection, error)

	GetSubjectCollection(
		ctx context.Context, userID model.UserID, subjectID model.SubjectID,
	) (model.UserSubjectCollection, error)

	GetSubjectEpisodesCollection(
		ctx context.Context, userID model.UserID, subjectID model.SubjectID,
	) (model.UserSubjectEpisodesCollection, error)

	UpdateSubjectCollection(
		ctx context.Context, userID model.UserID, subjectID model.SubjectID, data SubjectCollectionUpdate,
		at time.Time,
	) error

	UpdateEpisodeCollection(
		ctx context.Context,
		userID model.UserID,
		subjectID model.SubjectID,
		episodeIDs []model.EpisodeID,
		collection model.EpisodeCollection,
	) (model.UserSubjectEpisodesCollection, error)
}

type SubjectCollectionUpdate struct {
	Comment   null.String
	Tags      []string // nil 表示无数据，[]string{} 表示清空tag
	VolStatus null.Uint32
	EpStatus  null.Uint32
	Type      null.Null[model.SubjectCollection]
	Rate      null.Uint8
	Privacy   null.Null[model.CollectPrivacy]
}
