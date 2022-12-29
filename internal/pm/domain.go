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

package pm

import (
	"context"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/null"
)

type Repo interface {
	List(
		ctx context.Context,
		userID model.UserID,
		folder FolderType,
		offset int,
		limit int,
	) ([]PrivateMessageListItem, error)

	CountByFolder(
		ctx context.Context,
		userID model.UserID,
		folder FolderType,
	) (int64, error)

	ListRelated(
		ctx context.Context,
		userID model.UserID,
		id model.PrivateMessageID,
	) ([]PrivateMessage, error)

	CountTypes(ctx context.Context, userID model.UserID) (PrivateMessageTypeCounts, error)

	MarkRead(ctx context.Context, userID model.UserID, relatedID model.PrivateMessageID) error

	ListRecentContact(ctx context.Context, userID model.UserID) ([]model.UserID, error)

	Create(
		ctx context.Context,
		senderID model.UserID,
		receiverIDs []model.UserID,
		relatedIDFilter IDFilter,
		title string,
		content string,
	) ([]PrivateMessage, error)

	Delete(
		ctx context.Context,
		userID model.UserID,
		ids []model.PrivateMessageID,
	) error
}

type IDFilter struct {
	Type null.Null[model.PrivateMessageID]
}
