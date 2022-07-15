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

type TopicRepo interface {
	Get(ctx context.Context, topicType TopicType, id model.TopicID) (model.Topic, error)

	// Count all topic for a subject/group.
	Count(ctx context.Context, topicType TopicType, id uint32, statuses []model.TopicStatus) (int64, error)

	// List return paged topic list of a subject/group.
	// userID should not be filtered
	List(
		ctx context.Context,
		topicType TopicType,
		id uint32,
		statuses []model.TopicStatus,
		limit int, offset int,
	) ([]model.Topic, error)
}

type TopicType uint32

const (
	TopicTypeUnknown TopicType = iota
	TopicTypeSubject
	TopicTypeGroup
)
