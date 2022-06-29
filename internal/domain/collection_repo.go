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

type CollectionRepo interface {
	CountSubjectCollections(
		ctx context.Context,
		userID model.UserID,
		subjectType model.SubjectType,
		collectionType model.SubjectCollectionType,
		showPrivate bool,
	) (int64, error)

	ListSubjectCollection(
		ctx context.Context,
		userID model.UserID,
		subjectType model.SubjectType,
		collectionType model.SubjectCollectionType,
		showPrivate bool,
		limit, offset int,
	) ([]model.SubjectCollection, error)

	GetSubjectCollection(
		ctx context.Context, userID model.UserID, subjectID model.SubjectID,
	) (model.SubjectCollection, error)

	UpdateEpisodeCollection(
		ctx context.Context,
		userID model.UserID,
		subjectID model.SubjectID,
		episodeID model.EpisodeID,
		collectionType model.EpisodeCollectionType, updatedAt time.Time,
	) error

	UpdateSubjectCollection(
		ctx context.Context, userID model.UserID, subjectID model.SubjectID, data model.SubjectCollectionUpdate,
	) error

	GetEpisodeCollection(
		ctx context.Context, userID model.UserID, subjectID model.SubjectID,
	) (model.EpisodeCollection, error)
}
