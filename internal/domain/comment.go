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

type CommentRepo interface {
	Get(ctx context.Context, commentType CommentType, id model.CommentID) (model.Comment, error)

	GetByIDs(
		ctx context.Context, commentType CommentType, ids ...model.CommentID,
	) (map[model.CommentID]model.Comment, error)

	GetByRelateIDs(
		ctx context.Context, commentType CommentType, ids ...model.CommentID,
	) (map[model.CommentID][]model.Comment, error)

	// Count top comments for a topic/index/character/person/episode.
	Count(ctx context.Context, commentType CommentType, id uint32) (int64, error)

	// List return paged top comment list of a topic/index/character/person/episode.
	List(
		ctx context.Context, commentType CommentType, id uint32, limit int, offset int,
	) ([]model.Comment, error)
}

type CommentType uint32

const (
	CommentTypeSubjectTopic CommentType = iota + 1
	CommentTypeGroupTopic
	CommentIndex
	CommentCharacter
	CommentPerson
	CommentEpisode
)
