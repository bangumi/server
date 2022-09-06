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

	"go.uber.org/zap"

	"github.com/bangumi/server/internal/model"
)

type TopicRepo interface {
	Get(ctx context.Context, topicType TopicType, id model.TopicID) (model.Topic, error)

	// Count all topic for a subject/group.
	Count(ctx context.Context, topicType TopicType, id uint32, displays []model.TopicDisplay) (int64, error)

	// List return paged topic list of a subject/group.
	// userID should not be filtered
	List(
		ctx context.Context,
		topicType TopicType,
		id uint32,
		displays []model.TopicDisplay,
		limit int, offset int,
	) ([]model.Topic, error)

	GetTopicContent(ctx context.Context, topicType TopicType, id model.TopicID) (model.Comment, error)

	// CountReplies top comments for a topic/index/character/person/episode.
	// 一级回复
	// 小组和条目帖子会去除楼主发帖
	CountReplies(ctx context.Context, commentType CommentType, id model.TopicID) (int64, error)

	// ListReplies return paged top comment tree.
	//
	//  []model.Comment{
	//    一级回复
	// 		{
	//  	  一级回复对应的 **全部** 二级回复
	//	 		SubComments: []model.SubComments{}
	// 		}
	// }
	ListReplies(
		ctx context.Context, commentType CommentType, id model.TopicID, limit int, offset int,
	) ([]model.Comment, error)

	// CreateTopic()
	// CreateComment
	// CreateSubComment
}

type TopicType uint32

func (t TopicType) Zap() zap.Field {
	return zap.Uint32("topic_type", uint32(t))
}

const (
	TopicTypeUnknown TopicType = iota
	TopicTypeSubject
	TopicTypeGroup
)

type CommentType uint32

const (
	CommentTypeUnknown CommentType = iota
	CommentTypeSubjectTopic
	CommentTypeGroupTopic
	CommentIndex
	CommentCharacter
	CommentPerson
	CommentEpisode
)
