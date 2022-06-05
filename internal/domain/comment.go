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
	Get(ctx context.Context, commentType CommentType, id model.CommentIDType) (model.Comment, error)

	GetComments(
		ctx context.Context, commentType CommentType, id uint32, limit int, offset int,
	) (model.Comments, error)
}

type CommentType uint32

const (
	CommentTypeSubjectTopic CommentType = iota + 1
	CommentTypeGroupTopic
	CommentIndex
	CommentCharacter
	CommentEpisode
)

func ConvertModelCommentsToTree(comments []model.Comment, related uint32) []model.Comment {
	result := make([]model.Comment, 0)
	for _, v := range comments {
		if v.Related == related {
			if replies := ConvertModelCommentsToTree(comments, v.ID); len(replies) != 0 {
				v.Replies = replies
			}
			result = append(result, v)
		}
	}
	return result
}
