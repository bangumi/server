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

	"github.com/bangumi/server/internal/model"
)

type UserRepo interface {
	// GetByID find a user by uid.
	GetByID(ctx context.Context, userID model.UserID) (model.User, error)
	// GetByName find a user by username.
	GetByName(ctx context.Context, username string) (model.User, error)

	GetByIDs(ctx context.Context, ids ...model.UserID) (map[model.UserID]model.User, error)

	CountCollections(
		ctx context.Context,
		userID model.UserID,
		subjectType model.SubjectType,
		collectionType uint8,
		showPrivate bool,
	) (int64, error)

	ListCollections(
		ctx context.Context,
		userID model.UserID,
		subjectType model.SubjectType,
		collectionType uint8,
		showPrivate bool,
		limit, offset int,
	) ([]model.Collection, error)

	GetCollection(ctx context.Context, userID model.UserID, subjectID model.SubjectID) (model.Collection, error)
}
