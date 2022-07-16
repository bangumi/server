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

package topic

import (
	"time"

	"github.com/bangumi/server/internal/dal/dao"
	"github.com/bangumi/server/internal/model"
)

type mysqlTopic interface {
	GetCreateTime() time.Time
	GetUpdateTime() time.Time
	GetTitle() string
	GetID() uint32
	GetCreatorID() uint32
	GetState() uint8
	GetReplies() uint32
	GetObjectID() uint32
	GetStatus() uint8
}

var _ mysqlTopic = (*dao.GroupTopic)(nil)
var _ mysqlTopic = (*dao.SubjectTopic)(nil)

func wrapDao[T mysqlTopic](data []T, err error) ([]model.Topic, error) {
	if err != nil {
		return nil, err
	}

	var s = make([]model.Topic, len(data))
	for i, item := range data {
		s[i] = model.Topic{
			CreatedAt: item.GetCreateTime(),
			UpdatedAt: item.GetUpdateTime(),
			Title:     item.GetTitle(),
			ID:        model.TopicID(item.GetID()),
			CreatorID: model.UserID(item.GetCreatorID()),
			State:     model.TopicState(item.GetState()),
			Replies:   item.GetReplies(),
			ObjectID:  item.GetObjectID(),
			Status:    model.TopicStatus(item.GetStatus()),
		}
	}

	return s, err
}

type mysqlComment interface {
	CommentID() model.CommentID
	CreatorID() model.UserID
	IsSubComment() bool
	CreateAt() time.Time
	GetContent() string
	GetState() uint8
	RelatedTo() model.CommentID
	GetObjectID() uint32
}

func wrapCommentDao[T mysqlComment](data []T, err error) ([]mysqlComment, error) {
	if err != nil {
		return nil, err
	}

	var s = make([]mysqlComment, len(data))
	for i, item := range data {
		s[i] = item
	}

	return s, nil
}

var _ mysqlComment = (*dao.SubjectTopicComment)(nil)
var _ mysqlComment = (*dao.GroupTopicComment)(nil)
var _ mysqlComment = (*dao.EpisodeComment)(nil)
var _ mysqlComment = (*dao.CharacterComment)(nil)
var _ mysqlComment = (*dao.PersonComment)(nil)
